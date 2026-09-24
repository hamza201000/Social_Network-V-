package models

import (
	"time"
	"uuid"
)

type Session struct {
	ID     uuid.UUID
	UserID int
	Token int64
	ExpiresAt time.Time
}
