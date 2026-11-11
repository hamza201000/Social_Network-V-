//go:build cgo

package repository

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNotifications(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY, first_name TEXT NOT NULL DEFAULT '', last_name TEXT NOT NULL DEFAULT '',
			username TEXT NOT NULL, avatar_url TEXT NOT NULL DEFAULT ''
		);
		CREATE TABLE groups (id INTEGER PRIMARY KEY AUTOINCREMENT, admin_id INTEGER NOT NULL, title TEXT, description_gp TEXT);
		CREATE TABLE members_gp (id INTEGER PRIMARY KEY AUTOINCREMENT, group_id INTEGER, user_id INTEGER);
		CREATE TABLE events (id INTEGER PRIMARY KEY AUTOINCREMENT);
		INSERT INTO users (id, username, avatar_url) VALUES (1,'alice','a.jpg'),(2,'bob',''),(3,'carol','');
		INSERT INTO groups (id, admin_id, title, description_gp) VALUES (1, 1, 'Hikers', 'd');
		INSERT INTO members_gp (group_id, user_id) VALUES (1, 2), (1, 3);
	`)
	if err != nil {
		t.Fatal(err)
	}
	// use the real migration so the test breaks if the schema and the queries drift apart
	migration, err := os.ReadFile("../../database/migrations/000009_create_notifications.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(migration)); err != nil {
		t.Fatal(err)
	}
	repo := NewUserRepository(db)

	follow, err := repo.CreateNotification(2, 1, "new_follower", 0, 0, "started following you")
	if err != nil {
		t.Fatal(err)
	}
	if follow.ActorUsername != "alice" || follow.ActorAvatar != "a.jpg" || follow.GroupID != nil || follow.IsRead {
		t.Fatalf("unexpected follow notification: %+v", follow)
	}

	invite, err := repo.CreateNotification(2, 1, "group_invite", 1, 0, "invited you")
	if err != nil {
		t.Fatal(err)
	}
	if invite.GroupID == nil || *invite.GroupID != 1 || invite.GroupTitle != "Hikers" {
		t.Fatalf("unexpected invite notification: %+v", invite)
	}

	list, err := repo.GetNotifications(2, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Notifications) != 2 || list.UnreadCount != 2 || list.Notifications[0].ID != invite.ID {
		t.Fatalf("bob should have 2 unread, newest first: %+v", list)
	}
	if other, _ := repo.GetNotifications(3, 50); len(other.Notifications) != 0 {
		t.Fatalf("carol should have none: %+v", other)
	}

	// another user can't mark bob's notification
	if err := repo.MarkNotificationRead(follow.ID, 3); err != sql.ErrNoRows {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}
	if err := repo.MarkNotificationRead(follow.ID, 2); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.GetNotifications(2, 50); list.UnreadCount != 1 {
		t.Fatalf("expected 1 unread, got %d", list.UnreadCount)
	}
	if err := repo.MarkAllNotificationsRead(2); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.GetNotifications(2, 50); list.UnreadCount != 0 {
		t.Fatalf("expected 0 unread, got %d", list.UnreadCount)
	}

	ids, err := repo.GroupMemberIDs(1)
	if err != nil || len(ids) != 3 {
		t.Fatalf("group 1 should have admin + 2 members: %v %v", ids, err)
	}
}
