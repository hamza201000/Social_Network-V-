package config

import (
	handler "backend/internal/handlers"
	router "backend/internal/routers"
	"net/http"
)

func Config(userHandler *handler.UserHandler) {
	mux := router.Routers(userHandler)
	http.ListenAndServe(":8080", mux)
}
