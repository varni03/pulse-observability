# Pulse

> A lightweight observability backend for ingesting and querying structured logs.

Pulse is a minimal observability backend that ingests structured logs and lets you query them using filters and time windows.  
It stores events in Postgres (including JSON attributes) and exposes a simple HTTP API.

---
## Why I built this

I wanted to better understand how observability systems work, beyond using existing tools. Most logging platforms abstract away the ingestion and storage layers, so I built Pulse to explore those pieces directly.

This project focuses on the core mechanics of a logging backend: accepting structured events over HTTP, normalizing and storing them efficiently, and supporting flexible queries over time windows and filters. Building it from scratch helped me understand tradeoffs around schema design, indexing for time-series data, and handling semi-structured JSON attributes in Postgres.

Pulse is intentionally minimal, but it serves as a foundation for experimenting with observability features like metrics, alerting, and trace correlation.

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



