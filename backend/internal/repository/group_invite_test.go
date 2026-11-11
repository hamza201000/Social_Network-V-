//go:build cgo

package repository

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestGroupInvitations(t *testing.T) {
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
		CREATE TABLE followers (follower_id INTEGER NOT NULL, following_target_id INTEGER NOT NULL);
		CREATE TABLE groups (id INTEGER PRIMARY KEY AUTOINCREMENT, admin_id INTEGER NOT NULL, title TEXT, description_gp TEXT);
		CREATE TABLE members_gp (id INTEGER PRIMARY KEY AUTOINCREMENT, group_id INTEGER, user_id INTEGER);
		CREATE TABLE members_status (
			id INTEGER PRIMARY KEY AUTOINCREMENT, group_id INTEGER NOT NULL, user_id INTEGER NOT NULL,
			member_status TEXT NOT NULL DEFAULT 'pending', created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		INSERT INTO users (id, username) VALUES (1,'alice'),(2,'bob'),(3,'carol'),(4,'dave'),(5,'erin');
		-- alice follows bob, carol follows alice, alice and dave follow each other; erin is a stranger
		INSERT INTO followers VALUES (1,2),(3,1),(1,4),(4,1);
		INSERT INTO groups (id, admin_id, title, description_gp) VALUES (1, 1, 'g', 'd');
		INSERT INTO members_gp (group_id, user_id) VALUES (1, 4);
	`)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewUserRepository(db)

	if ok, err := repo.IsGroupMember(1, 1); err != nil || !ok {
		t.Fatalf("admin should be member: %v %v", ok, err)
	}
	if ok, err := repo.IsGroupMember(2, 1); err != nil || ok {
		t.Fatalf("bob should not be member: %v %v", ok, err)
	}
	if ok, err := repo.IsConnected(1, 3); err != nil || !ok {
		t.Fatalf("alice/carol connected: %v %v", ok, err)
	}
	if ok, err := repo.IsConnected(1, 5); err != nil || ok {
		t.Fatalf("alice/erin not connected: %v %v", ok, err)
	}

	created, err := repo.InviteUsrGp(2, 1)
	if err != nil || !created {
		t.Fatalf("invite bob: %v %v", created, err)
	}
	if created, _ := repo.InviteUsrGp(2, 1); created {
		t.Fatal("duplicate invite should be ignored")
	}
	if created, _ := repo.InviteUsrGp(4, 1); created {
		t.Fatal("member should not be invited")
	}

	users, err := repo.InvitableUsers(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, u := range users {
		got[u.Username] = u.Status
	}
	want := map[string]string{"bob": "invited", "carol": "none", "dave": "member"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("got %v want %v", got, want)
		}
	}

	if err := repo.JoinUsrGroup(3, 1); err != nil {
		t.Fatal(err)
	}
	var gid, uid int64
	db.QueryRow(`SELECT group_id, user_id FROM members_status WHERE member_status='pending'`).Scan(&gid, &uid)
	if gid != 1 || uid != 3 {
		t.Fatalf("join request stored wrong: group=%d user=%d", gid, uid)
	}
}
