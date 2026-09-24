package router

import (
	handler "backend/internal/handlers"
	"net/http"
)

func UserRouter(mux *http.ServeMux, userHandler *handler.UserHandler) {

	//login and register
	mux.HandleFunc("post api/login", userHandler.Login)
	mux.HandleFunc("post api/Register", userHandler.Register)
	// profile
	mux.HandleFunc("GET api/user/{username}", userHandler.GetUser)

}
