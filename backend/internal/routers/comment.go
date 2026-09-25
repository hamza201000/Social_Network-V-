package router

import (
	handler "backend/internal/handlers"
	"net/http"
)

func CommentRouter(mux *http.ServeMux, userHandler *handler.UserHandler) {

	//COMMMENT
	mux.HandleFunc("GET api/posts/{id}/comments", userHandler.GetComments)
	mux.HandleFunc("POST api/posts/{id}/comments", userHandler.CreateComment)

}
