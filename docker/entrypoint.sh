#!/bin/sh
set -eu

export POSTGRES_DSN=${POSTGRES_DSN:-"host=${POSTGRES_HOST:-db} port=${POSTGRES_PORT:-5432} user=${POSTGRES_USER:-postgres} password=${POSTGRES_PASSWORD:-postgres} dbname=${POSTGRES_DB:-postgres} sslmode=${POSTGRES_SSL_MODE:-disable}"}

exec /app/server "$@"
