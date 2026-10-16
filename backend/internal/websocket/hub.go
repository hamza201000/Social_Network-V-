package websocket

import (
	"backend/internal/models"
	"sync"
	"time"

	gorilla "github.com/gorilla/websocket"
)

type Client struct {
	connection *gorilla.Conn
	writeMu    sync.Mutex
}

func (client *Client) Send(message models.DataMessage) error {
	client.writeMu.Lock()
	defer client.writeMu.Unlock()

	if err := client.connection.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}
	return client.connection.WriteJSON(message)
}

func (client *Client) Close() error {
	return client.connection.Close()
}

type Hub struct {
	mu      sync.RWMutex
	clients map[int64]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[int64]map[*Client]struct{})}
}

func (hub *Hub) Register(userID int64, connection *gorilla.Conn) *Client {
	client := &Client{connection: connection}
	hub.mu.Lock()
	defer hub.mu.Unlock()
	if hub.clients[userID] == nil {
		hub.clients[userID] = make(map[*Client]struct{})
	}
	hub.clients[userID][client] = struct{}{}
	return client
}

func (hub *Hub) Remove(userID int64, client *Client) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	delete(hub.clients[userID], client)
	if len(hub.clients[userID]) == 0 {
		delete(hub.clients, userID)
	}
}

func (hub *Hub) IsOnline(userID int64) bool {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	return len(hub.clients[userID]) > 0
}

func (hub *Hub) SendToUser(userID int64, message models.DataMessage) error {
	hub.mu.RLock()
	clients := make([]*Client, 0, len(hub.clients[userID]))
	for client := range hub.clients[userID] {
		clients = append(clients, client)
	}
	hub.mu.RUnlock()

	var sendErr error
	for _, client := range clients {
		if err := client.Send(message); err != nil {
			hub.Remove(userID, client)
			client.Close()
			if sendErr == nil {
				sendErr = err
			}
		}
	}
	return sendErr
}
