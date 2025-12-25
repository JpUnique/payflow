package endpoint

import (
	"encoding/json"
	"net/http"

	"github.com/JpUnique/payflow/internal/repository"
	"github.com/google/uuid"
)

type PaymentRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func CreatePayment(repo *repository.TransactionRepository) http.HandlerFunc {
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

		txID := uuid.New()

		err := repo.Create(
			r.Context(),
			txID,
			req.Amount,
			req.Currency,
		)
		if err != nil {
			http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
			return
		}

		// Idempotent transaction creation
		txID, isIdemtpotent, err := repo.CreateIdempotent(r.Context(), idempotencyKey, req.Amount, req.Currency)
		if err != nil {
			http.Error(w, "Failed to process idempotent transaction", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"transaction_id": txID.String(),
			"status":         "PENDING",
			"idempotent":     isIdemtpotent,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
