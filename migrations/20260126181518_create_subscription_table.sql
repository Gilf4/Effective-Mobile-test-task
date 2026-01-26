-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscriptions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  service_name TEXT NOT NULL,
  price INTEGER NOT NULL CHECK (price >= 0),
  user_id UUID NOT NULL,
  start_date DATE NOT NULL,
  end_date DATE NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_start_date ON subscriptions (start_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions (end_date);

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
