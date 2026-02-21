# go-ride - Core Backend Architecture & Guidelines

## 1. Project Vision & Identity

`go-ride` is a highly scalable, real-time ride-hailing core system (Modular Monolith) built in Go. This is NOT a standard CRUD application. It is a domain-driven, event-driven, high-concurrency engine.

As an AI assistant contributing to this project, you MUST act as a Senior/Lead Go Engineer. You must prioritize system reliability, idempotency, thread safety, and strict separation of concerns over quick-and-dirty solutions.

## 2. Tech Stack (Strictly Enforced)

- **Language:** Go (1.21+)
- **HTTP Framework:** Fiber (v2) - Used strictly for external API delivery.
- **Internal RPC:** gRPC & Protobuf - Used for synchronous module-to-module communication.
- **Database:** PostgreSQL - **NO ORM ALLOWED**. Use `pgx` (or standard `database/sql`). Raw SQL queries, explicit transaction management, and connection pooling are mandatory.
- **Caching & Geo-Spatial:** Redis - Used for driver location tracking (`GEOADD`, `GEORADIUS`), rate limiting, and short-lived caching.
- **Message Broker:** Apache Kafka - Used for asynchronous, event-driven communication between modules (e.g., Trip completed -> Payment triggered).

## 3. Directory Structure & File Responsibilities

You must strictly follow this Modular Monolith directory structure. Never place business logic in the `cmd` or `presentation` layers.

```text
go-ride/
├── cmd/                               # Entrypoints for the application
│   ├── api/main.go                    # Initializes Fiber, gRPC servers, and wires dependencies
│   └── worker/main.go                 # Initializes background jobs and Kafka consumers
│
├── internal/                          # Private application code
│   ├── platform/                      # Shared infrastructure components (Cross-cutting concerns)
│   │   ├── config/                    # Environment variables mapping
│   │   ├── database/                  # PostgreSQL connection pool (pgx)
│   │   ├── cache/                     # Redis client setup
│   │   ├── logger/                    # Structured logging (e.g., zap)
│   │   ├── eventbus/                  # Kafka producer/consumer wrappers
│   │   └── middleware/                # Fiber middlewares (Auth, Rate Limit, Error Handler)
│   │
│   └── modules/                       # BUSINESS MODULES (The Core)
│       ├── identity/                  # User/Driver profiles, Auth (JWT)
│       ├── trip/                      # Trip state machine (Requested, Accepted, Ongoing, Completed)
│       ├── matching/                  # Finding nearest drivers via Redis Geo
│       ├── location/                  # Processing high-frequency GPS pings
│       └── payment/                   # Fares, holds, and captures
│
│       # --- INSIDE EACH MODULE (Clean Architecture) ---
│       # e.g., internal/modules/trip/
│       ├── domain/                    # 1. CORE: Pure Go. NO external deps (No Fiber, No Postgres)
│       │   ├── entity/                # Structs representing business objects (e.g., Trip, Status)
│       │   ├── repository/            # Interfaces for data access (e.g., TripRepository)
│       │   └── error.go               # Domain-specific errors (e.g., ErrTripAlreadyAccepted)
│       │
│       ├── application/               # 2. USE CASES: Orchestrates domain logic
│       │   ├── usecase/               # Business scenarios (e.g., CreateTrip, CompleteTrip)
│       │   └── dto/                   # Data Transfer Objects (Requests/Responses)
│       │
│       ├── presentation/              # 3. DELIVERY: Traffic routing. ONLY layer that knows Fiber/gRPC
│       │   ├── http/                  # Fiber handlers (Parses JSON -> DTO -> calls UseCase)
│       │   ├── grpc/                  # gRPC server implementations
│       │   └── consumer/              # Kafka event listeners
│       │
│       └── infrastructure/            # 4. ADAPTERS: Implements domain interfaces
│           ├── persistence/           # Postgres Raw SQL queries (Implements repository interface)
│           └── messaging/             # Publishes Kafka events
│
├── pkg/                               # Public/Shared utilities (No business logic)
│   └── errors/                        # Standardized error mapping
│
├── deploy/                            # Docker, K8s, Terraform files
└── scripts/                           # DB Migrations, CI/CD scripts
4. Actor Roles, Routing & Security
Rider: /api/v1/rider/... - High read throughput, bursts of writes.

Driver: /api/v1/driver/... - Constant high-frequency writes (GPS telemetry).

Admin: /api/v1/admin/... - Heavy analytics, overrides. Strict RBAC required.

5. System Design Principles & Edge Cases
When generating code, you MUST account for:

Idempotency: A driver might click "Accept Trip" twice. Handle retries gracefully without double-assigning.

Concurrency / Race Conditions: Two drivers accepting the same ride. Use DB optimistic locking (version col) or Redis distributed locks.

Eventual Consistency: Trip completion and Payment are async. Do not block the driver response waiting for the bank.

Error Handling: Never return raw Go errors or SQL panics to the client. Use central pkg/errors mapping.

6. Communication Matrix
External to Backend: REST via Fiber.

Module to Module (Read): gRPC.

Module to Module (Write/State Change): Kafka.
```

## 7. Strict Coding Style & Commenting Rules

When generating or modifying code, you MUST adhere to these personal conventions:

- **No AI-style Comments:** Do NOT write robotic, overly obvious comments (e.g., avoid `// Checks if user is active` for a simple `if` statement).
- **Purposeful Commenting:** Only write comments for complex business logic, edge cases, workarounds, or to explain the "why" behind a decision.
- **Lowercase Constraint:** CRITICAL: ALL comments MUST start with a lowercase letter. This is a strict personal convention (e.g., `// applying optimistic lock to prevent race condition`).
- **Human Touch:** Write code that looks like an experienced, pragmatic senior human engineer wrote it.
