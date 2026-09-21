package config

import (
	router "backend/internal/routers"
	"net/http"
)


func Config(){

	mux:=router.Routers()

	
	http.ListenAndServe(":8080",mux)

}