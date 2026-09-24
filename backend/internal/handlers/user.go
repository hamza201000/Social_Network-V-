package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	"encoding/json"
	"fmt"

	"net/http"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

func (userHandler UserHandler) Login(w http.ResponseWriter, r *http.Request) {

	var dataLogin models.LoginRequest
	var res models.Res
	err := json.NewDecoder(r.Body).Decode(&dataLogin)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
	}
	user, err := userHandler.userRepo.LoginUser(dataLogin)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
	}
	fmt.Println(user)
	token, err := userHandler.userRepo.CreateSession(user.Id)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
	}
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    token.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 ,
	}
	http.SetCookie(w, cookie)
	SendJSON(w, http.StatusOK, res)
}

func (repo *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var userData models.RegisterRequest
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

	err := r.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}
	userData.FirstName = r.FormValue("first_name")
	userData.LastName = r.FormValue("last_name")
	userData.Password = r.FormValue("password")
	userData.Username = r.FormValue("username")
	userData.Email = r.FormValue("email")

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
	err = repo.userRepo.RegisterUser(userData)
	if err != nil {
		http.Error(w, "internal server", http.StatusInternalServerError)
	}

}

func (userHandler UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {

}
