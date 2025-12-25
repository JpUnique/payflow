package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/JpUnique/payflow/internal/config"
	"github.com/JpUnique/payflow/internal/endpoint"
	"github.com/JpUnique/payflow/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//postgres connection
	db, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	//Redis connection could be added here
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	defer rdb.Close()

	//repository initialization
	txRepo := repository.NewTransactionRepository(db, rdb)

	//HTTP server setup
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(r.Context()); err != nil {
			http.Error(w, "Database connection error", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("OK"))
	})

	mux.Handle("/payment", endpoint.CreatePayment(txRepo))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
