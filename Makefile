include .env

GOOSE_DSN := "postgres://${DB_USER}:${DB_PASSWORD}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable"

MIGRATIONS_DIR := migrations

.PHONY: migrate-up migrate-down run

run:
	-@CONFIG_PATH=./.env go run ./cmd

migrate-up:
	@goose -dir ${MIGRATIONS_DIR} postgres ${GOOSE_DSN} up

migrate-down:
	@goose -dir ${MIGRATIONS_DIR} postgres ${GOOSE_DSN} down
