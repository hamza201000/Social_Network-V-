package router

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	"net/http"
)

func NotificationRouter(mux *http.ServeMux, userHandler *handler.UserHandler) {
	mux.HandleFunc("GET /api/notifications", middleware.Middleware(userHandler.UserRepo, userHandler.GetNotifications))
	mux.HandleFunc("PUT /api/notifications/read-all", middleware.Middleware(userHandler.UserRepo, userHandler.MarkAllNotificationsRead))
	mux.HandleFunc("PUT /api/notifications/{id}/read", middleware.Middleware(userHandler.UserRepo, userHandler.MarkNotificationRead))
}
