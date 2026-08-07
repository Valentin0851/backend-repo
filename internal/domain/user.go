package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is a profile owned by an authenticated account. OwnAds and Views are
// the source data; every aggregate returned by the API is calculated from them.
type User struct {
	ID           uuid.UUID
	Name         string
	Avatar       string
	RegisteredAt time.Time
	ProfileType  string
	Likes        int
	ChatsCount   int
	OwnAds       []OwnAd
	Views        []ViewedAd
}

type Ad struct {
	AdID        string `json:"adId"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory,omitempty"`
	ImageURL    string `json:"imageUrl,omitempty"`
	Price       int64  `json:"price"`
	ViewCount   int    `json:"viewCount"`
}

type ViewedAdEventType string

const (
	ViewedAdEventWatch ViewedAdEventType = "watch"
	ViewedAdEventLike  ViewedAdEventType = "like"
	ViewedAdEventBuy   ViewedAdEventType = "buy"
)

type ViewedAdEvent struct {
	Type             ViewedAdEventType `json:"type"`
	Time             time.Time         `json:"time"`
	UseAvitoDelivery *bool             `json:"useAvitoDelivery,omitempty"`
}

type Review struct {
	Comment   string    `json:"comment"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"createdAt"`
}

type OwnAd struct {
	Ad
	PublishedAt    time.Time  `json:"publishedAt"`
	FavoritesCount int        `json:"favoritesCount"`
	ContactsCount  int        `json:"contactsCount"`
	City           string     `json:"city,omitempty"`
	IsArchived     bool       `json:"isArchived"`
	IsSold         bool       `json:"isSold"`
	SoldAt         *time.Time `json:"soldAt,omitempty"`
	Review         *Review    `json:"review,omitempty"`
}

type ViewedAd struct {
	Ad
	ViewedAt     []ViewedAdEvent `json:"viewedAt"`
	LastViewedAt time.Time       `json:"lastViewedAt"`
	IsFavorite   bool            `json:"isFavorite"`
	FavoritedAt  *time.Time      `json:"favoritedAt,omitempty"`
	IsPurchased  bool            `json:"isPurchased"`
	PurchasedAt  *time.Time      `json:"purchasedAt,omitempty"`
}
