package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(
	ctx context.Context,
	id uuid.UUID,
	amount int64,
	currency string,
) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO transactions (id, amount, currency, status)
		 VALUES ($1, $2, $3, $4)`,
		id, amount, currency, "PENDING",
	)
	return err
}
