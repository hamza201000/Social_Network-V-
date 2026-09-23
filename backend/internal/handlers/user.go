package handler

import (
	"backend/internal/models"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {

}

func Register(w http.ResponseWriter, r *http.Request) {
	var userData models.RegisterRequest
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	err := r.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}
	userData.FirstName = r.FormValue("first_name")
	userData.LastName = r.FormValue("last_name")
	userData.Password= r.FormValue("password")
	userData.Username = r.FormValue("username")
	userData.Email = r.FormValue("email")
	userData.Password= r.FormValue("password")

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > 5*1024*1024 {
		http.Error(w, "image must be less than 5 MB", http.StatusBadRequest)
		return
	}
	// err=RegisterUser()
	// if (err!=nil){
	// 	http.Error(w, "internal server", http.StatusInternalServerError)
	// }

}

func GetUser(w http.ResponseWriter, r *http.Request) {

}
