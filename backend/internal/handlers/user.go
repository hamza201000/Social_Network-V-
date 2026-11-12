package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/services"
	chatws "backend/internal/websocket"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"net/http"
)

type UserHandler struct {
	UserRepo     *repository.UserRepository
	ImageStorage *services.ImageStorage
	Hub          *chatws.Hub // used to push notifications live
}

func NewUserHandler(userRepo *repository.UserRepository, imageStorage *services.ImageStorage, hub *chatws.Hub) *UserHandler {
	return &UserHandler{
		UserRepo:     userRepo,
		ImageStorage: imageStorage,
		Hub:          hub,
	}
}

func (userHandler UserHandler) Login(w http.ResponseWriter, r *http.Request) {

	var dataLogin models.LoginRequest
	var res models.Res
	err := json.NewDecoder(r.Body).Decode(&dataLogin)
	if err != nil {
		SendError(w, http.StatusBadRequest, "invalid request")
		return
	}
	dataLogin.Email = strings.TrimSpace(dataLogin.Email)
	if dataLogin.Email == "" || dataLogin.Password == "" {
		SendError(w, http.StatusBadRequest, "email/username and password are required")
		return
	}
	user, err := userHandler.UserRepo.LoginUser(dataLogin)
	if errors.Is(err, repository.ErrInvalidCredentials) {
		SendError(w, http.StatusUnauthorized, "invalid email/username or password")
		return
	}
	if err != nil {
		log.Printf("login: %v", err)
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
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
		MaxAge:   int(repository.SessionDuration.Seconds()),
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

	// the nickname is optional: without one, build a unique one from the name
	if userData.Username == "" {
		username, genErr := repo.generateUsername(userData.FirstName, userData.LastName)
		if genErr != nil {
			log.Printf("generate username: %v", genErr)
			SendError(w, http.StatusInternalServerError, "could not register user")
			return
		}
		userData.Username = username
	}

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
			"image/gif":  true,
		}
		if !allowedTypes[contentType] {
			SendError(
				w,
				http.StatusBadRequest,
				"only JPEG, PNG, WebP and GIF images are allowed",
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
		if errors.Is(err, repository.ErrEmailTaken) || errors.Is(err, repository.ErrUsernameTaken) {
			SendError(w, http.StatusConflict, err.Error())
			return
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
	cookie := r.Context().Value("session_id").(string)
	fmt.Print("session", cookie)
	err := userHandler.UserRepo.DeleteSession(cookie)
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

func (userHandler UserHandler) SearchUser(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	if search == "" {
		SendError(w, http.StatusBadRequest, "wrong data")
		return
	}
	user, err := userHandler.UserRepo.SearchUser(search)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}
	SendJSON(w, http.StatusOK, user)
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

// POST /api/user/follow?user_id=ID
// Follows a public user, sends a follow request to a private user, or
// unfollows / cancels the request when called again.
func (userHandler UserHandler) FollowUser(w http.ResponseWriter, r *http.Request) {
	user_target, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil || user_target < 1 {
		SendError(w, http.StatusBadRequest, "wrong data send")
		return
	}
	user_id := r.Context().Value("user_id").(int64)
	status, err := userHandler.UserRepo.FollowUser(user_id, user_target)
	if errors.Is(err, repository.ErrFollowSelf) {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		SendError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Printf("follow user %d: %v", user_target, err)
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	switch status {
	case "following":
		userHandler.notify(user_target, user_id, models.NotifNewFollower, 0, 0, "started following you")
	case "requested":
		userHandler.notify(user_target, user_id, models.NotifFollowRequest, 0, 0, "wants to follow you")
	}
	SendJSON(w, http.StatusOK, models.Res{Message: "follow status updated", Data: map[string]string{"follow_status": status}})
}

// POST /api/follow-requests/{user_id}  body: {"accept": true|false}
// The logged-in (private) user answers the follow request sent by user_id.
func (userHandler UserHandler) RespondFollowRequest(w http.ResponseWriter, r *http.Request) {
	requesterID, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil || requesterID < 1 {
		SendError(w, http.StatusBadRequest, "wrong user id")
		return
	}
	var req models.DecisionFollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	user_id := r.Context().Value("user_id").(int64)
	err = userHandler.UserRepo.RespondFollowRequest(user_id, requesterID, req.Accept)
	if errors.Is(err, repository.ErrNoFollowRequest) {
		SendError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		log.Printf("respond follow request: %v", err)
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	if req.Accept {
		userHandler.notify(requesterID, user_id, models.NotifFollowAccepted, 0, 0, "accepted your follow request")
		SendJSON(w, http.StatusOK, models.Res{Message: "follow request accepted"})
		return
	}
	SendJSON(w, http.StatusOK, models.Res{Message: "follow request declined"})
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
	user_id := r.Context().Value("user_id").(int64)
	allowed, err := userHandler.UserRepo.CanViewProfile(user_id, int64(user_target))
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	if !allowed {
		SendError(w, http.StatusForbidden, "this profile is private")
		return
	}
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
	user_id := r.Context().Value("user_id").(int64)
	allowed, err := userHandler.UserRepo.CanViewProfile(user_id, int64(user_target))
	if err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	if !allowed {
		SendError(w, http.StatusForbidden, "this profile is private")
		return
	}
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

func (userHandler UserHandler) UpdatePrivacy(w http.ResponseWriter, r *http.Request) {
	var req models.PrivacyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "invalid data")
		return
	}
	user_id := r.Context().Value("user_id").(int64)
	if err := userHandler.UserRepo.SetPrivacy(user_id, req.Private); err != nil {
		SendError(w, http.StatusInternalServerError, "internal server")
		return
	}
	SendJSON(w, http.StatusOK, models.Res{Message: "privacy updated", Data: map[string]bool{"is_private": req.Private}})
}

// generateUsername builds a free nickname like "john_doe" or "john_doe42".
func (repo *UserHandler) generateUsername(firstName, lastName string) (string, error) {
	clean := func(value string) string {
		var b strings.Builder
		for _, r := range strings.ToLower(value) {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
			}
		}
		return b.String()
	}
	base := strings.Trim(clean(firstName)+"_"+clean(lastName), "_")
	if len(base) < 3 {
		base = "user_" + base
	}
	if len(base) > 24 {
		base = base[:24]
	}
	candidate := base
	for i := 0; i < 50; i++ {
		taken, err := repo.UserRepo.UsernameTaken(candidate)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s%d", base, rand.Intn(10000))
	}
	return "", errors.New("no free username found")
}
