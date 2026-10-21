package models

import "time"

type DataMessage struct {
	ID             int64     `json:"id"`
	Type           string    `json:"type"`
	UsernameSender string    `json:"username_sender"`
	SenderID       int64     `json:"sender_id"`
	ReceiverID     int64     `json:"receiver_id"`
	Message        string    `json:"message"`
	Seen           bool      `json:"seen"`
	CreatedAt      time.Time `json:"created_at"`
}

type Client struct {
	UserID   int64        `json:"user_id"`
	Username string       `json:"username"`
	Avatar_url string   	`json:"avatar_url"`
	IsOnline bool         `json:"is_online"`
	LastMsg  *DataMessage `json:"last_msg"`
}

type NewMessage struct {
	ReceiverID int64  `json:"receiver_id"`
	Message    string `json:"message"`
}
