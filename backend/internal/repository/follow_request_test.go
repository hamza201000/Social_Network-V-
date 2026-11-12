//go:build cgo

package repository

import (
	"errors"
	"testing"
)

func isFollowing(t *testing.T, repo *UserRepository, follower, target int64) bool {
	t.Helper()
	var ok bool
	if err := repo.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM followers WHERE follower_id = ? AND following_target_id = ?)`,
		follower, target).Scan(&ok); err != nil {
		t.Fatal(err)
	}
	return ok
}

func TestFollowRequestsForPrivateProfiles(t *testing.T) {
	db := openMigratedDB(t)
	defer db.Close()
	if _, err := db.Exec(`
		INSERT INTO users (id, first_name, last_name, username, email, password_hash, private) VALUES
			(1,'A','A','alice','a@x','x',0), (2,'P','P','priv','p@x','x',1), (3,'Q','Q','pub','q@x','x',0);`); err != nil {
		t.Fatal(err)
	}
	repo := &UserRepository{db: db}

	if _, err := repo.FollowUser(1, 1); !errors.Is(err, ErrFollowSelf) {
		t.Fatalf("self follow: want ErrFollowSelf, got %v", err)
	}

	// public profile: followed directly
	if status, err := repo.FollowUser(1, 3); err != nil || status != "following" || !isFollowing(t, repo, 1, 3) {
		t.Fatalf("follow public: %q %v", status, err)
	}
	// calling again unfollows
	if status, err := repo.FollowUser(1, 3); err != nil || status != "none" || isFollowing(t, repo, 1, 3) {
		t.Fatalf("unfollow public: %q %v", status, err)
	}

	// private profile: only a request
	if status, err := repo.FollowUser(1, 2); err != nil || status != "requested" || isFollowing(t, repo, 1, 2) {
		t.Fatalf("follow private: %q %v", status, err)
	}
	profile, err := repo.GetUserByUsername(1, "priv")
	if err != nil || profile.FollowStatus != "requested" || profile.CanView {
		t.Fatalf("profile while requested: %+v %v", profile, err)
	}
	// declined: nothing happens, and answering twice fails
	if err := repo.RespondFollowRequest(2, 1, false); err != nil || isFollowing(t, repo, 1, 2) {
		t.Fatalf("decline: %v", err)
	}
	if err := repo.RespondFollowRequest(2, 1, true); !errors.Is(err, ErrNoFollowRequest) {
		t.Fatalf("answer twice: want ErrNoFollowRequest, got %v", err)
	}
	// new request, accepted
	repo.FollowUser(1, 2)
	if err := repo.RespondFollowRequest(2, 1, true); err != nil || !isFollowing(t, repo, 1, 2) {
		t.Fatalf("accept: %v", err)
	}
	profile, _ = repo.GetUserByUsername(1, "priv")
	if profile.FollowStatus != "following" || !profile.CanView {
		t.Fatalf("profile after accept: %+v", profile)
	}

	// going public accepts pending requests
	repo.FollowUser(3, 2)
	if err := repo.SetPrivacy(2, false); err != nil || !isFollowing(t, repo, 3, 2) {
		t.Fatalf("going public should accept pending requests: %v", err)
	}
}
