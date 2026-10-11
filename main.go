package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://app:bygvyt@localhost:5433/todo"
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /todos", func(w http.ResponseWriter, r *http.Request) {
		rows, err := pool.Query(r.Context(), "SELECT id, title, done FROM todos ORDER BY id")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		todos := []Todo{}
		for rows.Next() {
			var t Todo
			if err := rows.Scan(&t.ID, &t.Title, &t.Done); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			todos = append(todos, t)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todos)
	})

	mux.HandleFunc("POST /todos", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Title == "" {
			http.Error(w, "title wajib", 400)
			return
		}
		var t Todo
		err := pool.QueryRow(r.Context(),
			"INSERT INTO todos (title) VALUES ($1) RETURNING id, title, done", in.Title,
		).Scan(&t.ID, &t.Title, &t.Done)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(t)
	})

	log.Println("jalan di :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
