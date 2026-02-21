# ── stage 1: build ────────────────────────────────────────────────────────────
# using alpine-based go image to keep the builder layer small
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# dependency cache — these layers only invalidate when go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# copy the full source after deps are cached
COPY . .

# static binary with no cgo — safe for scratch/alpine runtime
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /src/main ./cmd/api


# ── stage 2: minimal runtime ─────────────────────────────────────────────────
FROM alpine:latest

# tls certs + timezone data for production sanity
RUN apk add --no-cache ca-certificates tzdata

# non-root user — never run a go binary as root in containers
RUN addgroup -S goride && adduser -S goride -G goride

WORKDIR /app

COPY --from=builder /src/main .
# .env acts as a fallback; docker-compose env vars take precedence via viper.AutomaticEnv()
COPY --from=builder /src/.env .

RUN chown -R goride:goride /app
USER goride

EXPOSE 3000

ENTRYPOINT ["./main"]
