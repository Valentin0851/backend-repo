package domain

import (
	"time"

	"github.com/google/uuid"
)

type ActionType string

const (
	ActionView     ActionType = "view"
	ActionMessage  ActionType = "message"
	ActionFavorite ActionType = "favorite"
	ActionPurchase ActionType = "purchase"
	ActionSale     ActionType = "sale"
)

type Action struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Type       ActionType
	CategoryID uuid.UUID
	Category   string
	CreatedAt  time.Time
}
