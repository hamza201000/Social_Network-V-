package router

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/repository"
	"net/http"
)

func ChatRouter(mux *http.ServeMux, userRepo *repository.UserRepository, chatHandler *handler.ChatHandler) {
	mux.HandleFunc("GET /api/chats", middleware.Middleware(userRepo, chatHandler.GetChats))
	mux.HandleFunc("GET /api/chats/{userID}/messages", middleware.Middleware(userRepo, chatHandler.GetConversation))
	mux.HandleFunc("POST /api/messages", middleware.Middleware(userRepo, chatHandler.SendMessage))
	mux.HandleFunc("PUT /api/messages/{id}/seen", middleware.Middleware(userRepo, chatHandler.MarkSeen))
	mux.HandleFunc("GET /api/ws", middleware.Middleware(userRepo, chatHandler.WebSocket))
	// group chat: only the admin and members of the group can read or post
	mux.HandleFunc("GET /api/groups/{group_id}/messages", middleware.Middleware(userRepo, chatHandler.GetGroupMessages))
	mux.HandleFunc("POST /api/groups/{group_id}/messages", middleware.Middleware(userRepo, chatHandler.SendGroupMessage))
}
