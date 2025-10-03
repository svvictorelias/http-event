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
)

type Event struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
}

type createEventRequest struct {
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

var db *sql.DB

func main() {
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal(`dont have an url`)
	}

	var err error

	db, err = sql.Open("postgress", dsn)
	if err != nil {
		log.Fatalf(`Error on connect: %v`, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r := chi.NewRouter()
	r.Use(middleware.Timeout(15 * time.Second))
	r.Post("/events", create)
	r.Get("/events", list)
	r.Get("/events/{id}", find)

	port := 3333

	addr := fmt.Sprintf(":%s", port)
	log.Fatal(http.ListenAndServe(addr, r))
}

func create(w http.ResponseWriter, r *http.Request) {
	var req createEventRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return
	}

	id := uuid.New()
	createdAt := time.Now()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO events (id, title, description, start_time,end_time,created_at) VALUES ($1,$2,$3,$4,$5,$6)`

	_, err := db.ExecContext(ctx, query, id, req.Title, req.Description, req.StartTime, req.EndTime, createdAt)

	if err != nil {
		log.Printf(`error %v:`, err)
		return
	}
	event := Event{
		ID: id,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

func list(w http.ResponseWriter, r *http.Request) {}
