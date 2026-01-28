#!/bin/sh
set -e

DB_DSN="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo "Running migrations..."
goose -dir migrations postgres "${DB_DSN}" up

echo "Starting application..."
exec "$@"
