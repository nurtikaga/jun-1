# Backend Developer Assessment

A clean-architecture Go service implementing the Product domain entity.

## Project structure

```
.
├── cmd/
│   └── server/          # Entry point: HTTP server wiring, graceful shutdown
│       ├── main.go
│       └── repo.go      # In-memory repository (replace with DB in production)
├── internal/
│   ├── domain/
│   │   ├── product/     # Core domain entity — no external dependencies
│   │   │   ├── product.go
│   │   │   └── product_test.go
│   │   └── errors/      # Domain-level sentinel errors and typed DomainError
│   │       └── errors.go
│   ├── app/             # Application-layer use-cases (context lives here)
│   │   ├── product_service.go
│   │   └── product_service_test.go
│   └── handler/         # HTTP delivery layer — translates HTTP ↔ app layer
│       ├── product_handler.go
│       └── product_handler_test.go
├── pkg/
│   ├── health/          # /healthz endpoint
│   └── logger/          # Structured JSON logger (log/slog)
├── .github/
│   └── workflows/
│       └── ci.yml       # Lint → Test → Build → Docker CI pipeline
├── .golangci.yml        # golangci-lint configuration
├── Dockerfile           # Multi-stage build (scratch runtime)
├── docker-compose.yml
├── Makefile
├── REVIEW.md            # Bug analysis of the provided buggy code
└── ANSWERS.md           # Answers to assessment questions
```

## Quick start

```bash
# Run tests
make test

# Run with race detector + coverage
make test-coverage

# Build binary
make build

# Run server (default :8080)
make run

# Lint
make lint

# All checks + build
make all
```

## Docker

```bash
make docker-up    # build image and start container
make docker-logs  # follow logs
make docker-down  # stop
```

## API

| Method | Path | Description |
|--------|------|-------------|
| `GET`  | `/healthz` | Health check |
| `POST` | `/products` | Create product |
| `GET`  | `/products/{id}` | Get product |
| `PATCH`| `/products/{id}/price` | Update price |
| `PATCH`| `/products/{id}/stock/decrease` | Decrease stock |

### Create product
```bash
curl -X POST http://localhost:8080/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"Widget","price_in_cents":999,"stock":100}'
```

### Decrease stock
```bash
curl -X PATCH http://localhost:8080/products/{id}/stock/decrease \
  -H 'Content-Type: application/json' \
  -d '{"quantity":10}'
```

## Key design decisions

- **Domain is pure.** No `context.Context`, no logging, no infrastructure imports inside `internal/domain/`.
- **int64 for money.** Price stored in cents — never `float64`.
- **Typed errors.** `errors.Is()` / `errors.As()` work correctly at every layer via wrapped sentinels.
- **ChangeTracker is explicit.** Fields are marked dirty only when a mutation succeeds.
- **Structured logging.** `log/slog` with JSON handler — only at the handler boundary.
- **Graceful shutdown.** `SIGINT`/`SIGTERM` drain in-flight requests before exit.
