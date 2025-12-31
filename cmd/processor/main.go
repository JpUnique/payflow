package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/JpUnique/payflow/internal/config"
	"github.com/JpUnique/payflow/internal/messaging"
	"github.com/JpUnique/payflow/internal/repository"
	"github.com/JpUnique/payflow/internal/events"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	db, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	repo := repository.NewTransactionRepository(db, nil)

	consumer := messaging.NewKafkaConsumer(
		cfg.KafkaBrokers,
		"payment.created",
		"payment-processor-group",
	)

	log.Println("[PROCESSOR] started")

	err = consumer.Start(ctx, func(event events.PaymentCreatedEvent) error {

		// Simulate processing delay
		time.Sleep(2 * time.Second)

		// Random success/failure
		status := "SUCCESS"
		if rand.Intn(10) < 2 {
			status = "FAILED"
		}

		log.Printf(
			"[PROCESSOR] updating tx_id=%s status=%s",
			event.TransactionID,
			status,
		)

		return repo.UpdateStatus(ctx, event.TransactionID, status)
	})

	if err != nil {
		log.Fatal(err)
	}
}
