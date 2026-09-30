package models

type User struct {
	Id           int64  `json:"user_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Password        string `json:"-"`
	AvatarUrl        string `json:"avatar_url"`
	DateOfBirth    string `json:"date_of_birth"`
	AboutMe        string `json:"about_me"`
	CreatedAt    string `json:"created_at"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarUrl        string `json:"avatar_url"`
	Password  string `json:"password"`
	DateOfBirth    string `json:"date_of_birth"`
	AboutMe        string `json:"about_me"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
