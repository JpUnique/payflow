package endpoint

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/JpUnique/payflow/internal/events"
	"github.com/JpUnique/payflow/internal/messaging"
	"github.com/JpUnique/payflow/internal/repository"
)

type PaymentRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func CreatePayment(repo *repository.TransactionRepository, producer *messaging.KafkaProducer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Extract Idempotency-Key from headers
		idempotencyKey := r.Header.Get("Idempotency-Key")
		if idempotencyKey == "" {
			http.Error(w, "Missing Idempotency-Key header", http.StatusBadRequest)
			return
		}

		var req PaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Idempotent transaction creation
		txID, isIdemtpotent, err := repo.CreateIdempotent(r.Context(), idempotencyKey, req.Amount, req.Currency)
		if err != nil {
			http.Error(w, "Failed to process idempotent transaction", http.StatusInternalServerError)
			return
		}
		// Emit Kafka event (async, non-blocking)
		event := events.PaymentCreatedEvent{
			EventType:     "payment.created",
			TransactionID: txID.String(),
			Amount:        req.Amount,
			Currency:      req.Currency,
			Status:        "PENDING",
			CreatedAt:     time.Now().UTC(),
		}

		go func() {
			err := producer.Publish(r.Context(), txID.String(), event)
			if err != nil {
				log.Printf("[WARNING] failed to publish payment.created event for transaction_id=%s: err=%v", txID, err)
			}
		}()

		response := map[string]interface{}{
			"transaction_id": txID.String(),
			"status":         "PENDING",
			"idempotent":     isIdemtpotent,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
