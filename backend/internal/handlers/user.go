package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	"database/sql"
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
		return
	}
	user, err := userHandler.userRepo.LoginUser(dataLogin)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	fmt.Println(user)
	token, err := userHandler.userRepo.CreateSession(user.Id)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    token.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60,
	}
	http.SetCookie(w, cookie)
	SendJSON(w, http.StatusOK, res)
}

func (repo *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var res models.Res
	var userData models.RegisterRequest
	fmt.Println("register request received")
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

	file, header, _ := r.FormFile("image")
	if file != nil {
		if header.Size > 5*1024*1024 {
			http.Error(w, "image must be less than 5 MB", http.StatusBadRequest)
			return
		}
		defer file.Close()
	}
	err = repo.userRepo.RegisterUser(userData)
	if err != nil {
		http.Error(w, "internal server", http.StatusInternalServerError)
		return
	}
	res.Message = "user registered successfully"
	SendJSON(w, http.StatusOK, res)
}

func (userHandler UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var res models.Res
	cookie, err := r.Cookie("session_id")
	if err != nil {
		SendError(w, http.StatusUnauthorized, "not logged in")
		return
	}

	err = userHandler.userRepo.DeleteSession(cookie.Value)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	res.Message = "logout successful"
	SendJSON(w, http.StatusOK, res)
}

func (userHandler UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	var res models.Res
	cookie, err := r.Cookie("session_id")
	if err != nil {
		SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := userHandler.userRepo.GetUserBySession(cookie.Value)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	res.Data = user
	SendJSON(w, http.StatusOK, res)
}

func (userHandler UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	var res models.Res
	username := r.URL.Query().Get("username")
	fmt.Println(username)
	if username == "" {
		SendError(w, http.StatusBadRequest, "username is required")
		return
	}

	user, err := userHandler.userRepo.GetUserByUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			SendError(w, http.StatusNotFound, "user not found")
			return
		}

		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	res.Data = user
	SendJSON(w, http.StatusOK, res)
}
