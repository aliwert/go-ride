<div align="center">

# go-ride

**A high-performance, real-time ride-hailing backend engine built in Go.**

Modular Monolith · Clean Architecture · Domain-Driven Design

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v2-00ACD7?style=for-the-badge&logo=go&logoColor=white)](https://gofiber.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![Apache Kafka](https://img.shields.io/badge/Kafka-Event_Bus-231F20?style=for-the-badge&logo=apachekafka&logoColor=white)](https://kafka.apache.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)

</div>

---

## OverviewAll done. Build passes, all five files created:

File Content
README.md Visual, badge-rich project overview with feature highlights, tech stack table, quick start, architecture diagram, and doc links
setup.md Full Docker Compose setup guide — prerequisites, one-click scripts, exact migration commands, env var reference
architecture.md Modular Monolith + Clean Architecture explanation, Ports & Adapters deep-dive (matching→location, location→tracking), initialization order, concurrency strategy
api.md Complete API reference — every endpoint with request/response schemas, auth requirements, error codes, and the WebSocket tracking protocol

**go-ride** is a production-grade ride-hailing core system that solves the hard problems: real-time GPS tracking, concurrent trip assignment, geospatial driver matching, and event-driven workflows — all inside a single deployable binary with clean module boundaries.

This is not a CRUD tutorial. It's an engineering-first backend that treats **concurrency**, **idempotency**, and **separation of concerns** as first-class citizens.

---

## Key Features

### Modular Monolith with Clean Architecture

Five self-contained business modules (`identity`, `location`, `trip`, `matching`, `tracking`) — each following strict Domain → Application → Infrastructure → Presentation layering. Modules communicate through **Ports & Adapters** interfaces, never by reaching into each other's internals.

### Real-Time Driver Tracking via WebSocket

Riders connect to `ws://host/ws/trip/:trip_id` and receive live `LOCATION_UPDATE` frames as the driver moves. The platform WebSocket Hub manages connections with `sync.RWMutex` for thread safety and automatic stale connection eviction.

### Redis Geospatial Driver Matching

Driver positions are stored using Redis `GEOADD` with TTL-based heartbeats. The matching module finds all drivers within a configurable radius via `GEOSEARCH` — no full table scans, no polling.

### Optimistic Locking for Trip State Machine

When two drivers race to accept the same ride, the database settles it: `UPDATE trips SET driver_id = $1 WHERE id = $2 AND status = 'REQUESTED'`. Exactly one driver wins. The loser gets a `409 TRIP_ALREADY_ACCEPTED` — no distributed locks required.

### Event-Driven Architecture (Kafka-Ready)

Kafka and ZooKeeper are provisioned in the Docker Compose stack. The infrastructure is wired for asynchronous module-to-module communication (e.g., trip completion → payment trigger) as the system grows.

### Centralized Error Handling

Every error flows through a global `ErrorHandler` that returns a consistent JSON envelope with machine-readable codes (`INVALID_CREDENTIALS`, `TRIP_ALREADY_ACCEPTED`, `NO_DRIVERS_AVAILABLE`) — no raw stack traces leak to clients.

---

## Tech Stack

| Layer       | Technology                | Purpose                                             |
| ----------- | ------------------------- | --------------------------------------------------- |
| Language    | Go 1.24                   | Performance, concurrency, type safety               |
| HTTP        | Fiber v2                  | High-throughput HTTP framework                      |
| WebSocket   | gofiber/contrib/websocket | Real-time bidirectional communication               |
| Database    | PostgreSQL 15             | Transactional data, optimistic locking              |
| Cache / Geo | Redis 7                   | Geospatial indexing, driver heartbeat TTL           |
| Messaging   | Apache Kafka              | Async event bus (provisioned, ready to integrate)   |
| Auth        | JWT (HMAC-SHA256)         | Stateless authentication with access/refresh tokens |
| Config      | Viper                     | `.env` files with OS env override                   |
| Container   | Docker + Compose          | One-command full stack deployment                   |

---

## Project Structure

```
go-ride/
├── cmd/api/                    → application entrypoint
├── internal/
│   ├── platform/               → shared infrastructure
│   │   ├── config/             → environment config (Viper)
│   │   ├── database/           → PostgreSQL connection pool (pgx)
│   │   ├── cache/              → Redis client
│   │   ├── middleware/         → JWT auth middleware
│   │   ├── apierror/          → global error handler
│   │   ├── websocket/         → WebSocket Hub
│   │   └── server/            → Fiber server lifecycle
│   └── modules/
│       ├── identity/           → registration, login, JWT issuance
│       ├── location/           → driver GPS tracking (Redis Geo)
│       ├── trip/               → trip state machine
│       ├── matching/           → geospatial driver matching
│       └── tracking/           → real-time WS location push
├── scripts/migrations/         → raw SQL migrations
├── docs/                       → detailed documentation
├── Dockerfile                  → multi-stage production build
└── docker-compose.yml          → full local infrastructure
```

---

## Quick Start

### Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (20.10+)
- [Git](https://git-scm.com/)

### Run

```bash
git clone https://github.com/aliwert/go-ride.git
cd go-ride
```

**Windows:**

```powershell
.\scripts\start.bat
```

**macOS / Linux:**

```bash
chmod +x scripts/start.sh && ./scripts/start.sh
```

Then apply database migrations:

```bash
docker exec -i goride-postgres psql -U postgres -d goride < scripts/migrations/000001_create_users_table.up.sql
docker exec -i goride-postgres psql -U postgres -d goride < scripts/migrations/000002_create_trips_table.up.sql
```

The API is live at **http://localhost:3000**.

> For the full walkthrough, see the [Setup Guide](docs/setup.md).

---

## Documentation

| Document                             | Description                                              |
| ------------------------------------ | -------------------------------------------------------- |
| [Setup Guide](docs/setup.md)         | Docker Compose setup, migrations, environment variables  |
| [Architecture](docs/architecture.md) | Modular Monolith, Clean Architecture, Ports & Adapters   |
| [API Reference](docs/api.md)         | Full endpoint documentation with schemas and error codes |

---

## Architecture Highlights

```
                          ┌──────────────┐
                          │   Fiber v2   │
                          │  HTTP + WS   │
                          └──────┬───────┘
                                 │
              ┌──────────────────┼──────────────────┐
              ▼                  ▼                   ▼
       ┌─────────┐       ┌─────────────┐     ┌──────────┐
       │Identity │       │  Location   │     │   Trip   │
       │ Module  │       │   Module    │     │  Module  │
       └─────────┘       └──────┬──────┘     └──────────┘
                                │
                   ┌────────────┼────────────┐
                   ▼                         ▼
            ┌─────────────┐          ┌──────────────┐
            │  Matching   │          │   Tracking   │
            │   Module    │          │    Module    │
            │             │          │  (WebSocket) │
            └─────────────┘          └──────────────┘
                   │                         │
            LocationPort              BroadcasterPort
            (interface)               (interface)
```

Cross-module communication flows through **port interfaces** — the matching module queries drivers via `LocationPort`, and the location module pushes updates via `BroadcasterPort`. Neither module knows about the other's storage layer.

---

## License

This project is open source and available under the [MIT License](LICENSE).

---

_You can use these services in your frontend projects, happy coding!_
