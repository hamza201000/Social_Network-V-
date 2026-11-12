//go:build cgo

package repository

import (
	"backend/internal/models"
	"errors"
	"testing"
)

func TestPasswordIsHashedAndCheckedOnLogin(t *testing.T) {
	db := openMigratedDB(t)
	defer db.Close()
	repo := &UserRepository{db: db}

	err := repo.RegisterUser(models.RegisterRequest{
		FirstName: "A", LastName: "B", Username: "alice", Email: "alice@x.com",
		Password: "secret123", DateOfBirth: "2000-01-01",
	})
	if err != nil {
		t.Fatal(err)
	}

	dup := models.RegisterRequest{FirstName: "C", LastName: "D", Username: "bob", Email: "ALICE@x.com", Password: "secret123", DateOfBirth: "2000-01-01"}
	if err := repo.RegisterUser(dup); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("duplicate email: want ErrEmailTaken, got %v", err)
	}
	dup.Email, dup.Username = "bob@x.com", "Alice"
	if err := repo.RegisterUser(dup); !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("duplicate username: want ErrUsernameTaken, got %v", err)
	}

	var stored string
	if err := db.QueryRow(`SELECT password_hash FROM users WHERE username = 'alice'`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == "secret123" || len(stored) < 50 {
		t.Fatalf("password stored in plain text: %q", stored)
	}

	for _, login := range []string{"alice@x.com", "alice"} {
		user, err := repo.LoginUser(models.LoginRequest{Email: login, Password: "secret123"})
		if err != nil || user.Id == 0 || user.Password != "" {
			t.Fatalf("login with %q: %+v %v", login, user, err)
		}
	}
	if _, err := repo.LoginUser(models.LoginRequest{Email: "alice", Password: "wrong-pass"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password: want ErrInvalidCredentials, got %v", err)
	}
	if _, err := repo.LoginUser(models.LoginRequest{Email: "nobody", Password: "secret123"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("unknown user: want ErrInvalidCredentials, got %v", err)
	}
}
