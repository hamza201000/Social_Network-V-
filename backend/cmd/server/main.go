package main

import (
	"backend/database"
	"backend/internal/config"
	handler "backend/internal/handlers"
	"backend/internal/repository"
	"backend/internal/services"
	chatws "backend/internal/websocket"
	"log"
)

func main() {
	db := database.ConfigDb()
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	backendRoot, err := config.BackendRoot()
	if err != nil {
		log.Fatalf("could not determine backend root: %v", err)
	}
	imageStorage := services.NewImageStorage(backendRoot)
	userHandler := handler.NewUserHandler(userRepo, imageStorage)
	messageRepo := repository.NewMessageRepository(db)
	chatHandler := handler.NewChatHandler(messageRepo, chatws.NewHub())

	if err := config.Config(userHandler, chatHandler); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
