//go:build cgo

package repository

import (
	"database/sql"
	"errors"
	"testing"

	"backend/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

func TestMessageRepository(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT NOT NULL);
		CREATE TABLE followers (
			follower_id INTEGER NOT NULL,
			following_target_id INTEGER NOT NULL,
			PRIMARY KEY (follower_id, following_target_id)
		);
		CREATE TABLE messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sender_id INTEGER NOT NULL,
			receiver_id INTEGER NOT NULL,
			message TEXT NOT NULL,
			seen BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO users (id, username)
		VALUES (1, 'alice'), (2, 'bob'), (3, 'carol'), (4, 'dave');
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewMessageRepository(db)
	allowed, err := repo.CanMessage(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if allowed {
		t.Fatal("users with no follow relationship should not be allowed to message")
	}

	if _, err := db.Exec(`INSERT INTO followers (follower_id, following_target_id) VALUES (2, 1)`); err != nil {
		t.Fatal(err)
	}
	allowed, err = repo.CanMessage(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Fatal("a follow in either direction should allow messaging")
	}
	if _, err := db.Exec(`
		INSERT INTO followers (follower_id, following_target_id)
		VALUES (1, 3), (4, 1)
	`); err != nil {
		t.Fatal(err)
	}

	chats, err := repo.GetChatUsers(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(chats) != 3 {
		t.Fatalf("expected followers and followed users before any messages, got %+v", chats)
	}
	for _, chat := range chats {
		if chat.LastMsg != nil {
			t.Errorf("expected no last message for user %d, got %+v", chat.UserID, chat.LastMsg)
		}
	}

	saved, err := repo.InsertMessage(1, 2, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == 0 || saved.UsernameSender != "alice" || saved.Seen {
		t.Fatalf("unexpected saved message: %+v", saved)
	}

	conversation, err := repo.GetMessagesBetweenUsers(2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(conversation) != 1 || conversation[0].ID != saved.ID {
		t.Fatalf("unexpected conversation messages: %+v", conversation)
	}

	chats, err = repo.GetChatUsers(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(chats) != 3 {
		t.Fatalf("unexpected chat list: %+v", chats)
	}
	var bobChat *models.Client
	for i := range chats {
		if chats[i].UserID == 2 {
			bobChat = &chats[i]
			break
		}
	}
	if bobChat == nil || bobChat.LastMsg == nil || bobChat.LastMsg.ID != saved.ID {
		t.Fatalf("expected Bob's chat to include the latest message: %+v", chats)
	}

	if err := repo.MarkMessageSeen(saved.ID, 1); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("sender must not mark their own message as seen; got %v", err)
	}
	if err := repo.MarkMessageSeen(saved.ID, 2); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkMessageSeen(saved.ID, 2); err != nil {
		t.Fatalf("marking an already-seen message should be safe: %v", err)
	}

	latest, err := repo.GetLatestMessage(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if latest.ID != saved.ID || !latest.Seen || latest.Type != "message" {
		t.Fatalf("unexpected latest message: %+v", latest)
	}
	if latest.UsernameSender == "" || latest.CreatedAt.IsZero() || latest.Message != "hello" {
		t.Fatalf("incomplete message data: %+v", latest)
	}
}
