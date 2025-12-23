package payment

import "time"

type Transaction struct {
	ID        string
	Reference string
	Amount    int64
	Currency  string
	Status    string
	CreatedAt time.Time
}
