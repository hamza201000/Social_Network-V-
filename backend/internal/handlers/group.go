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

func (userHandler UserHandler) GetInvitableUsers(w http.ResponseWriter, r *http.Request) {
	group_id, err := strconv.ParseInt(r.PathValue("group_id"), 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "wrong group data")
		return
	}
	user_id := r.Context().Value("user_id").(int64)

	isMember, err := userHandler.UserRepo.IsGroupMember(user_id, group_id)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}
	if !isMember {
		SendError(w, http.StatusForbidden, "only group members can invite users")
		return
	}

	users, err := userHandler.UserRepo.InvitableUsers(user_id, group_id)
	if err != nil {
		fmt.Println("error invitable users:", err)
		SendError(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}
	SendJSON(w, http.StatusOK, users)
}

func (userHandler UserHandler) InviteUsrGp(w http.ResponseWriter, r *http.Request) {
	group_id, err := strconv.ParseInt(r.PathValue("group_id"), 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "wrong group data")
		return
	}
	var req models.InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserId <= 0 {
		SendError(w, http.StatusBadRequest, "wrong user data")
		return
	}
	inviter_id := r.Context().Value("user_id").(int64)

	isMember, err := userHandler.UserRepo.IsGroupMember(inviter_id, group_id)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to check membership")
		return
	}
	if !isMember {
		SendError(w, http.StatusForbidden, "only group members can invite users")
		return
	}

	connected, err := userHandler.UserRepo.IsConnected(inviter_id, req.UserId)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to check followers")
		return
	}
	if !connected {
		SendError(w, http.StatusForbidden, "you can only invite your followers or people you follow")
		return
	}

	created, err := userHandler.UserRepo.InviteUsrGp(req.UserId, group_id)
	if err != nil {
		fmt.Println("error invite user:", err)
		SendError(w, http.StatusInternalServerError, "Failed to send invitation")
		return
	}
	if !created {
		SendError(w, http.StatusConflict, "user is already a member or already invited")
		return
	}
	if group, err := userHandler.UserRepo.GetGroup(group_id); err == nil {
		userHandler.notify(req.UserId, inviter_id, models.NotifGroupInvite, group_id, 0,
			fmt.Sprintf("invited you to join the group %q", group.Title))
	}
	SendJSON(w, http.StatusOK, models.Res{Message: "invitation sent successfully"})
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
func (userHandler UserHandler) JoinUsrGroup(w http.ResponseWriter, r *http.Request) {
	value := r.PathValue("group_id")
	group_id, err := strconv.Atoi(value)
	if err != nil {
		SendError(w, http.StatusBadRequest, "wrong group id")
		return
	}
	user_id := r.Context().Value("user_id").(int64)
	err = userHandler.UserRepo.JoinUsrGroup(user_id, int64(group_id))
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to retrieve groups")
		return
	}
	// tell the group admin someone wants to join
	if group, err := userHandler.UserRepo.GetGroup(int64(group_id)); err == nil {
		userHandler.notify(group.AdminId, user_id, models.NotifJoinRequest, int64(group_id), 0,
			fmt.Sprintf("wants to join your group %q", group.Title))
	}
	res := models.Res{
		Message: "request to join group sent successfully",
	}
	SendJSON(w, http.StatusOK, res)
}
func (userHandler UserHandler) SearchGroup(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	if search == "" {
		SendError(w, http.StatusBadRequest, "wrong data")
		return
	}
	group, err := userHandler.UserRepo.SearchGroup(search)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to retrieve groups")
		return
	}
	SendJSON(w, http.StatusOK, group)
}
func (userHandler UserHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	value := r.PathValue("group_id")
	group_id, err := strconv.Atoi(value)
	if err != nil {
		SendError(w, http.StatusBadRequest, "wrong data group")
		return
	}
	groups, err := userHandler.UserRepo.GetGroup(int64(group_id))
	if err != nil {
		if err == sql.ErrNoRows {
			SendError(w, http.StatusNotFound, "No group fund")
			return
		}
		SendError(w, http.StatusInternalServerError, "Failed to retrieve group")
		return
	}
	SendJSON(w, http.StatusOK, groups)
}

func (userHandler *UserHandler) CreatePostGp(w http.ResponseWriter, r *http.Request) {
	// userID, err := userHandler.requiredUserID(r)
	user_id := r.Context().Value("user_id").(int64)
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
	value := r.PathValue("group_id")
	group_id, err := strconv.Atoi(value)
	if err != nil {
		SendError(w, http.StatusBadRequest, "wrong data group")
		return
	}
	if err := services.ValidatePostRequest(post); err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	postID, err := userHandler.UserRepo.CreatePostGp(int64(group_id), user_id, post)
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
	// userID, err := userHandler.requiredUserID(r)
	user_id := r.Context().Value("user_id").(int64)

	value := r.PathValue("group_id")
	group_id, err := strconv.Atoi(value)
	if err != nil {
		SendError(w, http.StatusBadRequest, "wrong data group")
		return
	}
	user_id = user_id //i have to check authorized
	posts, err := userHandler.UserRepo.GetPostsGp(int64(group_id))
	if err != nil {
		fmt.Println("error getposts:::", err)
		SendError(w, http.StatusInternalServerError, "could not get posts")
		return
	}
	SendJSON(w, http.StatusOK, posts)
}
