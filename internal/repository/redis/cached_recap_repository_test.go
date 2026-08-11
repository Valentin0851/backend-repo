package redis

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	"github.com/google/uuid"
)

type fakeRecapSource struct {
	recap     *domain.Recap
	getErr    error
	saveErr   error
	getCalls  int
	saveCalls int
}

func (s *fakeRecapSource) Save(_ context.Context, recap *domain.Recap) error {
	s.saveCalls++
	if s.saveErr != nil {
		return s.saveErr
	}
	copy := *recap
	s.recap = &copy
	return nil
}

func (s *fakeRecapSource) GetByUserAndYear(
	_ context.Context,
	_ uuid.UUID,
	_ int,
) (*domain.Recap, error) {
	s.getCalls++
	if s.getErr != nil {
		return nil, s.getErr
	}
	copy := *s.recap
	return &copy, nil
}

type fakeCache struct {
	values   map[string][]byte
	getErr   error
	setErr   error
	setCalls int
}

func newFakeCache() *fakeCache {
	return &fakeCache{values: make(map[string][]byte)}
}

func (c *fakeCache) Get(_ context.Context, key string, dest any) (bool, error) {
	if c.getErr != nil {
		return false, c.getErr
	}
	value, ok := c.values[key]
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal(value, dest)
}

func (c *fakeCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	c.setCalls++
	if c.setErr != nil {
		return c.setErr
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.values[key] = encoded
	return nil
}

func (c *fakeCache) Delete(_ context.Context, key string) error {
	delete(c.values, key)
	return nil
}

type fakeRecapCacheMetrics struct {
	hits   int
	misses int
	errors map[string]int
}

func newFakeRecapCacheMetrics() *fakeRecapCacheMetrics {
	return &fakeRecapCacheMetrics{errors: make(map[string]int)}
}

func (m *fakeRecapCacheMetrics) Hit() {
	m.hits++
}

func (m *fakeRecapCacheMetrics) Miss() {
	m.misses++
}

func (m *fakeRecapCacheMetrics) Error(operation string) {
	m.errors[operation]++
}

func TestCachedRecapRepositoryReadsThroughCache(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	cached := &domain.Recap{ID: uuid.New(), UserID: userID, Year: 2026, TotalViews: 42}
	source := &fakeRecapSource{recap: &domain.Recap{ID: uuid.New(), UserID: userID, Year: 2026}}
	cache := newFakeCache()
	if err := cache.Set(ctx, recapCacheKey(userID, 2026), cached, RecapTTL); err != nil {
		t.Fatalf("prime cache: %v", err)
	}
	metrics := newFakeRecapCacheMetrics()

	repository := NewCachedRecapRepository(source, cache, nil, metrics)
	recap, err := repository.GetByUserAndYear(ctx, userID, 2026)
	if err != nil {
		t.Fatalf("get recap: %v", err)
	}
	if recap.ID != cached.ID || recap.TotalViews != 42 {
		t.Fatalf("recap = %#v, want cached recap %#v", recap, cached)
	}
	if source.getCalls != 0 {
		t.Fatalf("source get calls = %d, want 0 on cache hit", source.getCalls)
	}
	if metrics.hits != 1 || metrics.misses != 0 {
		t.Fatalf("cache metrics hits/misses = %d/%d, want 1/0", metrics.hits, metrics.misses)
	}
}

func TestCachedRecapRepositoryPopulatesCacheOnMiss(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	stored := &domain.Recap{ID: uuid.New(), UserID: userID, Year: 2026, TotalSales: 7}
	source := &fakeRecapSource{recap: stored}
	cache := newFakeCache()
	metrics := newFakeRecapCacheMetrics()

	repository := NewCachedRecapRepository(source, cache, nil, metrics)
	recap, err := repository.GetByUserAndYear(ctx, userID, 2026)
	if err != nil {
		t.Fatalf("get recap: %v", err)
	}
	if recap.ID != stored.ID || source.getCalls != 1 {
		t.Fatalf("recap/source calls = %#v/%d, want stored recap and one call", recap, source.getCalls)
	}
	if _, ok := cache.values[recapCacheKey(userID, 2026)]; !ok {
		t.Fatal("cache was not populated after source read")
	}
	if metrics.hits != 0 || metrics.misses != 1 {
		t.Fatalf("cache metrics hits/misses = %d/%d, want 0/1", metrics.hits, metrics.misses)
	}
}

func TestCachedRecapRepositorySaveUpdatesCache(t *testing.T) {
	ctx := context.Background()
	recap := &domain.Recap{ID: uuid.New(), UserID: uuid.New(), Year: 2026}
	source := &fakeRecapSource{}
	cache := newFakeCache()
	repository := NewCachedRecapRepository(source, cache, nil)

	if err := repository.Save(ctx, recap); err != nil {
		t.Fatalf("save recap: %v", err)
	}
	if source.saveCalls != 1 {
		t.Fatalf("source save calls = %d, want 1", source.saveCalls)
	}
	if _, ok := cache.values[recapCacheKey(recap.UserID, recap.Year)]; !ok {
		t.Fatal("cache was not updated after durable save")
	}
}

func TestCachedRecapRepositoryFailsOpenOnCacheError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	stored := &domain.Recap{ID: uuid.New(), UserID: userID, Year: 2026}
	source := &fakeRecapSource{recap: stored}
	cache := newFakeCache()
	cache.getErr = errors.New("redis unavailable")
	cache.setErr = errors.New("redis unavailable")
	metrics := newFakeRecapCacheMetrics()

	repository := NewCachedRecapRepository(source, cache, nil, metrics)
	recap, err := repository.GetByUserAndYear(ctx, userID, 2026)
	if err != nil {
		t.Fatalf("cache failure must not fail read: %v", err)
	}
	if recap.ID != stored.ID || source.getCalls != 1 {
		t.Fatalf("recap/source calls = %#v/%d, want PostgreSQL fallback", recap, source.getCalls)
	}
	if metrics.errors["get"] != 1 || metrics.errors["set"] != 1 {
		t.Fatalf("cache error metrics = %#v, want one get and one set error", metrics.errors)
	}

	source.saveErr = errors.New("postgres unavailable")
	if err := repository.Save(ctx, stored); !errors.Is(err, source.saveErr) {
		t.Fatalf("save error = %v, want PostgreSQL error", err)
	}
}
