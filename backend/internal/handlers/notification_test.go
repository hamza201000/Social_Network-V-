//go:build cgo

package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	chatws "backend/internal/websocket"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	gorilla "github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

// A connected user receives a notification over the websocket as soon as notify is called.
func TestNotifyPushesOverWebsocket(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT NOT NULL, avatar_url TEXT NOT NULL DEFAULT '');
		CREATE TABLE groups (id INTEGER PRIMARY KEY, admin_id INTEGER NOT NULL, title TEXT);
		CREATE TABLE events (id INTEGER PRIMARY KEY);
		INSERT INTO users (id, username) VALUES (1, 'alice'), (2, 'bob');
	`)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../../database/migrations/000009_create_notifications.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(migration)); err != nil {
		t.Fatal(err)
	}

	hub := chatws.NewHub()
	userHandler := UserHandler{UserRepo: repository.NewUserRepository(db), Hub: hub}

	// a websocket endpoint that registers the client as user 2 (bob)
	upgrader := gorilla.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		hub.Register(2, conn)
	}))
	defer server.Close()

	client, _, err := gorilla.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for i := 0; !hub.IsOnline(2); i++ {
		if i > 100 {
			t.Fatal("client never registered")
		}
		time.Sleep(10 * time.Millisecond)
	}

	userHandler.notify(2, 1, models.NotifNewFollower, 0, 0, "started following you")
	userHandler.notify(1, 1, models.NotifNewFollower, 0, 0, "self") // must be ignored

	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	var push models.NotificationPush
	if err := client.ReadJSON(&push); err != nil {
		t.Fatal(err)
	}
	if push.Type != "notification" || push.Notification.ActorUsername != "alice" ||
		push.Notification.Type != models.NotifNewFollower || push.Notification.UserID != 2 {
		t.Fatalf("unexpected push: %+v", push)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM notifications`).Scan(&count)
	if count != 1 {
		t.Fatalf("expected 1 stored notification (self-notification skipped), got %d", count)
	}
}
