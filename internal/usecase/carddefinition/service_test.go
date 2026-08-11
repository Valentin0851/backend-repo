package carddefinition

import (
	"context"
	"errors"
	"testing"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	apperrors "github.com/Valentin0851/avito-recap-backend/pkg/errors"
	"github.com/google/uuid"
)

type fakeRepository struct {
	created *domain.CardDefinition
}

func (f *fakeRepository) Create(_ context.Context, definition *domain.CardDefinition) error {
	copy := *definition
	f.created = &copy
	return nil
}

func (f *fakeRepository) List(context.Context) ([]domain.CardDefinition, error) {
	return nil, nil
}

func (f *fakeRepository) Update(context.Context, *domain.CardDefinition) error {
	return nil
}

func (f *fakeRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (f *fakeRepository) ListActiveForUser(context.Context, uuid.UUID) ([]domain.CardDefinition, error) {
	return nil, nil
}

func TestCreateValidCardDefinition(t *testing.T) {
	repository := &fakeRepository{}
	service := New(repository)
	adminID := uuid.New()
	definition := validDefinition()

	created, err := service.Create(context.Background(), adminID, &definition)
	if err != nil {
		t.Fatalf("create definition: %v", err)
	}
	if created.ID == uuid.Nil || created.CreatedBy != adminID {
		t.Fatalf("metadata was not assigned: %#v", created)
	}
	if repository.created == nil || repository.created.ID != created.ID {
		t.Fatal("definition was not persisted")
	}
}

func TestCreateRejectsMonthlyAnalysisForSnapshotMetric(t *testing.T) {
	service := New(&fakeRepository{})
	definition := validDefinition()
	definition.Metric = domain.CardMetricChats
	definition.Analysis = domain.CardAnalysisMonthlyAverage

	_, err := service.Create(context.Background(), uuid.New(), &definition)
	if !errors.Is(err, apperrors.ErrInvalidCardDefinition) {
		t.Fatalf("error = %v, want invalid card definition", err)
	}
}

func validDefinition() domain.CardDefinition {
	return domain.CardDefinition{
		Name:              "Просмотры",
		Kind:              domain.CardKindStatistic,
		Metric:            domain.CardMetricTotalViews,
		Analysis:          domain.CardAnalysisTotal,
		ConditionOperator: domain.CardConditionAlways,
		Title:             "Ваши просмотры",
		Layout:            "statistic",
		Theme:             "avito-purple",
		Icon:              "eye",
		Shareable:         true,
		SortOrder:         100,
		IsActive:          true,
	}
}
