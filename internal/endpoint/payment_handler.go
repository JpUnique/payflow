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

func CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	 err = repository.Create(uuid.New(), req.Amount, req.Currency)
	 if err != nil {
	     http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
	     return
	}

	// For demonstration, we will just return a success response
	response := map[string]interface{}{
		"transaction_id": uuid.New().String(),
		"status":         "PENDING",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
