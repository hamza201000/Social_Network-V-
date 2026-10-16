package config

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	router "backend/internal/routers"
	"net/http"
)

func Config(userHandler *handler.UserHandler, chatHandler *handler.ChatHandler) error {
	mux := router.Routers(userHandler, chatHandler)
	handler := middleware.CorsMiddleware(mux)
	return http.ListenAndServe(":8080", handler)
}
