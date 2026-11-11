//go:build cgo

package repository

import (
	"backend/internal/models"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// openMigratedDB returns an in-memory database built from the real migration files.
func openMigratedDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	files, err := filepath.Glob("../../database/migrations/*.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			t.Fatalf("%s: %v", filepath.Base(file), err)
		}
	}
	return db
}

func TestGroupContentOnlyVisibleToMembers(t *testing.T) {
	db := openMigratedDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO users (id, first_name, last_name, username, email, password_hash) VALUES
			(1,'A','A','admin','a@x','x'), (2,'M','M','member','m@x','x'), (3,'O','O','outsider','o@x','x');
		INSERT INTO groups (id, admin_id, title, description_gp) VALUES (1, 1, 'Secret club', 'd');
		INSERT INTO members_gp (group_id, user_id) VALUES (1, 2);
		-- post 1: group post by the member, with an image
		INSERT INTO posts (id, user_id, content, image_path, group_id) VALUES (1, 2, 'group only', 'group.jpg', 1);
		-- post 2: normal public post by the member
		INSERT INTO posts (id, user_id, content, image_path, group_id) VALUES (2, 2, 'hello world', 'public.jpg', 0);
		INSERT INTO comments (post_id, user_id, content, image_path) VALUES (1, 1, 'nice', 'comment.jpg');
	`)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewUserRepository(db)

	for _, tc := range []struct {
		viewer int64
		want   bool
	}{{1, true}, {2, true}, {3, false}} {
		if got, err := repo.CanViewPost(1, tc.viewer); err != nil || got != tc.want {
			t.Errorf("CanViewPost(group post, user %d) = %v, %v; want %v", tc.viewer, got, err, tc.want)
		}
		for _, img := range []string{"group.jpg", "comment.jpg"} {
			if got, err := repo.CanViewImage(img, tc.viewer); err != nil || got != tc.want {
				t.Errorf("CanViewImage(%s, user %d) = %v, %v; want %v", img, tc.viewer, got, err, tc.want)
			}
		}
	}

	// normal public post and its image are still visible to everyone
	if ok, err := repo.CanViewPost(2, 3); err != nil || !ok {
		t.Errorf("outsider should see public post: %v %v", ok, err)
	}
	if ok, err := repo.CanViewImage("public.jpg", 3); err != nil || !ok {
		t.Errorf("outsider should see public image: %v %v", ok, err)
	}

	// the feed never contains group posts, even for members
	for _, viewer := range []int64{2, 3} {
		posts, err := repo.GetPosts(viewer)
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range posts {
			if p.Id == 1 {
				t.Errorf("feed of user %d contains the group post", viewer)
			}
		}
	}

	// the member's profile shows the group post to members only
	contents := func(viewer int64) string {
		posts, _ := repo.GetPostByUsername("member", viewer)
		var out []string
		for _, p := range posts {
			out = append(out, p.Content)
		}
		return strings.Join(out, ",")
	}
	if got := contents(3); strings.Contains(got, "group only") || !strings.Contains(got, "hello world") {
		t.Errorf("outsider sees on profile: %q", got)
	}
	if got := contents(1); !strings.Contains(got, "group only") {
		t.Errorf("admin should see the group post on profile: %q", got)
	}
}

func TestRegisterUsesDefaultAvatar(t *testing.T) {
	db := openMigratedDB(t)
	defer db.Close()
	repo := NewUserRepository(db)
	if err := repo.RegisterUser(models.RegisterRequest{FirstName: "a", LastName: "b", Username: "noavatar", Email: "n@x", Password: "x"}); err != nil {
		t.Fatal(err)
	}
	var avatar string
	if err := db.QueryRow(`SELECT avatar_url FROM users WHERE username = 'noavatar'`).Scan(&avatar); err != nil {
		t.Fatal(err)
	}
	if avatar != DefaultAvatar {
		t.Fatalf("avatar = %q, want %q", avatar, DefaultAvatar)
	}
}
