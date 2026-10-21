package config

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	router "backend/internal/routers"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func BackendRoot() (string, error) {
	if root := os.Getenv("BACKEND_ROOT"); root != "" {
		return filepath.Abs(root)
	}
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", os.ErrNotExist
	}
	return filepath.Abs(filepath.Join(filepath.Dir(sourceFile), "..", ".."))
}

func Config(userHandler *handler.UserHandler, chatHandler *handler.ChatHandler) error {
	mux := router.Routers(userHandler, chatHandler)
	handler := middleware.CorsMiddleware(mux)
	return http.ListenAndServe(":8080", handler)
}
