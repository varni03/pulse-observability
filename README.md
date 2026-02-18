Pulse

Pulse is a minimal observability backend that ingests structured logs and lets you query them using filters and time windows.
It stores events in Postgres (including JSON attributes) and exposes a simple HTTP API.

What it does

Ingest structured log events (service, level, message, optional trace_id / span_id, and JSON attributes)

Query recent logs with filters: service, level, since, until, limit

Store data in Postgres with indexes for fast time-based queries

Architecture
Client / Service
   |
   |  POST /logs (JSON)
   v
Pulse API (Go)
   |
   |  INSERT (service upsert + event)
   v
Postgres (events + attributes JSONB)
   |
   |  GET /logs?filters
   v
User / CLI / Dashboard

API
Health
curl http://localhost:8080/health

Ingest a log
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{
    "service":"auth",
    "level":"info",
    "message":"login ok",
    "trace_id":"t1",
    "span_id":"s1",
    "attributes":{"user_id":123,"ip":"1.2.3.4","latency_ms":42}
  }'

Query logs
curl "http://localhost:8080/logs?limit=5"


Filter:

curl "http://localhost:8080/logs?service=auth&level=error&limit=10"


Time window:

SINCE=$(date -u -v-10M +"%Y-%m-%dT%H:%M:%SZ")
curl "http://localhost:8080/logs?since=$SINCE&limit=10"

Data model

services

id, name, created_at

events

id, service_id, level, message

trace_id, span_id

occurred_at

attributes (JSONB)

created_at

Indexes

events(occurred_at DESC)

events(service_id, level, occurred_at DESC)

Roadmap

/metrics endpoint for ingest and query stats

simple alerting rules (e.g. error-rate thresholds)

linking traces via trace_id / span_id
