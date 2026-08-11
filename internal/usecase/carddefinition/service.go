package carddefinition

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	"github.com/Valentin0851/avito-recap-backend/internal/usecase/ports"
	apperrors "github.com/Valentin0851/avito-recap-backend/pkg/errors"
	"github.com/google/uuid"
)

type Service struct {
	repository ports.CardDefinitionRepository
}

func New(repository ports.CardDefinitionRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(
	ctx context.Context,
	adminID uuid.UUID,
	definition *domain.CardDefinition,
) (*domain.CardDefinition, error) {
	if definition == nil {
		return nil, invalid("request body is required")
	}
	normalize(definition)
	if err := validate(definition); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	definition.ID = uuid.New()
	definition.CreatedBy = adminID
	definition.CreatedAt = now
	definition.UpdatedAt = now
	if err := s.repository.Create(ctx, definition); err != nil {
		return nil, fmt.Errorf("create card definition: %w", err)
	}
	return definition, nil
}

func (s *Service) List(ctx context.Context) ([]domain.CardDefinition, error) {
	definitions, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list card definitions: %w", err)
	}
	return definitions, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete card definition: %w", err)
	}
	return nil
}

func normalize(definition *domain.CardDefinition) {
	definition.Name = strings.TrimSpace(definition.Name)
	definition.Title = strings.TrimSpace(definition.Title)
	definition.Description = strings.TrimSpace(definition.Description)
	definition.ValueSuffix = strings.TrimSpace(definition.ValueSuffix)
	definition.Layout = strings.TrimSpace(definition.Layout)
	definition.Theme = strings.TrimSpace(definition.Theme)
	definition.Icon = strings.TrimSpace(definition.Icon)
}

func validate(definition *domain.CardDefinition) error {
	if !validLength(definition.Name, 1, 100) {
		return invalid("name must contain from 1 to 100 characters")
	}
	if definition.TargetUserID != nil && *definition.TargetUserID == uuid.Nil {
		return invalid("target_user_id must be a non-zero UUID")
	}
	if !oneOf(definition.Kind, domain.CardKindStatistic, domain.CardKindHighlight) {
		return invalid("kind must be statistic or highlight")
	}
	if !validMetric(definition.Metric) {
		return invalid("unsupported metric")
	}
	if !oneOf(
		definition.Analysis,
		domain.CardAnalysisTotal,
		domain.CardAnalysisMonthlyAverage,
		domain.CardAnalysisMonthlyMax,
	) {
		return invalid("analysis must be total, monthly_average or monthly_max")
	}
	if definition.Analysis != domain.CardAnalysisTotal && !supportsMonthlyAnalysis(definition.Metric) {
		return invalid("monthly analysis is available only for total_views, favorites, purchases and sales")
	}
	if !validCondition(definition.ConditionOperator) {
		return invalid("unsupported condition_operator")
	}
	if definition.ConditionOperator == domain.CardConditionAlways && definition.ConditionValue != nil {
		return invalid("condition_value must be omitted when condition_operator is always")
	}
	if definition.ConditionOperator != domain.CardConditionAlways && definition.ConditionValue == nil {
		return invalid("condition_value is required for the selected condition_operator")
	}
	if !validLength(definition.Title, 1, 160) {
		return invalid("title must contain from 1 to 160 characters")
	}
	if utf8.RuneCountInString(definition.Description) > 500 {
		return invalid("description must not exceed 500 characters")
	}
	if utf8.RuneCountInString(definition.ValueSuffix) > 40 {
		return invalid("value_suffix must not exceed 40 characters")
	}
	if definition.Layout != "statistic" && definition.Layout != "hero" {
		return invalid("layout must be statistic or hero")
	}
	if !validLength(definition.Theme, 1, 50) {
		return invalid("theme must contain from 1 to 50 characters")
	}
	if !validLength(definition.Icon, 1, 50) {
		return invalid("icon must contain from 1 to 50 characters")
	}
	if definition.SortOrder < 0 {
		return invalid("sort_order must not be negative")
	}
	return nil
}

func validMetric(metric domain.CardMetric) bool {
	return oneOf(
		metric,
		domain.CardMetricTotalViews,
		domain.CardMetricFavorites,
		domain.CardMetricChats,
		domain.CardMetricPurchases,
		domain.CardMetricSales,
		domain.CardMetricListingViews,
		domain.CardMetricContacts,
		domain.CardMetricReviews,
		domain.CardMetricActivityDays,
		domain.CardMetricCategories,
		domain.CardMetricDeals,
	)
}

func supportsMonthlyAnalysis(metric domain.CardMetric) bool {
	return oneOf(
		metric,
		domain.CardMetricTotalViews,
		domain.CardMetricFavorites,
		domain.CardMetricPurchases,
		domain.CardMetricSales,
	)
}

func validCondition(operator domain.CardConditionOperator) bool {
	return oneOf(
		operator,
		domain.CardConditionAlways,
		domain.CardConditionGT,
		domain.CardConditionGTE,
		domain.CardConditionLT,
		domain.CardConditionLTE,
		domain.CardConditionEQ,
	)
}

func oneOf[T comparable](value T, values ...T) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func validLength(value string, minimum, maximum int) bool {
	length := utf8.RuneCountInString(value)
	return length >= minimum && length <= maximum
}

func invalid(message string) error {
	return fmt.Errorf("%w: %s", apperrors.ErrInvalidCardDefinition, message)
}
