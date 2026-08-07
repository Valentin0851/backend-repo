package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	apperrors "github.com/Valentin0851/avito-recap-backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecapRepository struct {
	pool *pgxpool.Pool
}

func NewRecapRepository(pool *pgxpool.Pool) *RecapRepository {
	return &RecapRepository{pool: pool}
}

func (r *RecapRepository) Save(ctx context.Context, recap *domain.Recap) error {
	topCategories, err := json.Marshal(recap.TopCategories)
	if err != nil {
		return fmt.Errorf("marshal top categories: %w", err)
	}
	achievements, err := json.Marshal(recap.Achievements)
	if err != nil {
		return fmt.Errorf("marshal achievements: %w", err)
	}
	summary, err := json.Marshal(recap.Summary)
	if err != nil {
		return fmt.Errorf("marshal summary: %w", err)
	}
	cards, err := json.Marshal(recap.Cards)
	if err != nil {
		return fmt.Errorf("marshal cards: %w", err)
	}

	const query = `
		INSERT INTO recaps (
			id, user_id, year, total_views, total_messages, total_favorites,
			total_purchases, total_sales, top_categories, achievements,
			activity_days, summary, cards, generated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (user_id, year) DO UPDATE SET
			id = EXCLUDED.id,
			total_views = EXCLUDED.total_views,
			total_messages = EXCLUDED.total_messages,
			total_favorites = EXCLUDED.total_favorites,
			total_purchases = EXCLUDED.total_purchases,
			total_sales = EXCLUDED.total_sales,
			top_categories = EXCLUDED.top_categories,
			achievements = EXCLUDED.achievements,
			activity_days = EXCLUDED.activity_days,
			summary = EXCLUDED.summary,
			cards = EXCLUDED.cards,
			generated_at = EXCLUDED.generated_at`

	if _, err := r.pool.Exec(ctx, query,
		recap.ID,
		recap.UserID,
		recap.Year,
		recap.TotalViews,
		recap.TotalMessages,
		recap.TotalFavorites,
		recap.TotalPurchases,
		recap.TotalSales,
		topCategories,
		achievements,
		recap.ActivityDays,
		summary,
		cards,
		recap.GeneratedAt,
	); err != nil {
		return fmt.Errorf("save recap: %w", err)
	}
	return nil
}

func (r *RecapRepository) GetByUserAndYear(
	ctx context.Context,
	userID uuid.UUID,
	year int,
) (*domain.Recap, error) {
	const query = `
		SELECT id, user_id, year, total_views, total_messages, total_favorites,
		       total_purchases, total_sales, top_categories, achievements,
		       activity_days, summary, cards, generated_at
		FROM recaps
		WHERE user_id = $1 AND year = $2`

	var recap domain.Recap
	var topCategories []byte
	var achievements []byte
	var summary []byte
	var cards []byte
	err := r.pool.QueryRow(ctx, query, userID, year).Scan(
		&recap.ID,
		&recap.UserID,
		&recap.Year,
		&recap.TotalViews,
		&recap.TotalMessages,
		&recap.TotalFavorites,
		&recap.TotalPurchases,
		&recap.TotalSales,
		&topCategories,
		&achievements,
		&recap.ActivityDays,
		&summary,
		&cards,
		&recap.GeneratedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get recap: %w", err)
	}
	if err := json.Unmarshal(topCategories, &recap.TopCategories); err != nil {
		return nil, fmt.Errorf("unmarshal top categories: %w", err)
	}
	if err := json.Unmarshal(achievements, &recap.Achievements); err != nil {
		return nil, fmt.Errorf("unmarshal achievements: %w", err)
	}
	if err := json.Unmarshal(summary, &recap.Summary); err != nil {
		return nil, fmt.Errorf("unmarshal summary: %w", err)
	}
	if err := json.Unmarshal(cards, &recap.Cards); err != nil {
		return nil, fmt.Errorf("unmarshal cards: %w", err)
	}
	if recap.TopCategories == nil {
		recap.TopCategories = []domain.CategoryStat{}
	}
	if recap.Achievements == nil {
		recap.Achievements = []domain.Achievement{}
	}
	if recap.Cards == nil {
		recap.Cards = []domain.RecapCard{}
	}
	return &recap, nil
}
