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
