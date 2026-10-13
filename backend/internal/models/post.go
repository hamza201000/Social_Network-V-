package models

import "time"

type PostData struct {
	Id         int64     `json:"id"`
	User_id    int64     `json:"user_id"`
	Username   string    `json:"username"`
	Content    string    `json:"content"`
	Image      string    `json:"image"`
	Privacy    string    `json:"privacy"`
	Group_id   int64      `json:"group_id"`
	Created_at time.Time `json:"created_at"`
}

type PostRequest struct {
	Content     string
	Privacy     string
	FollowerIDs []int64
	Image       string
}
