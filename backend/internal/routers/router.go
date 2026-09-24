package router

import (
	handler "backend/internal/handlers"
	"net/http"
)

func Routers(userHandler *handler.UserHandler) *http.ServeMux {

	mux := http.NewServeMux()
	PostRouter(mux, userHandler)
	UserRouter(mux, userHandler)
	return mux
}
