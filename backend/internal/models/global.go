package models

type Res struct {
	Message string `json:"message"`
	Err     string `json:"err"`
	Data    any    `json:"data"`
}
