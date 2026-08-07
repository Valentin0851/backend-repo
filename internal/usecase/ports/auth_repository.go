package ports

import (
	"context"
	"time"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	"github.com/google/uuid"
)

type AuthRepository interface {
	CreateAccountWithSession(
		ctx context.Context,
		account *domain.Account,
		passwordHash string,
		tokenHash []byte,
		expiresAt time.Time,
	) error
	GetAccountByLogin(ctx context.Context, login string) (*domain.Account, string, error)
	CreateSession(ctx context.Context, accountID uuid.UUID, tokenHash []byte, expiresAt time.Time) error
	GetAccountBySessionHash(ctx context.Context, tokenHash []byte, now time.Time) (*domain.Account, error)
	DeleteSession(ctx context.Context, tokenHash []byte) error
}
