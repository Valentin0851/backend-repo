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

type RecapCacheMetrics interface {
	Hit()
	Miss()
	Error(operation string)
}

type noopRecapCacheMetrics struct{}

func (noopRecapCacheMetrics) Hit()         {}
func (noopRecapCacheMetrics) Miss()        {}
func (noopRecapCacheMetrics) Error(string) {}

// CachedRecapRepository decorates the durable PostgreSQL repository with Redis.
// PostgreSQL remains the source of truth: cache failures never fail API requests.
type CachedRecapRepository struct {
	source  ports.RecapRepository
	cache   ports.Cache
	logger  *slog.Logger
	metrics RecapCacheMetrics
}

func NewCachedRecapRepository(
	source ports.RecapRepository,
	cache ports.Cache,
	logger *slog.Logger,
	cacheMetrics ...RecapCacheMetrics,
) *CachedRecapRepository {
	if logger == nil {
		logger = slog.Default()
	}
	var metrics RecapCacheMetrics = noopRecapCacheMetrics{}
	if len(cacheMetrics) > 0 && cacheMetrics[0] != nil {
		metrics = cacheMetrics[0]
	}
	return &CachedRecapRepository{
		source:  source,
		cache:   cache,
		logger:  logger,
		metrics: metrics,
	}
}

func (r *CachedRecapRepository) Save(ctx context.Context, recap *domain.Recap) error {
	if err := r.source.Save(ctx, recap); err != nil {
		return err
	}

	key := recapCacheKey(recap.UserID, recap.Year)
	if err := r.cache.Set(ctx, key, recap, RecapTTL); err != nil {
		r.metrics.Error("set")
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
		r.metrics.Hit()
		return &cached, nil
	}
	if err == nil {
		r.metrics.Miss()
	}
	if err != nil {
		r.metrics.Error("get")
		r.logger.WarnContext(ctx, "failed to read recap cache",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		if deleteErr := r.cache.Delete(ctx, key); deleteErr != nil {
			r.metrics.Error("delete")
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
		r.metrics.Error("set")
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
