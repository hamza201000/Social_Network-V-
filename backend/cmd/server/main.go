package main

import (
	"backend/database"
	"backend/internal/config"
	handler "backend/internal/handlers"
	"backend/internal/repository"
	chatws "backend/internal/websocket"
	"log"
)

func main() {
	db := database.ConfigDb()
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	userHandler := handler.NewUserHandler(userRepo)
	messageRepo := repository.NewMessageRepository(db)
	chatHandler := handler.NewChatHandler(messageRepo, chatws.NewHub())

	if err := config.Config(userHandler, chatHandler); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
