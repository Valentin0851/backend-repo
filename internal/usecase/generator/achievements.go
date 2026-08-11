package generator

import (
	"github.com/Valentin0851/avito-recap-backend/internal/domain"
)

// AssignAchievements назначает ачивки на основе метрик
func AssignAchievements(metrics *UserMetrics) []domain.Achievement {
	result := make([]domain.Achievement, 0)

	for _, ach := range domain.DefaultAchievements {
		if checkCondition(ach, metrics) {
			result = append(result, ach)
		}
	}

	return result
}

func checkCondition(ach domain.Achievement, metrics *UserMetrics) bool {
	switch ach.Slug {
	case "curious":
		return metrics.TotalViews >= 500
	case "explorer":
		return metrics.TotalViews >= 1000
	case "social_butterfly":
		return metrics.TotalMessages >= 50
	case "seller_master":
		return metrics.TotalSales >= 5
	case "shopaholic":
		return metrics.TotalPurchases >= 10
	case "veteran":
		return metrics.ActivityDays >= 300
	case "enthusiast":
		return metrics.ActivityDays >= 100
	default:
		return false
	}
}
