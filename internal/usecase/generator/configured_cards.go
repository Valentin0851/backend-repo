package generator

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
)

func buildConfiguredCards(
	metrics ProfileMetrics,
	definitions []domain.CardDefinition,
) []domain.RecapCard {
	cards := make([]domain.RecapCard, 0, len(definitions))
	for _, definition := range definitions {
		value, ok := analyzeConfiguredMetric(metrics, definition.Metric, definition.Analysis)
		if !ok || !matchesConfiguredCondition(value, definition.ConditionOperator, definition.ConditionValue) {
			continue
		}
		cards = append(cards, domain.RecapCard{
			ID:          "custom_" + definition.ID.String(),
			Kind:        string(definition.Kind),
			Eyebrow:     definition.Name,
			Title:       definition.Title,
			Description: definition.Description,
			Value:       configuredValue(value, definition.ValueSuffix),
			Shareable:   definition.Shareable,
			Reason:      "Карточка создана администратором по настроенному правилу статистики.",
			Presentation: domain.RecapCardPresentation{
				Layout: definition.Layout,
				Theme:  definition.Theme,
				Icon:   definition.Icon,
			},
		})
	}
	return cards
}

func insertConfiguredCards(existing, configured []domain.RecapCard) []domain.RecapCard {
	if len(configured) == 0 {
		return existing
	}
	if len(existing) == 0 {
		if len(configured) > 9 {
			return configured[:9]
		}
		return configured
	}

	// Preserve the existing API guarantee of at most nine cards. Configured
	// cards take priority over lower-ranked built-in candidates, while the
	// mandatory overview and finale remain in the sequence.
	if len(configured) > 7 {
		configured = configured[:7]
	}
	builtInContent := existing[:len(existing)-1]
	maxBuiltIn := 8 - len(configured)
	if len(builtInContent) > maxBuiltIn {
		builtInContent = builtInContent[:maxBuiltIn]
	}
	result := make([]domain.RecapCard, 0, len(builtInContent)+len(configured)+1)
	result = append(result, builtInContent...)
	result = append(result, configured...)
	result = append(result, existing[len(existing)-1])
	return result
}

func analyzeConfiguredMetric(
	metrics ProfileMetrics,
	metric domain.CardMetric,
	analysis domain.CardAnalysis,
) (float64, bool) {
	if analysis == domain.CardAnalysisTotal {
		return configuredMetricTotal(metrics, metric)
	}
	monthly, ok := configuredMonthlyValues(metrics, metric)
	if !ok {
		return 0, false
	}
	if analysis == domain.CardAnalysisMonthlyAverage {
		total := 0
		for _, value := range monthly {
			total += value
		}
		return float64(total) / float64(len(monthly)), true
	}
	if analysis == domain.CardAnalysisMonthlyMax {
		maximum := 0
		for _, value := range monthly {
			if value > maximum {
				maximum = value
			}
		}
		return float64(maximum), true
	}
	return 0, false
}

func configuredMetricTotal(metrics ProfileMetrics, metric domain.CardMetric) (float64, bool) {
	switch metric {
	case domain.CardMetricTotalViews:
		return float64(metrics.Buyer.TotalViews), true
	case domain.CardMetricFavorites:
		return float64(metrics.Buyer.FavoritesCount), true
	case domain.CardMetricChats:
		return float64(metrics.Buyer.ChatsCount), true
	case domain.CardMetricPurchases:
		return float64(metrics.Buyer.PurchasesCount), true
	case domain.CardMetricSales:
		return float64(metrics.Seller.SalesCount), true
	case domain.CardMetricListingViews:
		return float64(metrics.Seller.ListingViews), true
	case domain.CardMetricContacts:
		return float64(metrics.Seller.ContactsReceived), true
	case domain.CardMetricReviews:
		return float64(metrics.Seller.ReviewsCount), true
	case domain.CardMetricActivityDays:
		return float64(metrics.ActivityDays), true
	case domain.CardMetricCategories:
		return float64(metrics.Combined.Categories), true
	case domain.CardMetricDeals:
		return float64(metrics.Combined.Deals), true
	default:
		return 0, false
	}
}

func configuredMonthlyValues(metrics ProfileMetrics, metric domain.CardMetric) ([12]int, bool) {
	var values [12]int
	for month, activity := range metrics.Monthly {
		switch metric {
		case domain.CardMetricTotalViews:
			values[month] = activity.Views
		case domain.CardMetricFavorites:
			values[month] = activity.Favorites
		case domain.CardMetricPurchases:
			values[month] = activity.Purchases
		case domain.CardMetricSales:
			values[month] = activity.Sales
		default:
			return [12]int{}, false
		}
	}
	return values, true
}

func matchesConfiguredCondition(
	value float64,
	operator domain.CardConditionOperator,
	threshold *float64,
) bool {
	if operator == domain.CardConditionAlways {
		return true
	}
	if threshold == nil {
		return false
	}
	switch operator {
	case domain.CardConditionGT:
		return value > *threshold
	case domain.CardConditionGTE:
		return value >= *threshold
	case domain.CardConditionLT:
		return value < *threshold
	case domain.CardConditionLTE:
		return value <= *threshold
	case domain.CardConditionEQ:
		return math.Abs(value-*threshold) < 1e-9
	default:
		return false
	}
}

func configuredValue(value float64, suffix string) string {
	var formatted string
	if math.Abs(value-math.Round(value)) < 1e-9 {
		formatted = strconv.FormatInt(int64(math.Round(value)), 10)
	} else {
		formatted = strconv.FormatFloat(value, 'f', 1, 64)
	}
	if strings.TrimSpace(suffix) == "" {
		return formatted
	}
	return fmt.Sprintf("%s %s", formatted, strings.TrimSpace(suffix))
}
