# Pulse

Pulse is a minimal observability backend that ingests structured logs and lets you query them using filters and time windows.  
It stores events in Postgres (including JSON attributes) and exposes a simple HTTP API.

---

## What it does
- Ingest structured log events (`service`, `level`, `message`, optional `trace_id` / `span_id`, and JSON attributes)
- Query recent logs with filters: `service`, `level`, `since`, `until`, `limit`
- Store data in Postgres with indexes for fast time-based queries

---

## Architecture
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


---

## API

### Health
'''bash
curl http://localhost:8080/health

Ingest a log
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{"service":"auth","level":"info","message":"login ok"}'
Query logs
curl "http://localhost:8080/logs?limit=5"
