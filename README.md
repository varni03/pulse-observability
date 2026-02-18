# Pulse

Pulse is a minimal observability backend that ingests structured logs and lets you query them with filters and time windows.  
It stores events in Postgres (including JSON attributes) and exposes a simple HTTP API.

## What it does
- Ingest log events with `service`, `level`, `message`, optional `trace_id`/`span_id`, and `attributes` (JSON)
- Query recent logs with filters: `service`, `level`, `since`, `until`, `limit`
- Stores data in Postgres with indexes for fast queries

## API

### Health
```bash
curl -s http://localhost:8080/health | python -m json.tool

Ingest a log
curl -s -X POST "http://localhost:8080/logs" \
  -H "Content-Type: application/json" \
  -d '{
    "service":"auth",
    "level":"info",
    "message":"login ok",
    "trace_id":"t1",
    "span_id":"s1",
    "attributes":{"user_id":123,"ip":"1.2.3.4","latency_ms":42}
  }' | python -m json.tool

Query logs
curl -s "http://localhost:8080/logs?limit=5" | python -m json.tool


Filter:

curl -s "http://localhost:8080/logs?service=auth&level=error&limit=10" | python -m json.tool


Time window:

SINCE=$(date -u -v-10M +"%Y-%m-%dT%H:%M:%SZ")
curl -s "http://localhost:8080/logs?since=$SINCE&limit=10" | python -m json.tool

Data model

services(id, name, created_at)

events(id, service_id, level, message, trace_id, span_id, occurred_at, attributes, created_at)

Indexes:

events(occurred_at DESC)

events(service_id, level, occurred_at DESC)

Next

metrics endpoint for ingest rate and error rate

simple alerting rules (error rate thresholds per service)

tracing links via trace_id/span_id


## Tiny architecture diagram (add under “What it does” if you want)
Paste this right after “What it does”:

```md
```text
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
User / Dashboard


## Next build step (real observability)
Add a **/metrics** endpoint (Prometheus style) that exposes:
- total ingested events
- ingests per level
- last ingest timestamp
- query count

Say “next” and I’ll give you the cleanest minimal `/metrics` implementation that fits your current style.
