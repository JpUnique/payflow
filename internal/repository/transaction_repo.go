package repository

import (
	"github.com/JpUnique/payflow/internal/database"
	"github.com/google/uuid"
)

func CreateTransaction(id uuid.UUID, amount int64, currency string) error {
	_, err := database.DB.Exec(
		"INSERT INTO transactions (id, amount, currency, status) VALUES ($1, $2, $3, $4)",
		id, amount, currency, "PENDING",
	)
	return err
}
