package redis

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	"github.com/Valentin0851/avito-recap-backend/internal/usecase/ports"
	"github.com/google/uuid"
)

const recapCacheKeyVersion = "v1"

// CachedRecapRepository decorates the durable PostgreSQL repository with Redis.
// PostgreSQL remains the source of truth: cache failures never fail API requests.
type CachedRecapRepository struct {
	source ports.RecapRepository
	cache  ports.Cache
	logger *slog.Logger
}

func NewCachedRecapRepository(
	source ports.RecapRepository,
	cache ports.Cache,
	logger *slog.Logger,
) *CachedRecapRepository {
	if logger == nil {
		logger = slog.Default()
	}
	return &CachedRecapRepository{
		source: source,
		cache:  cache,
		logger: logger,
	}
}

func (r *CachedRecapRepository) Save(ctx context.Context, recap *domain.Recap) error {
	if err := r.source.Save(ctx, recap); err != nil {
		return err
	}

	key := recapCacheKey(recap.UserID, recap.Year)
	if err := r.cache.Set(ctx, key, recap, RecapTTL); err != nil {
		r.logger.WarnContext(ctx, "failed to update recap cache",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
	}
	return nil
}

func (r *CachedRecapRepository) GetByUserAndYear(
	ctx context.Context,
	userID uuid.UUID,
	year int,
) (*domain.Recap, error) {
	key := recapCacheKey(userID, year)
	var cached domain.Recap
	found, err := r.cache.Get(ctx, key, &cached)
	if err == nil && found {
		return &cached, nil
	}
	if err != nil {
		r.logger.WarnContext(ctx, "failed to read recap cache",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		if deleteErr := r.cache.Delete(ctx, key); deleteErr != nil {
			r.logger.WarnContext(ctx, "failed to evict invalid recap cache entry",
				slog.String("key", key),
				slog.String("error", deleteErr.Error()),
			)
		}
	}

	recap, err := r.source.GetByUserAndYear(ctx, userID, year)
	if err != nil {
		return nil, err
	}
	if err := r.cache.Set(ctx, key, recap, RecapTTL); err != nil {
		r.logger.WarnContext(ctx, "failed to populate recap cache",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
	}
	return recap, nil
}

func recapCacheKey(userID uuid.UUID, year int) string {
	return fmt.Sprintf("recap:%s:%s:%d", recapCacheKeyVersion, userID, year)
}

var _ ports.RecapRepository = (*CachedRecapRepository)(nil)
