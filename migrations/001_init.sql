CREATE TABLE IF NOT EXISTS services (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS events (
  id BIGSERIAL PRIMARY KEY,
  service_id BIGINT NOT NULL REFERENCES services(id),
  level TEXT NOT NULL CHECK (level IN ('debug','info','warn','error')),
  message TEXT NOT NULL,
  trace_id TEXT,
  span_id TEXT,
  occurred_at TIMESTAMPTZ NOT NULL,
  attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_occurred_at
  ON events(occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_events_service_level_time
  ON events(service_id, level, occurred_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_services_name_unique
  ON services(name);

CREATE INDEX IF NOT EXISTS idx_events_service_id ON events(service_id);
