package handler

import (
	"backend/internal/models"
	"backend/internal/services"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func (userHandler UserHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var newGroup models.NewGroupDetails

	err := json.NewDecoder(r.Body).Decode(&newGroup)
	if err != nil {
		SendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	user_id := r.Context().Value("user_id").(int64)
	groupID, err := userHandler.UserRepo.CreateGroup(user_id, newGroup)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}
	response := models.Res{
		Message: "Group created successfully",
		Data:    models.Group{Id: (groupID)},
	}
	SendJSON(w, http.StatusOK, response)
}

func (userHandler UserHandler) GetGroups(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user_id").(int64)
	groups, err := userHandler.UserRepo.GetGroups(user_id)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to retrieve groups")
		return
	}
	SendJSON(w, http.StatusOK, groups)
}

func (userHandler UserHandler) SearchGroup(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	if search == "" {
		SendError(w, http.StatusBadRequest, "wrong data")
		return
	}
	user_id := r.Context().Value("user_id").(int64)
	group, err := userHandler.UserRepo.SearchGroup(user_id, search)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to retrieve groups")
		return
	}
	SendJSON(w, http.StatusOK, group)
}

func (userHandler UserHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	group_id, err := strconv.ParseInt(r.PathValue("group_id"), 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "wrong data group")
		return
	}
	user_id := r.Context().Value("user_id").(int64)
	group, err := userHandler.UserRepo.GetGroup(user_id, group_id)
	if err != nil {
		if err == sql.ErrNoRows {
			SendError(w, http.StatusNotFound, "No group found")
			return
		}
		SendError(w, http.StatusInternalServerError, "Failed to retrieve group")
		return
	}
	SendJSON(w, http.StatusOK, group)
}

func (userHandler *UserHandler) CreatePostGp(w http.ResponseWriter, r *http.Request) {
	// check membership before reading the body, so no image is saved for a refused request
	group, user_id, ok := userHandler.requireGroupMember(w, r)
	if !ok {
		return
	}
	post, err := parsePostRequest(w, r, userHandler.ImageStorage)
	if err != nil {
		sendImageRequestError(w, err)
		return
	}
	imagePersisted := false
	defer func() {
		if post.Image != "" && !imagePersisted {
			userHandler.removeNewImage(post.Image)
		}
	}()
	if err := services.ValidatePostRequest(post); err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	postID, err := userHandler.UserRepo.CreatePostGp(group.Id, user_id, post)
	if err != nil {
		fmt.Println("error createpost:::", err)
		SendError(w, http.StatusInternalServerError, "could not create post")
		return
	}
	imagePersisted = true
	SendJSON(w, http.StatusCreated, models.Res{
		Message: "post created successfully",
		Data: map[string]any{
			"id":        postID,
			"image":     post.Image,
			"image_url": imageURL(post.Image),
		},
	})
}

func (userHandler *UserHandler) GetPostGp(w http.ResponseWriter, r *http.Request) {
	group, _, ok := userHandler.requireGroupMember(w, r)
	if !ok {
		return
	}
	posts, err := userHandler.UserRepo.GetPostsGp(group.Id)
	if err != nil {
		fmt.Println("error getposts:::", err)
		SendError(w, http.StatusInternalServerError, "could not get posts")
		return
	}
	SendJSON(w, http.StatusOK, posts)
}
