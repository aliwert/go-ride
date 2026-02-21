# Architecture Overview

This document explains the structural decisions behind **go-ride** — a ride-hailing backend built as a **Modular Monolith** in Go.

---

## High-Level Structure

```
go-ride/
├── cmd/api/              → thin entrypoint (config, wiring, server start)
├── internal/
│   ├── platform/         → shared infra (database, cache, middleware, websocket hub)
│   └── modules/          → business modules (identity, location, trip, matching, tracking)
├── scripts/migrations/   → raw SQL migration files
├── docs/                 → project documentation
├── Dockerfile            → multi-stage container build
└── docker-compose.yml    → full local stack
```

The `internal/` directory enforces Go's visibility rules — nothing inside it is importable by external packages.

---

## Modular Monolith

Each business capability lives in its own module under `internal/modules/`. Modules are self-contained and follow identical internal layering. They communicate through well-defined interfaces, never by reaching into each other's database tables or internal types.

### Modules

| Module       | Responsibility                                                    |
| ------------ | ----------------------------------------------------------------- |
| **identity** | User registration, login, JWT token issuance                      |
| **location** | Driver GPS tracking via Redis Geospatial (`GEOADD` / `GEOSEARCH`) |
| **trip**     | Trip lifecycle state machine (`REQUESTED → ACCEPTED → COMPLETED`) |
| **matching** | Finding nearby drivers and fanning out ride notifications         |
| **tracking** | Real-time location push to riders via WebSocket                   |

Each module exposes a single `InitModule()` function that wires its internals and registers routes. The server orchestrates module initialization order and passes shared dependencies (DB pool, Redis client, auth middleware).

---

## Clean Architecture (per Module)

Every module follows a strict four-layer architecture. Dependencies point **inward** — outer layers depend on inner layers, never the reverse.

```
┌─────────────────────────────────────┐
│         presentation/               │  ← HTTP handlers, WS handlers, Kafka consumers
│  (Fiber routes, request parsing)    │
├─────────────────────────────────────┤
│         infrastructure/             │  ← Adapters (Postgres repos, Redis repos, adapters)
│  (implements domain interfaces)     │
├─────────────────────────────────────┤
│         application/                │  ← Use cases, DTOs, port interfaces
│  (orchestrates domain logic)        │
├─────────────────────────────────────┤
│         domain/                     │  ← Entities, repository interfaces, domain errors
│  (pure Go, zero external deps)      │
└─────────────────────────────────────┘
```

### Layer Rules

| Layer              | Allowed Dependencies                         | Forbidden                                  |
| ------------------ | -------------------------------------------- | ------------------------------------------ |
| **Domain**         | Standard library only                        | Fiber, pgx, Redis, any framework           |
| **Application**    | Domain layer, port interfaces                | Infrastructure, Presentation               |
| **Infrastructure** | Domain + Application (implements interfaces) | Presentation                               |
| **Presentation**   | Application layer (calls use cases)          | Domain internals, Infrastructure internals |

---

## Ports & Adapters (Hexagonal Architecture)

Cross-module communication uses the **Ports & Adapters** pattern to prevent direct coupling between modules.

### Problem

The `matching` module needs to query nearby drivers (owned by `location`). The `tracking` module needs to push GPS updates to WebSocket clients. If these modules imported each other's internal packages, we'd have a tightly coupled monolith.

### Solution

Each consuming module defines a **port** (interface) describing _what_ it needs. An **adapter** (implementation) bridges the gap at the wiring layer.

```
matching module                          location module
┌──────────────────┐                    ┌──────────────────┐
│  LocationPort    │───adapter────────▶│  *LocationUseCase │
│  (interface)     │                    │  .FindNearbyDrivers()
└──────────────────┘                    └──────────────────┘

location module                          tracking module
┌──────────────────┐                    ┌──────────────────┐
│  UpdateLocation  │───broadcasts──────▶│  BroadcasterPort │
│  (use case)      │                    │  (interface)      │
└──────────────────┘                    └──────────────────┘
```

### Matching → Location

| Component         | Package                            | Description                                                    |
| ----------------- | ---------------------------------- | -------------------------------------------------------------- |
| `LocationPort`    | `matching/application/port/`       | Interface: `FindNearbyDrivers(ctx, lat, lon, radius)`          |
| `LocationAdapter` | `matching/infrastructure/adapter/` | Wraps `*location.LocationUseCase` and satisfies `LocationPort` |

The `matching` module never imports `location/infrastructure` or touches Redis directly. It only knows about the port it defined.

### Location → Tracking (WebSocket Broadcast)

| Component         | Package                            | Description                                                   |
| ----------------- | ---------------------------------- | ------------------------------------------------------------- |
| `BroadcasterPort` | `tracking/application/port/`       | Interface: `BroadcastLocation(ctx, tripID, lat, lon)`         |
| `HubBroadcaster`  | `tracking/infrastructure/adapter/` | Serializes payload to JSON, pushes via platform WebSocket Hub |

When a driver sends a GPS ping with a `trip_id`, the location use case persists to Redis and then fires a non-blocking broadcast through the `BroadcasterPort`. The location module has zero knowledge of WebSockets.

### Platform WebSocket Hub

The Hub (`internal/platform/websocket/hub.go`) is a transport-level component, not a business module. It manages a thread-safe map of `tripID → []*websocket.Conn` using `sync.RWMutex`, with stale connection eviction on write failure.

---

## Initialization Order

Module wiring in `server.MountHandlers()` respects dependency direction:

```
1. tracking.InitModule(app, hub)         → returns BroadcasterPort
2. identity.InitModule(v1, dbPool, jwt)  → public routes
3. location.InitModule(v1, redis, auth, broadcaster) → returns *LocationUseCase
4. trip.InitModule(v1, dbPool, auth)
5. matching.InitModule(v1, locUC, auth)  → receives LocationUseCase via adapter
```

Tracking is initialized first because location depends on its `BroadcasterPort`. Location is initialized before matching because matching depends on its `*LocationUseCase`.

---

## Concurrency & Safety

| Concern                             | Mechanism                                                                           |
| ----------------------------------- | ----------------------------------------------------------------------------------- |
| Two drivers accepting the same trip | Optimistic locking via `UPDATE ... WHERE status = 'REQUESTED'` — only one succeeds  |
| WebSocket hub concurrent access     | `sync.RWMutex` guards the connection map                                            |
| Redis geospatial stale entries      | TTL heartbeat keys filter out inactive drivers                                      |
| Broadcast latency                   | Location → WS push runs in a goroutine (non-blocking to the driver's HTTP response) |

---

## Error Handling

All modules use a centralized `apierror` package. Handlers return `*apierror.AppError` with structured codes. The Fiber `GlobalErrorHandler` serializes them into a consistent JSON envelope:

```json
{
  "success": false,
  "error": {
    "code": "TRIP_ALREADY_ACCEPTED",
    "message": "this trip has already been accepted by another driver"
  }
}
```

See [docs/api.md](api.md) for the full error code reference.
