package models

import (
	"time"
	"uuid"
)

type Session struct {
	ID        int64
	User_id    int64
	Username  string
	Token     uuid.UUID
	ExpiresAt time.Time
}
