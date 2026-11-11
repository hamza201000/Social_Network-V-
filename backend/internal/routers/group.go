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
	mux.HandleFunc("GET /api/groups/{group_id}/invitable", middleware.Middleware(userHandler.UserRepo, userHandler.GetInvitableUsers))
	mux.HandleFunc("POST /api/groups/{group_id}/invite", middleware.Middleware(userHandler.UserRepo, userHandler.InviteUsrGp))
	mux.HandleFunc("POST /api/groups/{group_id}/join", middleware.Middleware(userHandler.UserRepo, userHandler.JoinUsrGroup))
	mux.HandleFunc("POST /api/groups/{group_id}/invitation", middleware.Middleware(userHandler.UserRepo, userHandler.RespondInvitation))
	mux.HandleFunc("GET /api/groups/{group_id}/members", middleware.Middleware(userHandler.UserRepo, userHandler.GetGroupMembers))
	mux.HandleFunc("GET /api/groups/{group_id}/requests", middleware.Middleware(userHandler.UserRepo, userHandler.GetJoinRequests))
	mux.HandleFunc("POST /api/groups/{group_id}/requests/{user_id}", middleware.Middleware(userHandler.UserRepo, userHandler.RespondJoinRequest))
	mux.HandleFunc("GET /api/groups/search", middleware.Middleware(userHandler.UserRepo, userHandler.SearchGroup))
	mux.HandleFunc("POST /api/groups/{group_id}/posts", middleware.Middleware(userHandler.UserRepo, userHandler.CreatePostGp))
	mux.HandleFunc("GET /api/groups/{group_id}/posts", middleware.Middleware(userHandler.UserRepo, userHandler.GetPostGp))
	mux.HandleFunc("POST /api/groups/{group_id}/events", middleware.Middleware(userHandler.UserRepo, userHandler.CreateEvent))
	mux.HandleFunc("GET /api/groups/{group_id}/events", middleware.Middleware(userHandler.UserRepo, userHandler.GetGroupEvents))
	mux.HandleFunc("POST /api/groups/{group_id}/events/{event_id}/response", middleware.Middleware(userHandler.UserRepo, userHandler.VoteEvent))
}
