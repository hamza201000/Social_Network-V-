package router

import (
	handler "backend/internal/handlers"
	"backend/internal/middleware"
	"net/http"
)

func PostRouter(mux *http.ServeMux, userHandler *handler.UserHandler) {

	//POST
	mux.HandleFunc("POST /api/posts", middleware.Middleware(userHandler.UserRepo, userHandler.CreatePost))
	mux.HandleFunc("GET /api/posts", middleware.Middleware(userHandler.UserRepo, userHandler.GetPosts))
	mux.HandleFunc("GET /api/posts/{id}", middleware.Middleware(userHandler.UserRepo, userHandler.GetPostById))

}
