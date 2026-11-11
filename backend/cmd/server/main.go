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
	// one hub shared by chat and notifications: a user has one set of websocket connections
	hub := chatws.NewHub()
	userHandler := handler.NewUserHandler(userRepo, imageStorage, hub)
	messageRepo := repository.NewMessageRepository(db)
	chatHandler := handler.NewChatHandler(messageRepo, hub)

	if err := config.Config(userHandler, chatHandler); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
