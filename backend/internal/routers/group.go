package router

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	"net/http"
)

func GroupRouter(mux *http.ServeMux, userHandler *handler.UserHandler) {

	//group

	mux.HandleFunc("POST /api/group", middleware.Middleware(userHandler.UserRepo, userHandler.CreateGroup))
	mux.HandleFunc("GET /api/group/{group_id}/invite", middleware.Middleware(userHandler.UserRepo, userHandler.InviteUsrGp))
	mux.HandleFunc("GET /api/group/{group_id}/join", middleware.Middleware(userHandler.UserRepo, userHandler.JoinUsrGroup))
	mux.HandleFunc("GET /api/group/search", middleware.Middleware(userHandler.UserRepo, userHandler.SearchGroup))
}
