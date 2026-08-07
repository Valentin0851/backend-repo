package domain

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID
	Login     string
	CreatedAt time.Time
}
