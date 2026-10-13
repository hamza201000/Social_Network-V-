package router

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	"net/http"
)

func GroupRouter(mux *http.ServeMux, userHandler *handler.UserHandler) {

	//group

	mux.HandleFunc("POST /api/groups", middleware.Middleware(userHandler.UserRepo, userHandler.CreateGroup))
	mux.HandleFunc("GET /api/groups", middleware.Middleware(userHandler.UserRepo, userHandler.GetGroups))
	mux.HandleFunc("GET /api/groups/{group_id}", middleware.Middleware(userHandler.UserRepo, userHandler.GetGroup))
	mux.HandleFunc("GET /api/groups/{group_id}/invite", middleware.Middleware(userHandler.UserRepo, userHandler.InviteUsrGp))
	mux.HandleFunc("GET /api/groups/{group_id}/join", middleware.Middleware(userHandler.UserRepo, userHandler.JoinUsrGroup))
	mux.HandleFunc("GET /api/groups/search", middleware.Middleware(userHandler.UserRepo, userHandler.SearchGroup))
	mux.HandleFunc("POST /api/groups/{group_id}/posts", middleware.Middleware(userHandler.UserRepo, userHandler.CreatePostGp))
}
