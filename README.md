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

## How to run 

1. Start database

```
## Via Makefile
make db

## Via Docker
docker-compose up -d db
```

2. Run migrations

```
## Via Makefile
make migrate-up

## Via Goose
DB_DRIVER=postgres
DB_STRING="host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
goose -dir ./migrations $(DB_DRIVER) $(DB_STRING) up
```

3. Run application

```
make run
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
