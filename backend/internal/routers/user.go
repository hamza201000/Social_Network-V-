package router

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	"net/http"
)

func UserRouter(mux *http.ServeMux, userHandler *handler.UserHandler) {

	//login and register
	mux.HandleFunc("POST /api/login", userHandler.Login)
	mux.HandleFunc("POST /api/register", userHandler.Register)
	// profile
	mux.HandleFunc("GET /api/user/{username}", middleware.Middleware(userHandler.UserRepo, userHandler.GetUserProfile))
	mux.HandleFunc("GET /api/myprofile", middleware.Middleware(userHandler.UserRepo, userHandler.GetMyProfile))
	mux.HandleFunc("GET /api/logout", userHandler.Logout)
	//user/${username}/follow
	mux.HandleFunc("GET /api/user/follow", middleware.Middleware(userHandler.UserRepo, userHandler.FollowUser))
	mux.HandleFunc("GET /api/user/{user_id}/followers", middleware.Middleware(userHandler.UserRepo, userHandler.GetFollowers))
	mux.HandleFunc("GET /api/user/{user_id}/following", middleware.Middleware(userHandler.UserRepo, userHandler.GetFollowing))
}
