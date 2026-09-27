package config

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	router "backend/internal/routers"
	"net/http"
)

func Config(userHandler *handler.UserHandler) {
	mux := router.Routers(userHandler)
	handler := middleware.CorsMiddleware(mux)
	http.ListenAndServe(":8080",handler )
}
