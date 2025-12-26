package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/JpUnique/payflow/internal/config"
	"github.com/JpUnique/payflow/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

var (
	db  *pgxpool.Pool
	rdb *redis.Client
)

func TestMain(m *testing.M) {
	// Load test config (can reuse normal config if env vars are set)
	cfg := config.Load()

	var err error
	db, err = pgxpool.New(context.Background(), cfg.DBUrl)
	if err != nil {
		panic(err)
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	// Run tests
	code := m.Run()

	// Cleanup
	db.Close()
	rdb.Close()

	os.Exit(code)
}

func cleanDatabase(t *testing.T) {
	ctx := context.Background()

	_, err := db.Exec(ctx, `DELETE FROM idempotency_keys`)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `DELETE FROM transactions`)
	require.NoError(t, err)

	rdb.FlushDB(ctx)
}

func TestCreateIdempotentPayment(t *testing.T) {
	cleanDatabase(t)

	ctx := context.Background()
	repo := repository.NewTransactionRepository(db, rdb)

	idempotencyKey := "integration-test-key-001"
	amount := int64(50000)
	currency := "USD"

	// 🔹 First request
	txID1, idem1, err := repo.CreateIdempotent(
		ctx,
		idempotencyKey,
		amount,
		currency,
	)

	require.NoError(t, err)
	require.False(t, idem1)
	require.NotEqual(t, uuid.Nil, txID1)

	// 🔹 Second request (same key)
	txID2, idem2, err := repo.CreateIdempotent(
		ctx,
		idempotencyKey,
		amount,
		currency,
	)

	require.NoError(t, err)
	require.True(t, idem2)
	require.Equal(t, txID1, txID2)

	// 🔹 Verify only ONE transaction exists
	var txCount int
	err = db.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM transactions`,
	).Scan(&txCount)
	require.NoError(t, err)
	require.Equal(t, 1, txCount)

	// 🔹 Verify idempotency key exists
	var storedTxID uuid.UUID
	err = db.QueryRow(
		ctx,
		`SELECT transaction_id FROM idempotency_keys WHERE key = $1`,
		idempotencyKey,
	).Scan(&storedTxID)
	require.NoError(t, err)
	require.Equal(t, txID1, storedTxID)

	// 🔹 Verify Redis cache
	val, err := rdb.Get(ctx, idempotencyKey).Result()
	require.NoError(t, err)
	require.Equal(t, txID1.String(), val)
}

func TestIdempotencyTTLExpiry(t *testing.T) {
	cleanDatabase(t)

	ctx := context.Background()
	repo := repository.NewTransactionRepository(db, rdb)

	key := "ttl-test-key"
	amount := int64(1000)
	currency := "EUR"

	txID1, idem1, err := repo.CreateIdempotent(ctx, key, amount, currency)
	require.NoError(t, err)
	require.False(t, idem1)

	// Simulate Redis TTL expiry
	rdb.Del(ctx, key)

	time.Sleep(100 * time.Millisecond)

	txID2, idem2, err := repo.CreateIdempotent(ctx, key, amount, currency)
	require.NoError(t, err)
	require.True(t, idem2)
	require.Equal(t, txID1, txID2)

	// Still only one transaction
	var count int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM transactions`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
