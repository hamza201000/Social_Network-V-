//go:build cgo

package handler

import (
	"backend/internal/models"
	"backend/internal/repository"
	chatws "backend/internal/websocket"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	gorilla "github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

func TestGroupMembershipFlow(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	files, _ := filepath.Glob("../../database/migrations/*.up.sql")
	sort.Strings(files)
	for _, f := range files {
		sqlText, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(string(sqlText)); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	for i, name := range []string{"alice", "bob", "carol"} {
		_, err := db.Exec(`INSERT INTO users (id, first_name, last_name, username, email, password_hash, date_of_birth)
			VALUES ($1, $2, 'x', $2, $3, 'h', '2000-01-01')`, i+1, name, name+"@x.io")
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO followers (follower_id, following_target_id) VALUES (2, 1)`); err != nil {
		t.Fatal(err)
	}

	repo := repository.NewUserRepository(db)
	hub := chatws.NewHub()
	h := UserHandler{UserRepo: repo, Hub: hub}
	groupID, err := repo.CreateGroup(1, models.NewGroupDetails{Title: "Hikers", Description: "d"})
	if err != nil {
		t.Fatal(err)
	}
	gid := strconv.FormatInt(groupID, 10)

	upgrader := gorilla.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.URL.Query().Get("user"), 10, 64)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err == nil {
			hub.Register(id, conn)
		}
	}))
	defer server.Close()
	sockets := map[int64]*gorilla.Conn{}
	for _, id := range []int64{1, 2, 3} {
		c, _, err := gorilla.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/?user="+strconv.FormatInt(id, 10), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		sockets[id] = c
	}
	for i := 0; i < 50 && !(hub.IsOnline(1) && hub.IsOnline(2) && hub.IsOnline(3)); i++ {
		time.Sleep(10 * time.Millisecond)
	}

	call := func(as int64, fn http.HandlerFunc, path map[string]string, body string) int {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		r = r.WithContext(context.WithValue(r.Context(), "user_id", as))
		for k, v := range path {
			r.SetPathValue(k, v)
		}
		w := httptest.NewRecorder()
		fn(w, r)
		return w.Code
	}
	expect := func(user int64, msgType, kind string) {
		t.Helper()
		c := sockets[user]
		c.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			var msg struct {
				Type         string              `json:"type"`
				Action       string              `json:"action"`
				Notification models.Notification `json:"notification"`
			}
			_, data, err := c.ReadMessage()
			if err != nil {
				t.Fatalf("user %d never received %s/%s: %v", user, msgType, kind, err)
			}
			json.Unmarshal(data, &msg)
			got := msg.Action
			if msg.Type == "notification" {
				got = msg.Notification.Type
			}
			if msg.Type == msgType && got == kind {
				return
			}
		}
	}
	g := map[string]string{"group_id": gid}

	if code := call(1, h.InviteUsrGp, g, `{"user_id":3}`); code != http.StatusForbidden {
		t.Fatalf("invite stranger: %d", code)
	}
	if code := call(1, h.InviteUsrGp, g, `{"user_id":2}`); code != http.StatusOK {
		t.Fatalf("invite bob: %d", code)
	}
	expect(2, "notification", models.NotifGroupInvite)
	expect(2, "group_update", "invited")
	if code := call(1, h.InviteUsrGp, g, `{"user_id":2}`); code != http.StatusConflict {
		t.Fatalf("invite twice: %d", code)
	}

	if code := call(2, h.RespondInvitation, g, `{"accept":true}`); code != http.StatusOK {
		t.Fatalf("bob accepts: %d", code)
	}
	expect(1, "notification", models.NotifInviteAccepted)
	expect(1, "group_update", "member_joined")

	if code := call(3, h.JoinUsrGroup, g, ``); code != http.StatusOK {
		t.Fatalf("carol joins: %d", code)
	}
	expect(1, "notification", models.NotifJoinRequest) // only the creator is told
	if code := call(3, h.JoinUsrGroup, g, ``); code != http.StatusConflict {
		t.Fatalf("carol joins twice: %d", code)
	}
	req := map[string]string{"group_id": gid, "user_id": "3"}

	if code := call(3, h.RespondJoinRequest, req, `{"accept":true}`); code != http.StatusForbidden {
		t.Fatalf("non-member answering request: %d", code)
	}
	if code := call(3, h.GetJoinRequests, g, ``); code != http.StatusForbidden {
		t.Fatalf("non-member listing requests: %d", code)
	}
	if code := call(2, h.GetJoinRequests, g, ``); code != http.StatusForbidden {
		t.Fatalf("member (not creator) listing requests: %d", code)
	}
	if code := call(2, h.RespondJoinRequest, req, `{"accept":true}`); code != http.StatusForbidden {
		t.Fatalf("member (not creator) answering request: %d", code)
	}
	if code := call(1, h.GetJoinRequests, g, ``); code != http.StatusOK {
		t.Fatalf("creator listing requests: %d", code)
	}

	if code := call(1, h.RespondJoinRequest, req, `{"accept":false}`); code != http.StatusOK {
		t.Fatalf("creator refusing: %d", code)
	}
	expect(3, "notification", models.NotifJoinRefused)
	call(3, h.JoinUsrGroup, g, ``)
	if code := call(1, h.RespondJoinRequest, req, `{"accept":true}`); code != http.StatusOK {
		t.Fatalf("creator accepting: %d", code)
	}
	expect(3, "notification", models.NotifJoinAccepted)
	expect(2, "group_update", "member_joined") // other members see the new member live

	if code := call(1, h.RespondJoinRequest, req, `{"accept":true}`); code != http.StatusNotFound {
		t.Fatalf("answering twice: %d", code)
	}

	members, _ := repo.GroupMembers(groupID)
	if len(members) != 3 {
		t.Fatalf("members: %+v", members)
	}
}
