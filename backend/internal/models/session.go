package models

import (
	"time"
	"uuid"
)

type Session struct {
	ID        int64
	UserID    int
	Token     uuid.UUID
	ExpiresAt time.Time
}
