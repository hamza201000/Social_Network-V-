package router

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	"net/http"
)

func Routers(userHandler *handler.UserHandler, chatHandler *handler.ChatHandler) *http.ServeMux {

	mux := http.NewServeMux()
	PostRouter(mux, userHandler)
	CommentRouter(mux, userHandler)
	UserRouter(mux, userHandler)
	GroupRouter(mux, userHandler)
	ChatRouter(mux, userHandler.UserRepo, chatHandler)
	middleware.Logging(mux)
	return mux
}
