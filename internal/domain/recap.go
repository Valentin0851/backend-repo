package domain

import (
	"time"

	"github.com/google/uuid"
)

type CategoryStat struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// RecapItem is a compact listing representation that is safe to persist with
// a generated recap. It intentionally contains no seller, buyer or chat data.
type RecapItem struct {
	AdID        string `json:"ad_id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Price       int64  `json:"price"`
}

type BuyerRecapSummary struct {
	HasData                bool       `json:"has_data"`
	ViewedAdsCount         int        `json:"viewed_ads_count"`
	TotalViews             int        `json:"total_views"`
	FavoritesCount         int        `json:"favorites_count"`
	ChatsCount             int        `json:"chats_count"`
	PurchasesCount         int        `json:"purchases_count"`
	AvitoDeliveryPurchases int        `json:"avito_delivery_purchases"`
	MainCategory           string     `json:"main_category,omitempty"`
	LargestPurchase        *RecapItem `json:"largest_purchase,omitempty"`
}

type SellerRecapSummary struct {
	HasData           bool       `json:"has_data"`
	ListingsCount     int        `json:"listings_count"`
	ActiveListings    int        `json:"active_listings"`
	ArchivedListings  int        `json:"archived_listings"`
	SalesCount        int        `json:"sales_count"`
	ListingViews      int        `json:"listing_views"`
	LikesReceived     int        `json:"likes_received"`
	FavoritesReceived int        `json:"favorites_received"`
	ContactsReceived  int        `json:"contacts_received"`
	ReviewsCount      int        `json:"reviews_count"`
	AverageRating     *float64   `json:"average_rating,omitempty"`
	MainCategory      string     `json:"main_category,omitempty"`
	StarListing       *RecapItem `json:"star_listing,omitempty"`
}

type CombinedRecapSummary struct {
	HasBuyerData  bool   `json:"has_buyer_data"`
	HasSellerData bool   `json:"has_seller_data"`
	Categories    int    `json:"categories_count"`
	Deals         int    `json:"deals_count"`
	MainCategory  string `json:"main_category,omitempty"`
}

type RecapSummary struct {
	Headline    string               `json:"headline"`
	Description string               `json:"description"`
	Buyer       BuyerRecapSummary    `json:"buyer"`
	Seller      SellerRecapSummary   `json:"seller"`
	Combined    CombinedRecapSummary `json:"combined"`
}

type RecapChartSeries struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Color  string `json:"color"`
	Values []int  `json:"values"`
}

type RecapChartSegment struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Color string `json:"color"`
	Value int    `json:"value"`
}

type RecapChartHighlight struct {
	Index int    `json:"index"`
	Label string `json:"label"`
	Value int    `json:"value"`
}

// RecapVisualization is a versioned, backend-driven chart specification.
// Cartesian charts use labels and series; part-to-whole charts use segments.
type RecapVisualization struct {
	Version   int                  `json:"version"`
	Type      string               `json:"type"`
	Unit      string               `json:"unit,omitempty"`
	Stacked   bool                 `json:"stacked,omitempty"`
	Labels    []string             `json:"labels,omitempty"`
	Series    []RecapChartSeries   `json:"series,omitempty"`
	Segments  []RecapChartSegment  `json:"segments,omitempty"`
	Highlight *RecapChartHighlight `json:"highlight,omitempty"`
}

type RecapCardPresentation struct {
	Layout string `json:"layout"`
	Theme  string `json:"theme"`
	Icon   string `json:"icon"`
}

type RecapCardCTA struct {
	Label  string            `json:"label"`
	Action string            `json:"action"`
	Params map[string]string `json:"params,omitempty"`
}

// RecapCard contains ready-to-render copy and presentation hints. The frontend
// only needs to render the ordered slice and execute the semantic CTA action.
type RecapCard struct {
	ID            string                `json:"id"`
	Kind          string                `json:"kind"`
	Eyebrow       string                `json:"eyebrow,omitempty"`
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	Value         string                `json:"value,omitempty"`
	ImageURL      string                `json:"image_url,omitempty"`
	Shareable     bool                  `json:"shareable"`
	Reason        string                `json:"reason"`
	Presentation  RecapCardPresentation `json:"presentation"`
	Visualization *RecapVisualization   `json:"visualization,omitempty"`
	CTA           *RecapCardCTA         `json:"cta,omitempty"`
}

type Recap struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Year           int
	TotalViews     int
	TotalMessages  int
	TotalFavorites int
	TotalPurchases int
	TotalSales     int
	TopCategories  []CategoryStat
	Achievements   []Achievement
	ActivityDays   int
	Summary        RecapSummary
	Cards          []RecapCard
	GeneratedAt    time.Time
}
