package handler

import (
	"backend/internal/models"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
)

// notify saves a notification for userID and pushes it live to every open
// websocket of that user. It never fails the caller's request: errors are only logged.
// groupID and eventID can be 0 when they don't apply.
func (userHandler UserHandler) notify(userID, actorID int64, notifType string, groupID, eventID int64, message string) {
	if userID == actorID {
		return // never notify someone about their own action
	}
	notification, err := userHandler.UserRepo.CreateNotification(userID, actorID, notifType, groupID, eventID, message)
	if err != nil {
		log.Printf("create %s notification for user %d: %v", notifType, userID, err)
		return
	}
	if userHandler.Hub == nil {
		return
	}
	push := models.NotificationPush{Type: "notification", Notification: notification}
	if err := userHandler.Hub.SendToUser(userID, push); err != nil {
		log.Printf("push notification to user %d: %v", userID, err)
	}
}

// GET /api/notifications
func (userHandler UserHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)
	list, err := userHandler.UserRepo.GetNotifications(userID, 50)
	if err != nil {
		log.Printf("get notifications: %v", err)
		SendError(w, http.StatusInternalServerError, "Failed to retrieve notifications")
		return
	}
	SendJSON(w, http.StatusOK, list)
}

// PUT /api/notifications/{id}/read
func (userHandler UserHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)
	notificationID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || notificationID < 1 {
		SendError(w, http.StatusBadRequest, "Invalid notification id")
		return
	}
	err = userHandler.UserRepo.MarkNotificationRead(notificationID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		SendError(w, http.StatusNotFound, "Notification not found")
		return
	}
	if err != nil {
		log.Printf("mark notification read: %v", err)
		SendError(w, http.StatusInternalServerError, "Failed to update notification")
		return
	}
	SendJSON(w, http.StatusOK, models.Res{Message: "notification marked as read"})
}

// PUT /api/notifications/read-all
func (userHandler UserHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)
	if err := userHandler.UserRepo.MarkAllNotificationsRead(userID); err != nil {
		log.Printf("mark all notifications read: %v", err)
		SendError(w, http.StatusInternalServerError, "Failed to update notifications")
		return
	}
	SendJSON(w, http.StatusOK, models.Res{Message: "all notifications marked as read"})
}
