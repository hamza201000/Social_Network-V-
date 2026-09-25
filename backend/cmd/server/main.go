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
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	userHandler := handler.NewUserHandler(UserRepo)
	userHandler.SetContentRepositories(postRepo, commentRepo)
	config.Config(userHandler)
	defer db.Close()
}
