package repository

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type TransactionRepository struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewTransactionRepository(db *pgxpool.Pool, rdb *redis.Client) *TransactionRepository {
	return &TransactionRepository{db: db, rdb: rdb}
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

func (r *TransactionRepository) CreateIdempotent(
	ctx context.Context,
	idempotencyKey string,
	amount int64,
	currency string,
) (uuid.UUID, bool, error) {

	// check redis cache first
	val, err := r.rdb.Get(ctx, idempotencyKey).Result()
	if err == nil {
		existingTxID, _ := uuid.Parse(val)
		return existingTxID, true, nil
	}
	// redis miss, proceed with DB transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		log.Printf(
			"[IDEMPOTENCY] redis_hit key=%s tx_id=%s",
			idempotencyKey,
			val,
		)
		return uuid.Nil, false, err
	}
	defer tx.Rollback(ctx)

	//  Check if idempotency key exists
	var existingTxID uuid.UUID
	err = tx.QueryRow(
		ctx,
		`SELECT transaction_id FROM idempotency_keys WHERE key = $1`,
		idempotencyKey,
	).Scan(&existingTxID)

	if err == nil {
		log.Printf(
			"[IDEMPOTENCY] db_hit key=%s tx_id=%s",
			idempotencyKey,
			existingTxID,
		)
		// Found in postgres then write to redis for next time caching
		r.rdb.Set(ctx, idempotencyKey, existingTxID.String(), 24*time.Hour)
		return existingTxID, true, nil
	}
	if err != pgx.ErrNoRows {
		return uuid.Nil, false, err
	}

	// Insert new transaction
	txID := uuid.New()
	_, err = tx.Exec(
		ctx,
		`INSERT INTO transactions (id, amount, currency, status)
		 VALUES ($1, $2, $3, $4)`,
		txID, amount, currency, "PENDING",
	)
	if err != nil {
		return uuid.Nil, false, err
	}

	// Store idempotency key
	_, err = tx.Exec(
		ctx,
		`INSERT INTO idempotency_keys (key, transaction_id)
		 VALUES ($1, $2)`,
		idempotencyKey, txID,
	)
	if err != nil {
		return uuid.Nil, false, err
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, false, err
	}

	// writre to redis cache
	r.rdb.Set(ctx, idempotencyKey, txID.String(), 5*time.Minute)

	return txID, false, nil
}
