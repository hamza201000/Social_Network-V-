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
		CREATE TABLE groups (id INTEGER PRIMARY KEY AUTOINCREMENT, admin_id INTEGER NOT NULL, title TEXT, description_gp TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE members_gp (id INTEGER PRIMARY KEY AUTOINCREMENT, group_id INTEGER, user_id INTEGER, created_at DATETIME DEFAULT CURRENT_TIMESTAMP);
		CREATE TABLE members_status (
			id INTEGER PRIMARY KEY AUTOINCREMENT, group_id INTEGER NOT NULL, user_id INTEGER NOT NULL,
			member_status TEXT NOT NULL DEFAULT 'pending', created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			invited_by INTEGER DEFAULT NULL
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

	created, err := repo.InviteUsrGp(1, 2, 1)
	if err != nil || !created {
		t.Fatalf("invite bob: %v %v", created, err)
	}
	if created, _ := repo.InviteUsrGp(1, 2, 1); created {
		t.Fatal("duplicate invite should be ignored")
	}
	if created, _ := repo.InviteUsrGp(1, 4, 1); created {
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

	// group status as seen by different users
	statusOf := func(viewer int64) string {
		g, err := repo.GetGroup(viewer, 1)
		if err != nil {
			t.Fatal(err)
		}
		return g.MyStatus
	}
	if s := statusOf(1); s != "admin" {
		t.Fatalf("alice status %q", s)
	}
	if s := statusOf(2); s != "invited" {
		t.Fatalf("bob status %q", s)
	}
	if g, _ := repo.GetGroup(1, 1); g.MemberCount != 2 {
		t.Fatalf("member count %d, want 2 (admin + dave)", g.MemberCount)
	}

	// carol asks to join; asking twice fails; the admin accepts
	if out, _, err := repo.RequestToJoin(3, 1); err != nil || out != JoinRequested {
		t.Fatalf("carol join: %q %v", out, err)
	}
	if _, _, err := repo.RequestToJoin(3, 1); err != ErrAlreadyRequested {
		t.Fatalf("second request: %v", err)
	}
	if reqs, _ := repo.JoinRequests(1); len(reqs) != 1 || reqs[0].Username != "carol" {
		t.Fatalf("requests: %+v", reqs)
	}
	if err := repo.RespondJoinRequest(1, 3, true); err != nil {
		t.Fatal(err)
	}
	if s := statusOf(3); s != "member" {
		t.Fatalf("carol status %q", s)
	}
	if err := repo.RespondJoinRequest(1, 3, true); err != ErrNoJoinRequest {
		t.Fatalf("answering twice: %v", err)
	}

	// erin asks and is refused
	repo.RequestToJoin(5, 1)
	if err := repo.RespondJoinRequest(1, 5, false); err != nil {
		t.Fatal(err)
	}
	if s := statusOf(5); s != "none" {
		t.Fatalf("erin status %q", s)
	}

	// bob accepts his invitation and learns who invited him
	inviter, err := repo.RespondInvitation(2, 1, true)
	if err != nil || inviter != 1 {
		t.Fatalf("bob accept: inviter=%d err=%v", inviter, err)
	}
	if _, err := repo.RespondInvitation(2, 1, true); err != ErrNoInvitation {
		t.Fatalf("accept twice: %v", err)
	}
	if _, _, err := repo.RequestToJoin(2, 1); err != ErrAlreadyMember {
		t.Fatalf("member asking to join: %v", err)
	}

	members, err := repo.GroupMembers(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 4 || members[0].Username != "alice" || members[0].Role != "admin" {
		t.Fatalf("members: %+v", members)
	}

	// an invited user who asks to join is let in directly
	db.Exec(`INSERT INTO groups (id, admin_id, title) VALUES (2, 4, 'g2')`)
	repo.InviteUsrGp(4, 1, 2)
	out, inviter, err := repo.RequestToJoin(1, 2)
	if err != nil || out != JoinAccepted || inviter != 4 {
		t.Fatalf("invited join: %q %d %v", out, inviter, err)
	}
}
