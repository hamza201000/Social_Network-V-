package router

import (
	handler "backend/internal/handlers"
	"net/http"
)

func PostRouter(mux *http.ServeMux) {

	//POST
	mux.HandleFunc("GET api/posts",handler.GetPosts)
	mux.HandleFunc("GET api/posts/{id}",handler.GetPostById)
	mux.HandleFunc("POST api/posts",handler.CreatePost)
	//COMMMENT
	mux.HandleFunc("GET api/posts/{id}/comments",handler.GetComments)
	mux.HandleFunc("POST api/posts/{id}/comments",handler.CreateComment)


}
