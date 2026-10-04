package router

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	"net/http"
)

func Routers(userHandler *handler.UserHandler) *http.ServeMux {

	mux := http.NewServeMux()
	PostRouter(mux, userHandler)
	CommentRouter(mux, userHandler)
	UserRouter(mux, userHandler)
	GroupRouter(mux, userHandler)
	middleware.Logging(mux)
	return mux
}
