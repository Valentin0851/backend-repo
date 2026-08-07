package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Valentin0851/avito-recap-backend/internal/config"
	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	"github.com/Valentin0851/avito-recap-backend/internal/repository/postgres"
	"github.com/Valentin0851/avito-recap-backend/migrations"
	apperrors "github.com/Valentin0851/avito-recap-backend/pkg/errors"
	"github.com/google/uuid"
)

func TestRegistrationTransactionRollsBack(t *testing.T) {
	if os.Getenv("DB_PORT") == "" {
		t.Skip("set DB_PORT to run the PostgreSQL integration test")
	}

	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL())
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	if err := migrations.Apply(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	account := &domain.Account{
		ID:        uuid.New(),
		Login:     "rollback-" + uuid.NewString()[:8],
		CreatedAt: time.Now().UTC(),
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM accounts WHERE id = $1`, account.ID)
	}()

	repository := postgres.NewAuthRepository(pool)
	err = repository.CreateAccountWithSession(
		ctx,
		account,
		"test-password-hash",
		[]byte("invalid-length"),
		time.Now().UTC().Add(time.Hour),
	)
	if err == nil {
		t.Fatal("registration unexpectedly succeeded with invalid session hash")
	}

	if _, _, err := repository.GetAccountByLogin(ctx, account.Login); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("account survived failed registration transaction: %v", err)
	}
}
