package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ActionRepository struct {
	pool *pgxpool.Pool
}

func NewActionRepository(pool *pgxpool.Pool) *ActionRepository {
	return &ActionRepository{pool: pool}
}

func (r *ActionRepository) GetByUserAndYear(
	ctx context.Context,
	userID uuid.UUID,
	year int,
) ([]domain.Action, error) {
	const query = `
		SELECT a.id, a.user_id, a.type, a.category_id, c.name, a.created_at
		FROM actions a
		JOIN categories c ON c.id = a.category_id
		WHERE a.user_id = $1
		  AND a.created_at >= $2
		  AND a.created_at < $3
		ORDER BY a.created_at, a.id`

	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	rows, err := r.pool.Query(ctx, query, userID, start, end)
	if err != nil {
		return nil, fmt.Errorf("list user actions: %w", err)
	}
	defer rows.Close()

	actions := make([]domain.Action, 0)
	for rows.Next() {
		var action domain.Action
		if err := rows.Scan(
			&action.ID,
			&action.UserID,
			&action.Type,
			&action.CategoryID,
			&action.Category,
			&action.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}
		actions = append(actions, action)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate actions: %w", err)
	}
	return actions, nil
}
