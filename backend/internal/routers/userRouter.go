package router

import (
	handler "backend/internal/handlers"
	"net/http"
)

func UserRouter(mux *http.ServeMux) {

	//login and register
	mux.HandleFunc("post api/login", handler.Login)
	mux.HandleFunc("post api/Register", handler.Register)
	// profile
	mux.HandleFunc("GET api/user/{username}", handler.GetUser)

}
