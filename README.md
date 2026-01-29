# Effective Mobile Task

## Настройка окружения

Создайте файл `.env` на основе примера:

```bash
cp .env.example .env
```

Откройте `.env` и при необходимости измените переменные окружения(можно не менять):

```env
#APP
APP_ENV=local

# PostgreSQL
DB_USER=admin
DB_PASSWORD=admin
DB_NAME=subscriptions_db
DB_HOST=db
DB_PORT=5432

# Server
SERVER_PORT=8080
READ_TIMEOUT=5s
WRITE_TIMEOUT=5s
```

## Docker Compose

```bash
docker compose up
```

Приложение будет доступно по адресу: `http://localhost:8080`

Swagger документация по адрессу: `http://localhost:8080/swagger`

## Способ 2: Локальная разработка

### 1. Установка goose

Goose — инструмент для управления миграциями базы данных.

**Через Go:**
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

**Через Homebrew (macOS):**
```bash
brew install goose
```

### 2. Запуск базы данных

```bash
docker compose up -d db
```

### 3. Применение миграций

```bash
make migrate-up
```

### 4. Запуск приложения

```bash
make run
```

Приложение будет доступно по адресу: `http://localhost:8080`

## Доступ к API

API документация (Swagger) доступна по адресу:
`http://localhost:8080/swagger`

## Команда для работы с генерацией swagger документации

- `make swag`

## Команды для работы с миграциями

- `make migrate-up` — применить все миграции
- `make migrate-down` — откатить последнюю миграцию
