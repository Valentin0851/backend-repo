package ports

import (
	"context"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetByID(ctx context.Context, accountID, id uuid.UUID) (*domain.User, error)
	ListProfiles(ctx context.Context, accountID uuid.UUID) ([]domain.User, error)
	Create(ctx context.Context, accountID uuid.UUID, user *domain.User) error
	Update(ctx context.Context, accountID uuid.UUID, user *domain.User) error
	Delete(ctx context.Context, accountID, id uuid.UUID) error
}
