package ports

import (
	"context"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	"github.com/google/uuid"
)

type ActionRepository interface {
	GetByUserAndYear(ctx context.Context, userID uuid.UUID, year int) ([]domain.Action, error)
}
