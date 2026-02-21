# Setup Guide

This guide walks you through running the entire **go-ride** stack locally using Docker Compose. No Go toolchain installation is required — everything builds and runs inside containers.

---

## Prerequisites

| Tool                                                              | Version | Purpose                        |
| ----------------------------------------------------------------- | ------- | ------------------------------ |
| [Docker Desktop](https://www.docker.com/products/docker-desktop/) | 20.10+  | Container runtime & Compose V2 |
| [Git](https://git-scm.com/)                                       | 2.x     | Clone the repository           |

> **Note:** Docker Desktop ships with `docker compose` (V2). If you're on Linux without Desktop, install the [Compose plugin](https://docs.docker.com/compose/install/linux/) separately.

---

## 1. Clone the Repository

```bash
git clone https://github.com/aliwert/go-ride.git
cd go-ride
```

---

## 2. Start the Infrastructure

The project provides one-click scripts that build the API image and spin up all dependencies.

### Windows

```powershell
.\scripts\start.bat
```

### macOS / Linux

```bash
chmod +x scripts/start.sh
./scripts/start.sh
```

Both scripts execute `docker-compose up --build -d` under the hood.

### Manual Start (any OS)

```bash
docker compose up --build -d
```

After a successful launch you should see five healthy containers:

| Container          | Port   | Description                  |
| ------------------ | ------ | ---------------------------- |
| `goride-api`       | `3000` | Go Fiber HTTP server         |
| `goride-postgres`  | `5432` | PostgreSQL 15                |
| `goride-redis`     | `6379` | Redis 7                      |
| `goride-kafka`     | `9092` | Confluent Kafka              |
| `goride-zookeeper` | `2181` | ZooKeeper (Kafka dependency) |

Verify with:

```bash
docker compose ps
```

---

## 3. Run Database Migrations

The API container depends on Postgres being healthy, but tables are not created automatically. Run the migration scripts against the running Postgres container:

```bash
docker exec -i goride-postgres psql -U postgres -d goride < scripts/migrations/000001_create_users_table.up.sql
```

```bash
docker exec -i goride-postgres psql -U postgres -d goride < scripts/migrations/000002_create_trips_table.up.sql
```

> **Rollback:** Replace `.up.sql` with `.down.sql` to reverse a migration.

---

## 4. Verify the API

Once migrations are applied, the API is ready:

```bash
curl http://localhost:3000/api/v1/identity/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","first_name":"John","last_name":"Doe","role":"RIDER"}'
```

You should receive a `201 Created` response with user details and JWT tokens.

---

## 5. Tail Logs

```bash
# all services
docker compose logs -f

# api only
docker compose logs -f api
```

---

## 6. Stop Everything

```bash
docker compose down
```

To also wipe the database volume:

```bash
docker compose down -v
```

---

## Environment Variables

The Docker Compose file injects all required env vars into the API container, overriding any `.env` defaults via Viper's `AutomaticEnv()`. You can tune values in the `environment` block of the `api` service inside `docker-compose.yml`:

| Variable       | Default (Compose)                                                   | Description                     |
| -------------- | ------------------------------------------------------------------- | ------------------------------- |
| `APP_ENV`      | `development`                                                       | Runtime environment             |
| `PORT`         | `3000`                                                              | HTTP listen port                |
| `DATABASE_URL` | `postgres://postgres:postgres@postgres:5432/goride?sslmode=disable` | PostgreSQL DSN                  |
| `REDIS_URL`    | `redis://redis:6379/0`                                              | Redis connection string         |
| `KAFKA_URL`    | `kafka:29092`                                                       | Kafka broker address            |
| `JWT_SECRET`   | `your-beautiful-secret-key`                                         | HMAC signing key for JWT tokens |

> **Production:** Never commit real secrets. Use Docker secrets, Vault, or a `.env` file excluded from version control.
