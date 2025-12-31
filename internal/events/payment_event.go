package events

import "time"

type PaymentCreatedEvent struct {
	EventType     string    `json:"event_type"`
	TransactionID string    `json:"transaction_id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
