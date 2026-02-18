# Pulse

> A lightweight observability backend for ingesting and querying structured logs.

Pulse is a minimal observability backend that ingests structured logs and lets you query them using filters and time windows.  
It stores events in Postgres (including JSON attributes) and exposes a simple HTTP API.

---
## Why I built this

I built Pulse to better understand how observability systems work at a systems level. Instead of relying on existing logging platforms, I wanted to implement the ingestion, storage, and query layers myself and see the tradeoffs firsthand.

The goal was to explore how structured logs move through a backend: from HTTP ingestion, to schema design and indexing, to efficient querying over time windows and filters. Working on this helped me think more deeply about time-series data modeling, JSON storage in Postgres, and building simple but extensible infrastructure.

Pulse is intentionally small, but it provides a base for experimenting with ideas like metrics, alerting, and trace correlation.

--- 

## What it does

- Ingest structured log events (service, level, message, optional trace_id / span_id, and JSON attributes)
- Query recent logs with filters: service, level, since, until, limit
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
```bash
curl http://localhost:8080/health
```

## Ingest a log
```bash
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{"service":"auth","level":"info","message":"login ok"}'
```
## Query logs
```bash
curl "http://localhost:8080/logs?limit=5"
```

---

Built with Go + Postgres.  
Pulse is a minimal observability backend focused on simplicity, performance, and extensibility.



