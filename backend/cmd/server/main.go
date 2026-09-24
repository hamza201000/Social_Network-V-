package main

import (
	"backend/database"
	"backend/internal/config"
	handler "backend/internal/handlers"
	"backend/internal/repository"
)

func main() {

	db := database.ConfigDb()
	UserRepo := repository.NewUserRepository(db)
	userHandler := handler.NewUserHandler(UserRepo)
	config.Config(userHandler)
	defer db.Close()
}
