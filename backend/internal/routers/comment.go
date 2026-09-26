package router

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	"net/http"
)

func CommentRouter(mux *http.ServeMux, userHandler *handler.UserHandler) {

	//COMMMENT
	mux.HandleFunc("GET api/posts/{id}/comments", middleware.Middleware(userHandler.UserRepo, userHandler.GetComments))
	mux.HandleFunc("POST api/posts/{id}/comments", middleware.Middleware(userHandler.UserRepo, userHandler.CreateComment))

}
