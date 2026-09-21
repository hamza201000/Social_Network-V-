package router

import "net/http"

func Routers()*http.ServeMux {

	mux:=http.NewServeMux()
	PostRouter(mux)
	
	return  mux
}
