package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/JpUnique/payflow/internal/config"
	"github.com/JpUnique/payflow/internal/endpoint"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	db, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	err = db.Ping(ctx)
	if err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/payment", endpoint.CreatePayment)

	log.Println("Starting server on :8080")

	log.Fatal(http.ListenAndServe(":8080", mux))

}
func healthHandler(w http.ResponseWriter, r *http.Request) {
	err := db.Ping(r.Context())
	if err != nil {
		http.Error(w, "Database connection error", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
