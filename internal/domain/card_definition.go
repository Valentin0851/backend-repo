package domain

import (
	"time"

	"github.com/google/uuid"
)

type CardDefinitionKind string
type CardMetric string
type CardAnalysis string
type CardConditionOperator string

const (
	CardKindStatistic CardDefinitionKind = "statistic"
	CardKindHighlight CardDefinitionKind = "highlight"

	CardMetricTotalViews   CardMetric = "total_views"
	CardMetricFavorites    CardMetric = "favorites"
	CardMetricChats        CardMetric = "chats"
	CardMetricPurchases    CardMetric = "purchases"
	CardMetricSales        CardMetric = "sales"
	CardMetricListingViews CardMetric = "listing_views"
	CardMetricContacts     CardMetric = "contacts"
	CardMetricReviews      CardMetric = "reviews"
	CardMetricActivityDays CardMetric = "activity_days"
	CardMetricCategories   CardMetric = "categories"
	CardMetricDeals        CardMetric = "deals"

	CardAnalysisTotal          CardAnalysis = "total"
	CardAnalysisMonthlyAverage CardAnalysis = "monthly_average"
	CardAnalysisMonthlyMax     CardAnalysis = "monthly_max"

	CardConditionAlways CardConditionOperator = "always"
	CardConditionGT     CardConditionOperator = "gt"
	CardConditionGTE    CardConditionOperator = "gte"
	CardConditionLT     CardConditionOperator = "lt"
	CardConditionLTE    CardConditionOperator = "lte"
	CardConditionEQ     CardConditionOperator = "eq"
)

type CardDefinition struct {
	ID                uuid.UUID             `json:"id"`
	Name              string                `json:"name"`
	TargetUserID      *uuid.UUID            `json:"target_user_id,omitempty"`
	Kind              CardDefinitionKind    `json:"kind"`
	Metric            CardMetric            `json:"metric"`
	Analysis          CardAnalysis          `json:"analysis"`
	ConditionOperator CardConditionOperator `json:"condition_operator"`
	ConditionValue    *float64              `json:"condition_value,omitempty"`
	Title             string                `json:"title"`
	Description       string                `json:"description"`
	ValueSuffix       string                `json:"value_suffix"`
	Layout            string                `json:"layout"`
	Theme             string                `json:"theme"`
	Icon              string                `json:"icon"`
	Shareable         bool                  `json:"shareable"`
	SortOrder         int                   `json:"sort_order"`
	IsActive          bool                  `json:"is_active"`
	CreatedBy         uuid.UUID             `json:"created_by"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}
