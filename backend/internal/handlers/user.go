package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"net/http"
)

type UserHandler struct {
	UserRepo     *repository.UserRepository
	ImageStorage *services.ImageStorage
}

func NewUserHandler(userRepo *repository.UserRepository, imageStorage *services.ImageStorage) *UserHandler {
	return &UserHandler{
		UserRepo:     userRepo,
		ImageStorage: imageStorage,
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
	fmt.Println(dataLogin)
	user, err := userHandler.UserRepo.LoginUser(dataLogin)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	fmt.Println(user)
	token, err := userHandler.UserRepo.CreateSession(user.Id)
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
		MaxAge:   3800,
	}
	http.SetCookie(w, cookie)
	SendJSON(w, http.StatusOK, res)
}

func (repo *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var res models.Res
	var userData models.RegisterRequest

	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)
	if err := r.ParseMultipartForm(1 * 1024 * 1024); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			SendError(w, http.StatusRequestEntityTooLarge, "request is too large")
			return
		}
		SendError(w, http.StatusBadRequest, "request is too large or invalid")
		return
	}

	userData.FirstName = strings.TrimSpace(r.FormValue("first_name"))
	userData.LastName = strings.TrimSpace(r.FormValue("last_name"))
	userData.Username = strings.TrimSpace(r.FormValue("username"))
	userData.Email = strings.TrimSpace(r.FormValue("email"))
	userData.Password = r.FormValue("password")
	userData.DateOfBirth = strings.TrimSpace(r.FormValue("date_of_birth"))
	userData.AboutMe = strings.TrimSpace(r.FormValue("about_me"))

	err := services.RegisterRequest(userData)
	if err != nil {
		SendError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	file, header, err := r.FormFile("avatar")
	if errors.Is(err, http.ErrMissingFile) {
		file, header, err = r.FormFile("image")
	}

	if err == nil && file != nil {
		defer file.Close()
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
		userData.AvatarUrl, err = repo.ImageStorage.Save(file, contentType)
		if err != nil {
			log.Printf("store profile image: %v", err)
			SendError(w, http.StatusInternalServerError, "could not store image")
			return
		}
	} else if err != nil && !errors.Is(err, http.ErrMissingFile) {
		SendError(w, http.StatusBadRequest, "invalid image")
		return
	}
	err = repo.UserRepo.RegisterUser(userData)
	if err != nil {
		if userData.AvatarUrl != "" {
			repo.removeNewImage(userData.AvatarUrl)
		}
		log.Printf("register user: %v", err)
		SendError(
			w,
			http.StatusInternalServerError,
			"could not register user",
		)
		return
	}

	res.Message = "user registered successfully"
	if userData.AvatarUrl != "" {
		res.Data = map[string]string{
			"avatar_url": userData.AvatarUrl,
			"image_url":  "/api/images/" + userData.AvatarUrl,
		}
	}

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

	err = userHandler.UserRepo.DeleteSession(cookie.Value)
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
	user, err := userHandler.UserRepo.GetUserBySession(cookie.Value)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	res.Data = user
	SendJSON(w, http.StatusOK, res)
}

func (userHandler UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	fmt.Println("userbyusername", username)
	if username == "" {
		SendError(w, http.StatusBadRequest, "username is required")
		return
	}
	user_id := r.Context().Value("user_id").(int64)
	user, err := userHandler.UserRepo.GetUserByUsername(user_id, username)
	if err != nil {
		if err == sql.ErrNoRows {
			SendError(w, http.StatusNotFound, "user not found")
			return
		}

		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	SendJSON(w, http.StatusOK, user)
}
func (userHandler UserHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in")
	following_target_id := r.URL.Query().Get("user_id")
	user_target, err := strconv.Atoi(following_target_id)
	if err != nil {
		SendError(w, http.StatusBadRequest, "wrong data send")
		return
	}
	user_id := r.Context().Value("user_id").(int64)
	err = userHandler.UserRepo.FollowUser(user_id, int64(user_target))
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	res := models.Res{Message: "user followed successfully"}
	SendJSON(w, http.StatusOK, res)
}
func (userHandler UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user_id").(int64)
	user, err := userHandler.UserRepo.GetMyUserData(user_id)
	if err != nil {
		if err == sql.ErrNoRows {
			SendError(w, http.StatusNotFound, "user not found")
			return
		}
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	SendJSON(w, http.StatusOK, user)
}

func (userHandler UserHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in")
	Value := r.PathValue("user_id")
	user_target, err := strconv.Atoi(Value)
	if err != nil {
		fmt.Println("error strcnv", err)
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	// user_id := r.Context().Value("user_id").(int64)
	followers, err := userHandler.UserRepo.Followers(int64(user_target))
	if err != nil {
		fmt.Println("error data base GetFollowers ", err)
		if err == sql.ErrNoRows {
			SendError(w, http.StatusNotFound, "no connection found")
		}
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	SendJSON(w, http.StatusOK, followers)
}
func (userHandler UserHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in")
	Value := r.PathValue("user_id")
	user_target, err := strconv.Atoi(Value)
	if err != nil {
		fmt.Println("error strcnv", err)
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	// user_id := r.Context().Value("user_id").(int64)
	connections, err := userHandler.UserRepo.Following(int64(user_target))
	if err != nil {
		fmt.Println("error data base getfoolowin ", err)
		if err == sql.ErrNoRows {
			SendError(w, http.StatusNotFound, "no connection found")
		}
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	SendJSON(w, http.StatusOK, connections)
}
