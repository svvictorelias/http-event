package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Event represents the persisted event
type Event struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
}

type createEventRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
}

var db *sql.DB

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	var err error
	// sql.Open does not establish connections immediately, but returns a DB handle safe for concurrent use
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	db.SetConnMaxIdleTime(5 * time.Minute)

	// verify connectivity with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// small timeout middleware for safety
	r.Use(middleware.Timeout(15 * time.Second))

	r.Post("/events", createEventHandler)
	r.Get("/events", listEventsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func createEventHandler(w http.ResponseWriter, r *http.Request) {
	var req createEventRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON payload: "+err.Error())
		return
	}

	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		respondError(w, http.StatusBadRequest, "start_time must be RFC3339 timestamp")
		return
	}
	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		respondError(w, http.StatusBadRequest, "end_time must be RFC3339 timestamp")
		return
	}

	id := uuid.New()
	createdAt := time.Now().UTC()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Insert into DB
	query := `INSERT INTO events (id, title, description, start_time, end_time, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = db.ExecContext(ctx, query, id, req.Title, req.Description, start, end, createdAt)
	if err != nil {
		log.Printf("insert error: %v", err)
		respondError(w, http.StatusInternalServerError, "could not insert event")
		return
	}

	e := Event{
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

// listEventsHandler handles GET /events
func listEventsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `SELECT id, title, description, start_time, end_time, created_at FROM events ORDER BY start_time ASC`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("select error: %v", err)
		respondError(w, http.StatusInternalServerError, "could not query events")
		return
	}
	defer rows.Close()

}

func respondError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
