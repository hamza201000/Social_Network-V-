package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"net/http"
)

type UserHandler struct {
	userRepo    *repository.UserRepository
	postRepo    *repository.PostRepository
	commentRepo *repository.CommentRepository
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

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		10*1024*1024,
	)

	err := r.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		SendError(
			w,
			http.StatusRequestEntityTooLarge,
			"request is too large",
		)
		return
	}

	userData.FirstName = strings.TrimSpace(r.FormValue("first_name"))
	userData.LastName = strings.TrimSpace(r.FormValue("last_name"))
	userData.Username = strings.TrimSpace(r.FormValue("username"))
	userData.Email = strings.TrimSpace(r.FormValue("email"))
	userData.Password = r.FormValue("password")

	err = services.RegisterRequest(userData)
	if err != nil {
		SendError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	file, header, err := r.FormFile("image")

	if err == nil && file != nil {
		defer file.Close()

		// Maximum image size: 5 MB
		if header.Size > 5*1024*1024 {
			SendError(
				w,
				http.StatusBadRequest,
				"image must be less than 5 MB",
			)
			return
		}
		contentType := header.Header.Get("Content-Type")
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/webp": true,
		}
		if !allowedTypes[contentType] {
			SendError(
				w,
				http.StatusBadRequest,
				"only JPEG, PNG and WebP images are allowed",
			)
			return
		}
	}
	err = repo.userRepo.RegisterUser(userData)
	if err != nil {
		fmt.Println("register user error:", err)
		SendError(
			w,
			http.StatusInternalServerError,
			"could not register user",
		)
		return
	}

	res.Message = "user registered successfully"

	w.Header().Set("Content-Type", "application/json")
	SendJSON(w, http.StatusCreated, res)
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
