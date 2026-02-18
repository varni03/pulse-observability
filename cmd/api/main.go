package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type healthResponse struct {
	Status string `json:"status"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

type LogEvent struct {
	Service   string    `json:"service"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"trace_id,omitempty"`
	SpanID    string    `json:"span_id,omitempty"`
}

var allowedLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

func logsQueryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		q := r.URL.Query()
		service := strings.TrimSpace(q.Get("service"))
		level := strings.ToLower(strings.TrimSpace(q.Get("level")))

		limit := 50
		if v := strings.TrimSpace(q.Get("limit")); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				if n < 1 {
					n = 1
				}
				if n > 200 {
					n = 200
				}
				limit = n
			}
		}

		var since *time.Time
		if v := strings.TrimSpace(q.Get("since")); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				writeError(w, http.StatusBadRequest, "since must be RFC3339")
				return
			}
			tt := t.UTC()
			since = &tt
		}

		var until *time.Time
		if v := strings.TrimSpace(q.Get("until")); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				writeError(w, http.StatusBadRequest, "until must be RFC3339")
				return
			}
			tt := t.UTC()
			until = &tt
		}

		if level != "" && !allowedLevels[level] {
			writeError(w, http.StatusBadRequest, "level must be one of: debug, info, warn, error")
			return
		}

		sqlStr := `
SELECT e.id, s.name, e.level, e.message, e.trace_id, e.span_id, e.occurred_at, e.attributes
FROM events e
JOIN services s ON s.id = e.service_id
WHERE ($1 = '' OR s.name = $1)
  AND ($2 = '' OR e.level = $2)
  AND ($3::timestamptz IS NULL OR e.occurred_at >= $3)
  AND ($4::timestamptz IS NULL OR e.occurred_at <= $4)
ORDER BY e.occurred_at DESC
LIMIT $5;
`
		rows, err := db.Query(sqlStr, service, level, since, until, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to query logs")
			return
		}
		defer rows.Close()

		out := make([]StoredEvent, 0, limit)
		for rows.Next() {
			var e StoredEvent
			var attrsBytes []byte
			if err := rows.Scan(&e.ID, &e.Service, &e.Level, &e.Message, &e.TraceID, &e.SpanID, &e.OccurredAt, &attrsBytes); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to read logs")
				return
			}

			e.OccurredAt = e.OccurredAt.UTC()

			e.Attributes = map[string]any{}
			if len(attrsBytes) > 0 {
				_ = json.Unmarshal(attrsBytes, &e.Attributes)
			}

			out = append(out, e)

		}
		if err := rows.Err(); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read logs")
			return
		}

		writeJSON(w, http.StatusOK, out)
	}
}

func insertEvent(db *sql.DB, e LogEvent, attrs map[string]any) (int64, time.Time, []byte, error) {
	// Upsert service and get id
	var serviceID int64
	err := db.QueryRow(
		`INSERT INTO services(name) VALUES ($1)
		 ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id`,
		e.Service,
	).Scan(&serviceID)
	if err != nil {
		return 0, time.Time{}, nil, err
	}

	attrsBytes := []byte(`{}`)
	if attrs != nil {
		b, err := json.Marshal(attrs)
		if err != nil {
			return 0, time.Time{}, nil, err
		}
		attrsBytes = b
	}

	// Insert + return what was stored
	var id int64
	var occurredAt time.Time
	var storedAttrs []byte

	err = db.QueryRow(
		`INSERT INTO events(service_id, level, message, trace_id, span_id, occurred_at, attributes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, occurred_at, attributes`,
		serviceID,
		e.Level,
		e.Message,
		nullIfEmpty(e.TraceID),
		nullIfEmpty(e.SpanID),
		e.Timestamp,
		attrsBytes,
	).Scan(&id, &occurredAt, &storedAttrs)
	if err != nil {
		return 0, time.Time{}, nil, err
	}

	return id, occurredAt, storedAttrs, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func logsIngestHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var in struct {
			Service    string         `json:"service"`
			Level      string         `json:"level"`
			Message    string         `json:"message"`
			TraceID    string         `json:"trace_id"`
			SpanID     string         `json:"span_id"`
			Attributes map[string]any `json:"attributes"`
		}

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		// reject trailing junk after the first JSON object
		if err := dec.Decode(&struct{}{}); err == nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		in.Service = strings.TrimSpace(in.Service)
		in.Level = strings.ToLower(strings.TrimSpace(in.Level))
		in.Message = strings.TrimSpace(in.Message)

		if in.Service == "" {
			writeError(w, http.StatusBadRequest, "service is required")
			return
		}
		if in.Message == "" {
			writeError(w, http.StatusBadRequest, "message is required")
			return
		}
		if !allowedLevels[in.Level] {
			writeError(w, http.StatusBadRequest, "level must be one of: debug, info, warn, error")
			return
		}

		event := LogEvent{
			Service:   in.Service,
			Level:     in.Level,
			Message:   in.Message,
			Timestamp: time.Now().UTC(),
			TraceID:   strings.TrimSpace(in.TraceID),
			SpanID:    strings.TrimSpace(in.SpanID),
		}

		id, occurredAt, storedAttrs, err := insertEvent(db, event, in.Attributes)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to persist log event")
			return
		}

		out := StoredEvent{
			ID:         id,
			Service:    event.Service,
			Level:      event.Level,
			Message:    event.Message,
			OccurredAt: occurredAt.UTC(),
			Attributes: map[string]any{},
		}

		if strings.TrimSpace(event.TraceID) != "" {
			t := event.TraceID
			out.TraceID = &t
		}
		if strings.TrimSpace(event.SpanID) != "" {
			s := event.SpanID
			out.SpanID = &s
		}

		if len(storedAttrs) > 0 {
			_ = json.Unmarshal(storedAttrs, &out.Attributes)
		}

		writeJSON(w, http.StatusCreated, out)

	}
}

type StoredEvent struct {
	ID         int64          `json:"id"`
	Service    string         `json:"service"`
	Level      string         `json:"level"`
	Message    string         `json:"message"`
	TraceID    *string        `json:"trace_id,omitempty"`
	SpanID     *string        `json:"span_id,omitempty"`
	OccurredAt time.Time      `json:"timestamp"`
	Attributes map[string]any `json:"attributes"`
}

func main() {
	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			logsIngestHandler(db)(w, r)
			return
		}
		if r.Method == http.MethodGet {
			logsQueryHandler(db)(w, r)
			return
		}
		w.Header().Set("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)

	})
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
