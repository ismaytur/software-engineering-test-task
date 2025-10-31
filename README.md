# Simple CRUD Interface

REST API that manages users.

## Prerequisites

- [Docker](https://www.docker.com/get-started/)
- [Go](https://go.dev/dl/)
- Make

## Environment configuration

Create a `.env` file in the repository as the sample below.

```env
# Postgres
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=postgres
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_SSL_MODE=disable

# Logging (optional overrides)
LOG_OUTPUT=stdout             # stdout | file | both
# LOG_FILE=/var/log/app.json  # required when LOG_OUTPUT includes file
LOG_LEVEL=info                # debug | info | warn | error
API_KEY_CACHE_TTL=5m          # duration for in-memory API key cache
```

## Makefile quick reference

Run `make help` any time to list all targets and descriptions. Common workflows are summarized below.

### One-time setup

```bash
make install      # install modules and needed tools 
```

### Local development

```bash
make validate     # generate mocks + go vet + staticcheck + gosec + unit tests
make run          # run the HTTP server locally (uses .env DSN)
make swagger      # regenerate docs in ./docs
```

### Testing

```bash
make test                 # unit tests (TEST_ARGS overrides are supported)
make test-integration     # integration tests (requires Docker)

make coverage             # unit coverage -> coverage/unit.out
make coverage-html        # open HTML report for unit coverage
make coverage-integration # integration coverage -> coverage/integration.out
```

### Database migrations

```bash
make migrate-up           # apply all migrations
make migrate-down         # roll back the last migration
make migrate-status       # show applied migration status
make create-migration     # interactive prompt, generates empty SQL file
```

All migrations run through `goose` and reuse the DSN defined in `.env`.

### Docker workflow

```bash
make app                  # cross-compile for container + docker-compose up app
make db                   # start postgres container only
make up                   # build and run full compose stack
make down                 # stop and remove compose services
```

The `app` target cross-compiles the binary for Linux using `APP_GOOS`/`APP_GOARCH` (defaults to the host architecture). Override if you need a different image target, e.g. `APP_GOARCH=amd64 make app`.

### Housekeeping

```bash
make clean                # remove bin/, coverage/, generated mocks
```

## Project structure

```
cmd/                 # application entrypoint
internal/app         # application bootstrap (DI, router wiring)
internal/controller  # HTTP controllers
internal/handler     # gin route definitions
internal/repository  # persistence layer
internal/service     # business logic (unit/integration tests live here)
migrations/          # goose SQL migrations
docs/                # Swagger artifacts
docker/entrypoint.sh # runtime DSN assembly for container builds
```

## Structured logging

- The app uses `log/slog` with JSON output by default.
- Configure via environment variables:
  - `LOG_OUTPUT`: `stdout` (default), `file`, or `both`.
  - `LOG_FILE`: absolute path used when `LOG_OUTPUT` is `file` or `both`; directories are created with 0700 permissions.
  - `LOG_LEVEL`: `debug`, `info` (default), `warn`, or `error`.
- HTTP requests automatically produce structured logs with timing, status, method, route, and request IDs.
- Services and repositories emit contextual logs 

## API key authentication

- All HTTP calls must include `X-API-Key`. Missing keys return `401 Unauthorized`; invalid keys return `403 Forbidden`.
- Keys are stored (sha256sum hashed) in `api_keys`. Insert new keys manually.
- Lookups are cached in-memory for `API_KEY_CACHE_TTL` to reduce database traffic.

## API endpoints

- `GET /api/v1/users/` – list all users
- `GET /api/v1/users/username/{username}` – fetch by username
- `GET /api/v1/users/id/{id}` – fetch by numeric ID
- `GET /api/v1/users/uuid/{uuid}` – fetch by UUID
- `POST /api/v1/users/` – create user
- `PATCH /api/v1/users/uuid/{uuid}` – update by UUID
- `PATCH /api/v1/users/id/{id}` – update by ID
- `DELETE /api/v1/users/uuid/{uuid}` – delete by UUID
- `DELETE /api/v1/users/id/{id}` – delete by ID

Base URL defaults to `http://localhost:8080` when running via `make run` or `make app`.

### Swagger / OpenAPI

- JSON: `docs/swagger.json`
- YAML: `docs/swagger.yaml`
- Generated with `make swagger` (uses `swag init -g ./cmd/main.go`)
