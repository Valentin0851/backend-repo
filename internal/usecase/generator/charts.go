package generator

import (
	"fmt"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
)

var monthLabels = []string{
	"Янв", "Фев", "Мар", "Апр", "Май", "Июн",
	"Июл", "Авг", "Сен", "Окт", "Ноя", "Дек",
}

var monthNames = []string{
	"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
	"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
}

func buildCategoryMixCard(metrics ProfileMetrics) *domain.RecapCard {
	if len(metrics.CategoryStats) < 2 {
		return nil
	}

	const namedCategoryLimit = 4
	segmentColors := []string{"avito-blue", "avito-purple", "avito-green", "avito-orange"}
	segmentCount := min(len(metrics.CategoryStats), namedCategoryLimit)
	segments := make([]domain.RecapChartSegment, 0, segmentCount+1)
	total := 0
	for _, category := range metrics.CategoryStats {
		total += category.Count
	}
	if total == 0 {
		return nil
	}

	for index := 0; index < segmentCount; index++ {
		category := metrics.CategoryStats[index]
		segments = append(segments, domain.RecapChartSegment{
			Key:   fmt.Sprintf("category_%d", index+1),
			Label: category.Category,
			Color: segmentColors[index],
			Value: category.Count,
		})
	}
	if len(metrics.CategoryStats) > namedCategoryLimit {
		other := 0
		for _, category := range metrics.CategoryStats[namedCategoryLimit:] {
			other += category.Count
		}
		segments = append(segments, domain.RecapChartSegment{
			Key:   "other",
			Label: "Другое",
			Color: "avito-red",
			Value: other,
		})
	}

	leader := metrics.CategoryStats[0]
	share := leader.Count * 100 / total
	return &domain.RecapCard{
		ID:          "category_mix",
		Kind:        "chart",
		Eyebrow:     fmt.Sprintf("Интересы %d года", metrics.Year),
		Title:       leader.Category + " — ваш главный интерес",
		Description: fmt.Sprintf("На эту категорию пришлось %d%% вашей активности на Авито.", share),
		Value:       fmt.Sprintf("%d%%", share),
		Shareable:   true,
		Reason:      "Категории отсортированы по покупкам, избранному, продажам, просмотрам и активности объявлений.",
		Presentation: domain.RecapCardPresentation{
			Layout: "chart",
			Theme:  "avito-purple",
			Icon:   "categories",
		},
		Visualization: &domain.RecapVisualization{
			Version:  1,
			Type:     "donut",
			Segments: segments,
			Highlight: &domain.RecapChartHighlight{
				Index: 0,
				Label: leader.Category,
				Value: leader.Count,
			},
		},
		CTA: categoryCTA(leader.Category),
	}
}

func buildActivityRhythmCard(metrics ProfileMetrics) *domain.RecapCard {
	views := make([]int, 12)
	favorites := make([]int, 12)
	listings := make([]int, 12)
	deals := make([]int, 12)

	peakIndex := 0
	peakValue := 0
	for index, activity := range metrics.Monthly {
		views[index] = activity.Views
		favorites[index] = activity.Favorites
		listings[index] = activity.Listings
		deals[index] = activity.Purchases + activity.Sales

		monthTotal := views[index] + favorites[index] + listings[index] + deals[index]
		if monthTotal > peakValue {
			peakIndex = index
			peakValue = monthTotal
		}
	}
	if peakValue == 0 {
		return nil
	}

	series := make([]domain.RecapChartSeries, 0, 4)
	series = appendSeriesIfNotEmpty(series, "views", "Просмотры", "avito-blue", views)
	series = appendSeriesIfNotEmpty(series, "favorites", "Избранное", "avito-red", favorites)
	series = appendSeriesIfNotEmpty(series, "listings", "Объявления", "avito-purple", listings)
	series = appendSeriesIfNotEmpty(series, "deals", "Сделки", "avito-green", deals)

	return &domain.RecapCard{
		ID:          "activity_rhythm",
		Kind:        "chart",
		Eyebrow:     fmt.Sprintf("Ритм %d года", metrics.Year),
		Title:       monthNames[peakIndex] + " стал самым активным месяцем",
		Description: fmt.Sprintf("В этом месяце было %s — больше, чем в любом другом.", formatActionCount(peakValue)),
		Value:       monthNames[peakIndex],
		Shareable:   true,
		Reason:      "Датированные просмотры, избранное, публикации и сделки сгруппированы по месяцам выбранного года.",
		Presentation: domain.RecapCardPresentation{
			Layout: "chart",
			Theme:  "avito-blue",
			Icon:   "calendar",
		},
		Visualization: &domain.RecapVisualization{
			Version: 1,
			Type:    "bar",
			Stacked: true,
			Labels:  append([]string(nil), monthLabels...),
			Series:  series,
			Highlight: &domain.RecapChartHighlight{
				Index: peakIndex,
				Label: monthNames[peakIndex],
				Value: peakValue,
			},
		},
	}
}

func appendSeriesIfNotEmpty(
	series []domain.RecapChartSeries,
	key string,
	label string,
	color string,
	values []int,
) []domain.RecapChartSeries {
	for _, value := range values {
		if value > 0 {
			return append(series, domain.RecapChartSeries{
				Key:    key,
				Label:  label,
				Color:  color,
				Values: values,
			})
		}
	}
	return series
}

func formatActionCount(value int) string {
	form := "действий"
	lastTwo := value % 100
	last := value % 10
	if lastTwo < 11 || lastTwo > 14 {
		switch last {
		case 1:
			form = "действие"
		case 2, 3, 4:
			form = "действия"
		}
	}
	return fmt.Sprintf("%d %s", value, form)
}
