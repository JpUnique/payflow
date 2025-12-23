package endpoint

import (
	"encoding/json"
	"net/http"

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
	response := map[string]interface{}{
		"transaction_id": uuid.New().String(),
		"status":         "PENDING",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
