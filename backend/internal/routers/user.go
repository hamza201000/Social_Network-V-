package router

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	"net/http"
)

func UserRouter(mux *http.ServeMux, userHandler *handler.UserHandler) {

	//login and register
	mux.HandleFunc("POST /api/login", middleware.Middleware(userHandler.UserRepo, userHandler.Login))
	mux.HandleFunc("POST /api/register", middleware.Middleware(userHandler.UserRepo, userHandler.Register))
	// profile  
	mux.HandleFunc("GET /api/user/{username}", middleware.Middleware(userHandler.UserRepo, userHandler.GetUserProfile))
	mux.HandleFunc("GET /api/logout", userHandler.Logout)
}
