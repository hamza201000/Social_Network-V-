//go:build cgo

package handler

import (
	"backend/internal/repository"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestGroupRoutesRejectNonMembers(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	files, _ := filepath.Glob("../../database/migrations/*.up.sql")
	sort.Strings(files)
	for _, file := range files {
		content, _ := os.ReadFile(file)
		if _, err := db.Exec(string(content)); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
	}
	_, err = db.Exec(`
		INSERT INTO users (id, first_name, last_name, username, email, password_hash) VALUES
			(1,'A','A','admin','a@x','x'), (2,'M','M','member','m@x','x'), (3,'O','O','outsider','o@x','x');
		INSERT INTO groups (id, admin_id, title, description_gp) VALUES (1, 1, 'Secret club', 'd');
		INSERT INTO members_gp (group_id, user_id) VALUES (1, 2);
	`)
	if err != nil {
		t.Fatal(err)
	}
	h := &UserHandler{UserRepo: repository.NewUserRepository(db)}

	call := func(fn http.HandlerFunc, method string, userID int64) int {
		req := httptest.NewRequest(method, "/api/groups/1/x", nil)
		req.SetPathValue("group_id", "1")
		req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
		rec := httptest.NewRecorder()
		fn(rec, req)
		return rec.Code
	}

	routes := map[string]http.HandlerFunc{
		"GET posts":   h.GetPostGp,
		"GET members": h.GetGroupMembers,
	}
	for name, fn := range routes {
		if code := call(fn, http.MethodGet, 3); code != http.StatusForbidden {
			t.Errorf("%s as outsider: got %d, want 403", name, code)
		}
		for _, member := range []int64{1, 2} {
			if code := call(fn, http.MethodGet, member); code != http.StatusOK {
				t.Errorf("%s as user %d: got %d, want 200", name, member, code)
			}
		}
	}
	if code := call(h.CreatePostGp, http.MethodPost, 3); code != http.StatusForbidden {
		t.Errorf("POST posts as outsider: got %d, want 403", code)
	}
	// the group page itself stays readable so outsiders can see the title and ask to join
	if code := call(h.GetGroup, http.MethodGet, 3); code != http.StatusOK {
		t.Errorf("GET group as outsider: got %d, want 200", code)
	}
}
