package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Group update actions pushed over the websocket.
const (
	actionInvited        = "invited"
	actionJoinRequested  = "join_requested"
	actionMemberJoined   = "member_joined"
	actionRequestRefused = "request_refused"
	actionInviteDeclined = "invite_declined"
)

// pushGroupUpdate tells the open pages of every group member, plus the
// extra users given, that something changed in the group. Nothing is stored.
func (userHandler UserHandler) pushGroupUpdate(groupID int64, action string, userID int64, extra ...int64) {
	if userHandler.Hub == nil {
		return
	}
	memberIDs, err := userHandler.UserRepo.GroupMemberIDs(groupID)
	if err != nil {
		log.Printf("group update: get members of group %d: %v", groupID, err)
		return
	}
	push := models.GroupUpdatePush{Type: "group_update", GroupId: groupID, Action: action, UserId: userID}
	sent := map[int64]bool{}
	for _, id := range append(memberIDs, extra...) {
		if id == 0 || sent[id] {
			continue
		}
		sent[id] = true
		if err := userHandler.Hub.SendToUser(id, push); err != nil {
			log.Printf("group update to user %d: %v", id, err)
		}
	}
}

// groupFromPath loads the group named by {group_id} as seen by the current user.
// It writes the error response and returns false when it fails.
func (userHandler UserHandler) groupFromPath(w http.ResponseWriter, r *http.Request) (models.Group, int64, bool) {
	userID := r.Context().Value("user_id").(int64)
	groupID, err := strconv.ParseInt(r.PathValue("group_id"), 10, 64)
	if err != nil || groupID < 1 {
		SendError(w, http.StatusBadRequest, "wrong group id")
		return models.Group{}, userID, false
	}
	group, err := userHandler.UserRepo.GetGroup(userID, groupID)
	if errors.Is(err, sql.ErrNoRows) {
		SendError(w, http.StatusNotFound, "group not found")
		return group, userID, false
	}
	if err != nil {
		log.Printf("get group %d: %v", groupID, err)
		SendError(w, http.StatusInternalServerError, "Failed to retrieve group")
		return group, userID, false
	}
	return group, userID, true
}

func isMemberStatus(status string) bool {
	return status == "admin" || status == "member"
}

// requireGroupMember loads the group from {group_id} and checks that the current
// user is its admin or a member. It writes 403 and returns false otherwise.
func (userHandler UserHandler) requireGroupMember(w http.ResponseWriter, r *http.Request) (models.Group, int64, bool) {
	group, userID, ok := userHandler.groupFromPath(w, r)
	if !ok {
		return group, userID, false
	}
	if !isMemberStatus(group.MyStatus) {
		SendError(w, http.StatusForbidden, "you must be a member of this group")
		return group, userID, false
	}
	return group, userID, true
}

// GET /api/groups/{group_id}/members
func (userHandler UserHandler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	group, _, ok := userHandler.requireGroupMember(w, r)
	if !ok {
		return
	}
	members, err := userHandler.UserRepo.GroupMembers(group.Id)
	if err != nil {
		log.Printf("get members of group %d: %v", group.Id, err)
		SendError(w, http.StatusInternalServerError, "Failed to retrieve members")
		return
	}
	SendJSON(w, http.StatusOK, members)
}

// GET /api/groups/{group_id}/invitable
func (userHandler UserHandler) GetInvitableUsers(w http.ResponseWriter, r *http.Request) {
	group, userID, ok := userHandler.groupFromPath(w, r)
	if !ok {
		return
	}
	if !isMemberStatus(group.MyStatus) {
		SendError(w, http.StatusForbidden, "only group members can invite users")
		return
	}
	users, err := userHandler.UserRepo.InvitableUsers(userID, group.Id)
	if err != nil {
		log.Printf("invitable users: %v", err)
		SendError(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}
	SendJSON(w, http.StatusOK, users)
}

// POST /api/groups/{group_id}/invite  body: {"user_id": N}
func (userHandler UserHandler) InviteUsrGp(w http.ResponseWriter, r *http.Request) {
	group, inviterID, ok := userHandler.groupFromPath(w, r)
	if !ok {
		return
	}
	var req models.InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserId <= 0 {
		SendError(w, http.StatusBadRequest, "wrong user data")
		return
	}
	if !isMemberStatus(group.MyStatus) {
		SendError(w, http.StatusForbidden, "only group members can invite users")
		return
	}
	connected, err := userHandler.UserRepo.IsConnected(inviterID, req.UserId)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to check followers")
		return
	}
	if !connected {
		SendError(w, http.StatusForbidden, "you can only invite your followers or people you follow")
		return
	}

	created, err := userHandler.UserRepo.InviteUsrGp(inviterID, req.UserId, group.Id)
	if err != nil {
		log.Printf("invite user: %v", err)
		SendError(w, http.StatusInternalServerError, "Failed to send invitation")
		return
	}
	if !created {
		SendError(w, http.StatusConflict, "user is already a member or already invited")
		return
	}

	userHandler.notify(req.UserId, inviterID, models.NotifGroupInvite, group.Id, 0,
		fmt.Sprintf("invited you to join %q", group.Title))
	userHandler.pushGroupUpdate(group.Id, actionInvited, req.UserId, req.UserId)
	SendJSON(w, http.StatusOK, models.Res{Message: "invitation sent successfully"})
}

