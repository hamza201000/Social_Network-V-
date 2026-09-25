package models

import "time"

type CommentData struct {
	Id         int64     `json:"id"`
	User_id    int64     `json:"user_id"`
	Username   string    `json:"username"`
	Post_id    int64     `json:"post_id"`
	Content    string    `json:"content"`
	Image      string    `json:"image"`
	Created_at time.Time `json:"created_at"`
}

type CommentRequest struct {
	Content string
	Image   string
}
