package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func eventGroupID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("group_id"), 10, 64)
}

func (userHandler UserHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.NewEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		SendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	event.Title = strings.TrimSpace(event.Title)
	event.EventAt = strings.TrimSpace(event.EventAt)
	if event.Title == "" || event.EventAt == "" {
		SendError(w, http.StatusBadRequest, "title and event_at are required")
		return
	}
	if _, err := time.Parse(time.RFC3339, event.EventAt); err != nil {
		SendError(w, http.StatusBadRequest, "event_at must be an RFC3339 date and time")
		return
	}
	groupID, err := eventGroupID(r)
	if err != nil || groupID < 1 {
		SendError(w, http.StatusBadRequest, "Invalid group id")
		return
	}
	userID := r.Context().Value("user_id").(int64)
	eventID, err := userHandler.UserRepo.CreateEvent(userID, groupID, event)
	if errors.Is(err, repository.ErrNotGroupMember) {
		SendError(w, http.StatusForbidden, "You must belong to this group to create an event")
		return
	}
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to create event")
		return
	}
	SendJSON(w, http.StatusCreated, models.Res{Message: "Event created successfully", Data: map[string]int64{"id": eventID}})
}

func (userHandler UserHandler) GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	groupID, err := eventGroupID(r)
	if err != nil || groupID < 1 {
		SendError(w, http.StatusBadRequest, "Invalid group id")
		return
	}
	userID := r.Context().Value("user_id").(int64)
	events, err := userHandler.UserRepo.GetGroupEvents(userID, groupID)
	if errors.Is(err, repository.ErrNotGroupMember) {
		SendError(w, http.StatusForbidden, "You must belong to this group to view its events")
		return
	}
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to retrieve events")
		return
	}
	SendJSON(w, http.StatusOK, events)
}

func (userHandler UserHandler) VoteEvent(w http.ResponseWriter, r *http.Request) {
	groupID, err := eventGroupID(r)
	if err != nil || groupID < 1 {
		SendError(w, http.StatusBadRequest, "Invalid group id")
		return
	}
	eventID, err := strconv.ParseInt(r.PathValue("event_id"), 10, 64)
	if err != nil || eventID < 1 {
		SendError(w, http.StatusBadRequest, "Invalid event id")
		return
	}
	var vote models.EventVote
	if err := json.NewDecoder(r.Body).Decode(&vote); err != nil {
		SendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if vote.Response != "Going" && vote.Response != "Not going" {
		SendError(w, http.StatusBadRequest, "response must be Going or Not going")
		return
	}
	userID := r.Context().Value("user_id").(int64)
	err = userHandler.UserRepo.VoteEvent(userID, groupID, eventID, vote.Response)
	if errors.Is(err, repository.ErrNotGroupMember) {
		SendError(w, http.StatusForbidden, "You must belong to this group to respond")
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println("this error ???")
		SendError(w, http.StatusNotFound, "Event not found")
		return
	}
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to save event response")
		return
	}
	SendJSON(w, http.StatusOK, models.Res{Message: "Event response saved successfully"})
}
