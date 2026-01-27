# Effective Mobile Task

## Установка и настройка проекта

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

После установки убедитесь, что goose доступен в PATH:

```bash
goose --version
```

### 2. Заполнение .env файла

Создайте файл `.env` на основе примера:

```bash
cp .env.example .env
```

Откройте `.env` и при необходимости измените переменные окружения:

```env
# PostgreSQL
DB_USER=admin
DB_PASSWORD=admin
DB_NAME=subscriptions_db
DB_HOST=db
DB_PORT=5432

# Server
SERVER_PORT=8080
```

### 3. Запуск базы данных:

```bash
docker-compose up -d
```

### 4. Применение миграций

Примените миграции к базе данных:

```bash
make migrate-up
```

## Команды для работы с миграциями

- `make migrate-up` — применить все миграции
- `make migrate-down` — откатить последнюю миграцию
