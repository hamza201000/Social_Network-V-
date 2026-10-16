package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	chatws "backend/internal/websocket"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	gorilla "github.com/gorilla/websocket"
)

type ChatHandler struct {
	MessageRepo *repository.MessageRepository
	Hub         *chatws.Hub
}

func NewChatHandler(messageRepo *repository.MessageRepository, hub *chatws.Hub) *ChatHandler {
	return &ChatHandler{MessageRepo: messageRepo, Hub: hub}
}

var chatUpgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" || origin == "http://localhost:3000"
	},
}

func (handler *ChatHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := authenticatedUserID(r)
	if !ok {
		SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	otherUserID, err := strconv.ParseInt(r.PathValue("userID"), 10, 64)
	if err != nil || otherUserID < 1 || otherUserID == currentUserID {
		SendError(w, http.StatusBadRequest, "Invalid user id")
		return
	}

	messages, err := handler.MessageRepo.GetMessagesBetweenUsers(currentUserID, otherUserID)
	if err != nil {
		log.Printf("get chat messages: %v", err)
		SendError(w, http.StatusInternalServerError, "Failed to retrieve messages")
		return
	}
	SendJSON(w, http.StatusOK, messages)
}

func (handler *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user_id").(int64)

	chats, err := handler.MessageRepo.GetChatUsers(user_id)
	if err != nil {
		log.Printf("get chats: %v", err)
		SendError(w, http.StatusInternalServerError, "Failed to retrieve chats")
		return
	}
	for i := range chats {
		chats[i].IsOnline = handler.Hub.IsOnline(chats[i].UserID)
	}
	SendJSON(w, http.StatusOK, chats)
}

func (handler *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	senderID, ok := authenticatedUserID(r)
	if !ok {
		SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var request models.NewMessage
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
		SendError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	request.Message = strings.TrimSpace(request.Message)
	if request.ReceiverID < 1 || request.ReceiverID == senderID || request.Message == "" || len(request.Message) > 5000 {
		SendError(w, http.StatusBadRequest, "receiver_id and a message of 1 to 5000 bytes are required")
		return
	}

	message, err := handler.saveMessage(senderID, request)
	if errors.Is(err, repository.ErrMessageNotAllowed) {
		SendError(w, http.StatusForbidden, "You can message a user only if one of you follows the other")
		return
	}
	if err != nil {
		log.Printf("send chat message: %v", err)
		SendError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}
	SendJSON(w, http.StatusCreated, message)
}

func (handler *ChatHandler) MarkSeen(w http.ResponseWriter, r *http.Request) {
	receiverID, ok := authenticatedUserID(r)
	if !ok {
		SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	messageID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || messageID < 1 {
		SendError(w, http.StatusBadRequest, "Invalid message id")
		return
	}
	if err := handler.MessageRepo.MarkMessageSeen(messageID, receiverID); errors.Is(err, sql.ErrNoRows) {
		SendError(w, http.StatusNotFound, "Message not found")
		return
	} else if err != nil {
		log.Printf("mark message seen: %v", err)
		SendError(w, http.StatusInternalServerError, "Failed to mark message as seen")
		return
	}
	SendJSON(w, http.StatusOK, map[string]bool{"seen": true})
}

func (handler *ChatHandler) WebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	connection, err := chatUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade chat websocket: %v", err)
		return
	}
	client := handler.Hub.Register(userID, connection)
	defer func() {
		handler.Hub.Remove(userID, client)
		connection.Close()
	}()
	connection.SetReadLimit(1 << 20)

	for {
		var request models.NewMessage
		if err := connection.ReadJSON(&request); err != nil {
			if gorilla.IsUnexpectedCloseError(err, gorilla.CloseGoingAway, gorilla.CloseAbnormalClosure) {
				log.Printf("read chat websocket: %v", err)
			}
			return
		}

		request.Message = strings.TrimSpace(request.Message)
		if request.ReceiverID < 1 || request.ReceiverID == userID || request.Message == "" || len(request.Message) > 5000 {
			if err := client.Send(websocketError("receiver_id and a message of 1 to 5000 bytes are required")); err != nil {
				return
			}
			continue
		}

		message, err := handler.saveMessage(userID, request)
		if errors.Is(err, repository.ErrMessageNotAllowed) {
			if sendErr := client.Send(websocketError("You can message a user only if one of you follows the other")); sendErr != nil {
				return
			}
			continue
		}
		if err != nil {
			log.Printf("save websocket message: %v", err)
			if sendErr := client.Send(websocketError("Failed to send message")); sendErr != nil {
				return
			}
			continue
		}
		if err := client.Send(message); err != nil {
			log.Printf("acknowledge websocket message: %v", err)
			return
		}
	}
}

func (handler *ChatHandler) saveMessage(senderID int64, request models.NewMessage) (models.DataMessage, error) {
	allowed, err := handler.MessageRepo.CanMessage(senderID, request.ReceiverID)
	if err != nil {
		return models.DataMessage{}, err
	}
	if !allowed {
		return models.DataMessage{}, repository.ErrMessageNotAllowed
	}

	message, err := handler.MessageRepo.InsertMessage(senderID, request.ReceiverID, request.Message)
	if err != nil {
		return models.DataMessage{}, err
	}
	if err := handler.Hub.SendToUser(request.ReceiverID, message); err != nil {
		log.Printf("deliver chat message to user %d: %v", request.ReceiverID, err)
	}
	return message, nil
}

func authenticatedUserID(r *http.Request) (int64, bool) {
	
	userID, ok := r.Context().Value("user_id").(int64)
	return userID, ok && userID > 0
}

func websocketError(message string) models.DataMessage {
	return models.DataMessage{
		Type:    "error",
		Message: message,
	}
}