// POST /api/groups/{group_id}/join
// Sends a join request to the group members. If the user was invited, they join directly.
func (userHandler UserHandler) JoinUsrGroup(w http.ResponseWriter, r *http.Request) {
	group, userID, ok := userHandler.groupFromPath(w, r)
	if !ok {
		return
	}
	outcome, inviterID, err := userHandler.UserRepo.RequestToJoin(userID, group.Id)
	switch {
	case errors.Is(err, repository.ErrAlreadyMember):
		SendError(w, http.StatusConflict, "you are already a member of this group")
		return
	case errors.Is(err, repository.ErrAlreadyRequested):
		SendError(w, http.StatusConflict, "your join request is already pending")
		return
	case err != nil:
		log.Printf("join group %d: %v", group.Id, err)
		SendError(w, http.StatusInternalServerError, "Failed to send join request")
		return
	}

	if outcome == repository.JoinAccepted {
		userHandler.notify(inviterID, userID, models.NotifInviteAccepted, group.Id, 0,
			fmt.Sprintf("accepted your invitation to join %q", group.Title))
		userHandler.pushGroupUpdate(group.Id, actionMemberJoined, userID)
		SendJSON(w, http.StatusOK, models.Res{Message: "you joined the group", Data: map[string]string{"status": "member"}})
		return
	}

	// only the creator of the group can answer join requests, so only they are told
	userHandler.notify(group.AdminId, userID, models.NotifJoinRequest, group.Id, 0,
		fmt.Sprintf("wants to join %q", group.Title))
	userHandler.pushGroupUpdate(group.Id, actionJoinRequested, userID, userID)
	SendJSON(w, http.StatusOK, models.Res{Message: "request to join group sent successfully", Data: map[string]string{"status": "pending"}})
}

// POST /api/groups/{group_id}/invitation  body: {"accept": true|false}
// The invited user accepts or declines their invitation.
func (userHandler UserHandler) RespondInvitation(w http.ResponseWriter, r *http.Request) {
	group, userID, ok := userHandler.groupFromPath(w, r)
	if !ok {
		return
	}
	var req models.DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	inviterID, err := userHandler.UserRepo.RespondInvitation(userID, group.Id, req.Accept)
	if errors.Is(err, repository.ErrNoInvitation) {
		SendError(w, http.StatusNotFound, "you have no pending invitation to this group")
		return
	}
	if err != nil {
		log.Printf("respond invitation group %d: %v", group.Id, err)
		SendError(w, http.StatusInternalServerError, "Failed to answer the invitation")
		return
	}

	if req.Accept {
		userHandler.notify(inviterID, userID, models.NotifInviteAccepted, group.Id, 0,
			fmt.Sprintf("accepted your invitation to join %q", group.Title))
		userHandler.pushGroupUpdate(group.Id, actionMemberJoined, userID)
		SendJSON(w, http.StatusOK, models.Res{Message: "you joined the group", Data: map[string]string{"status": "member"}})
		return
	}
	userHandler.notify(inviterID, userID, models.NotifInviteDeclined, group.Id, 0,
		fmt.Sprintf("declined your invitation to join %q", group.Title))
	userHandler.pushGroupUpdate(group.Id, actionInviteDeclined, userID, userID)
	SendJSON(w, http.StatusOK, models.Res{Message: "invitation declined", Data: map[string]string{"status": "none"}})
}

// GET /api/groups/{group_id}/requests  (group creator only)
func (userHandler UserHandler) GetJoinRequests(w http.ResponseWriter, r *http.Request) {
	group, _, ok := userHandler.groupFromPath(w, r)
	if !ok {
		return
	}
	if group.MyStatus != "admin" {
		SendError(w, http.StatusForbidden, "only the creator of the group can see join requests")
		return
	}
	requests, err := userHandler.UserRepo.JoinRequests(group.Id)
	if err != nil {
		log.Printf("join requests group %d: %v", group.Id, err)
		SendError(w, http.StatusInternalServerError, "Failed to retrieve join requests")
		return
	}
	SendJSON(w, http.StatusOK, requests)
}

// POST /api/groups/{group_id}/requests/{user_id}  body: {"accept": true|false}
// Only the creator of the group can accept or refuse a join request.
func (userHandler UserHandler) RespondJoinRequest(w http.ResponseWriter, r *http.Request) {
	group, responderID, ok := userHandler.groupFromPath(w, r)
	if !ok {
		return
	}
	if group.MyStatus != "admin" {
		SendError(w, http.StatusForbidden, "only the creator of the group can answer join requests")
		return
	}
	requesterID, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil || requesterID < 1 {
		SendError(w, http.StatusBadRequest, "wrong user id")
		return
	}
	var req models.DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = userHandler.UserRepo.RespondJoinRequest(group.Id, requesterID, req.Accept)
	if errors.Is(err, repository.ErrNoJoinRequest) {
		SendError(w, http.StatusNotFound, "no pending join request from this user")
		return
	}
	if err != nil {
		log.Printf("respond join request group %d: %v", group.Id, err)
		SendError(w, http.StatusInternalServerError, "Failed to answer the join request")
		return
	}

	if req.Accept {
		userHandler.notify(requesterID, responderID, models.NotifJoinAccepted, group.Id, 0,
			fmt.Sprintf("accepted your request to join %q", group.Title))
		userHandler.pushGroupUpdate(group.Id, actionMemberJoined, requesterID)
		SendJSON(w, http.StatusOK, models.Res{Message: "join request accepted"})
		return
	}
	userHandler.notify(requesterID, responderID, models.NotifJoinRefused, group.Id, 0,
		fmt.Sprintf("refused your request to join %q", group.Title))
	userHandler.pushGroupUpdate(group.Id, actionRequestRefused, requesterID, requesterID)
	SendJSON(w, http.StatusOK, models.Res{Message: "join request refused"})
}
