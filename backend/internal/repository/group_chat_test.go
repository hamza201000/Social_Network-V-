//go:build cgo

package repository

import (
	"errors"
	"testing"
)

func TestGroupChatOnlyForMembers(t *testing.T) {
	db := openMigratedDB(t)
	defer db.Close()
	if _, err := db.Exec(`
		INSERT INTO users (id, first_name, last_name, username, email, password_hash) VALUES
			(1,'A','A','admin','a@x','x'), (2,'M','M','member','m@x','x'), (3,'O','O','outsider','o@x','x');
		INSERT INTO groups (id, admin_id, title, description_gp) VALUES (1, 1, 'Club', 'd');
		INSERT INTO members_gp (group_id, user_id) VALUES (1, 2);`); err != nil {
		t.Fatal(err)
	}
	repo := NewMessageRepository(db)

	if _, err := repo.InsertGroupMessage(2, 1, "hello 👋"); err != nil {
		t.Fatalf("member should post: %v", err)
	}
	if _, err := repo.InsertGroupMessage(1, 1, "hi"); err != nil {
		t.Fatalf("admin should post: %v", err)
	}
	if _, err := repo.InsertGroupMessage(3, 1, "let me in"); !errors.Is(err, ErrNotGroupChatMember) {
		t.Fatalf("outsider post: want ErrNotGroupChatMember, got %v", err)
	}
	if _, err := repo.GetGroupMessages(3, 1); !errors.Is(err, ErrNotGroupChatMember) {
		t.Fatalf("outsider read: want ErrNotGroupChatMember, got %v", err)
	}
	messages, err := repo.GetGroupMessages(1, 1)
	if err != nil || len(messages) != 2 || messages[0].Message != "hello 👋" || messages[0].UsernameSender != "member" {
		t.Fatalf("admin read: %+v %v", messages, err)
	}
	ids, err := repo.GroupMemberIDs(1)
	if err != nil || len(ids) != 2 {
		t.Fatalf("member ids: %v %v", ids, err)
	}
}

func TestCanMessageNeedsFollowEitherWay(t *testing.T) {
	db := openMigratedDB(t)
	defer db.Close()
	if _, err := db.Exec(`
		INSERT INTO users (id, first_name, last_name, username, email, password_hash, private) VALUES
			(1,'S','S','sender','s@x','x',1), (2,'P','P','public','p@x','x',0),
			(3,'F','F','follower','f@x','x',1), (4,'T','T','followed','t@x','x',1);
		INSERT INTO followers (follower_id, following_target_id) VALUES (3, 1), (1, 4);`); err != nil {
		t.Fatal(err)
	}
	repo := NewMessageRepository(db)
	cases := []struct {
		receiver int64
		want     bool
	}{
		{2, false}, // public profile but no follow either way
		{3, true},  // receiver follows the sender
		{4, true},  // sender follows the receiver
	}
	for _, c := range cases {
		got, err := repo.CanMessage(1, c.receiver)
		if err != nil || got != c.want {
			t.Errorf("CanMessage(1, %d) = %v, %v; want %v", c.receiver, got, err, c.want)
		}
		back, err := repo.CanMessage(c.receiver, 1)
		if err != nil || back != c.want {
			t.Errorf("CanMessage(%d, 1) = %v, %v; want %v", c.receiver, back, err, c.want)
		}
	}
}
