package router

import (
	handler "backend/internal/handlers"
	"net/http"
)

func PostRouter(mux *http.ServeMux, userHandler *handler.UserHandler) {

	//POST
	mux.HandleFunc("GET api/posts", userHandler.GetPosts)
	mux.HandleFunc("GET api/posts/{id}", userHandler.GetPostById)
	mux.HandleFunc("POST api/posts", userHandler.CreatePost)

}
