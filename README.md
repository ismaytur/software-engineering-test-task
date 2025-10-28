# Simple CRUD Interface

This is a simple CRUD application

## Prerequisites

1. Tools
- [Docker](https://www.docker.com/get-started/)
- [Go](https://go.dev/dl/)
- [Goose](https://github.com/pressly/goose)

2. Environment variables

Create `.env` file in main directory with the following template:

```
## Postgres
POSTGRES_USER=postgres user name
POSTGRES_PASSWORD=postgres password
POSTGRES_DB=postgres database name
POSTGRES_HOST=host name
POSTGRES_PORT=port (5432 is default)
POSTGRES_SSL_MODE=see postgres for available modes (e.g. disable|require)
```

## Development workflow

### Install dependencies

```bash
go mod download
```

### Start Postgres

```bash
# via Makefile
make db

# or directly with Docker Compose
docker-compose up -d db
```

### Apply migrations

```bash
# using the Makefile shortcut
make migrate-up

# or directly with goose
DB_DRIVER=postgres \
DB_STRING="host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable" \
goose -dir ./migrations $(DB_DRIVER) $(DB_STRING) up
```

### Run the service

```bash
make run
```

### Run tests

Mocks are generated automatically through `go generate` (the Makefile does this for you):

```bash
make test               # unit suite
make test-integration   # integration suite (requires Docker)
```

### Coverage reports

```bash
make coverage                  # unit coverage summary (coverage-unit.out)
make coverage-html             # opens HTML report for unit coverage
make coverage-integration      # integration coverage summary (coverage-integration.out)
make coverage-integration-html # opens HTML report for integration coverage
```

## API Endpoints

- `GET /api/v1/users/` - List all users.
- `GET /api/v1/users/username/{username}` - Retrieve a user by username.
- `GET /api/v1/users/id/{id}` - Retrieve a user by numeric ID.
- `GET /api/v1/users/uuid/{uuid}` - Retrieve a user by UUID.
- `POST /api/v1/users/` - Create a new user.
- `PATCH /api/v1/users/uuid/{uuid}` - Update an existing user by UUID.
- `DELETE /api/v1/users/uuid/{uuid}` - Remove a user by UUID.
- `PATCH /api/v1/users/id/{id}` - Update an existing user by numeric ID.
- `DELETE /api/v1/users/id/{id}` - Remove a user by numeric ID.

OpenAPI documentation (Swagger): [`./docs/swagger.json`](docs/swagger.json)
