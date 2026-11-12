package repository

import (
	"backend/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken    = errors.New("this email is already registered")
	ErrUsernameTaken = errors.New("this username is already taken")
)

// DefaultAvatar is the image (in database/upload) used for users who didn't upload one.
const DefaultAvatar = "icon_user.png"

func (repo *UserRepository) RegisterUser(user models.RegisterRequest) error {
	if user.AvatarUrl == "" {
		user.AvatarUrl = DefaultAvatar
	}
	// never store the plain password, only its bcrypt hash (salt included)
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	transaction, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	var emailTaken, usernameTaken bool
	err = transaction.QueryRow(`SELECT
		EXISTS (SELECT 1 FROM users WHERE lower(email) = lower($1) OR lower(username) = lower($1)),
		EXISTS (SELECT 1 FROM users WHERE lower(username) = lower($2) OR lower(email) = lower($2))`,
		user.Email, user.Username).Scan(&emailTaken, &usernameTaken)
	if err != nil {
		return err
	}
	if emailTaken {
		return ErrEmailTaken
	}
	if user.Username == "" {
		// no nickname given: build a unique one from the name, e.g. mina_stone, mina_stone2
		user.Username, err = uniqueUsername(transaction, user.FirstName, user.LastName)
		if err != nil {
			return err
		}
	} else if usernameTaken {
		return ErrUsernameTaken
	}

	_, err = transaction.Exec(`INSERT INTO users
		(first_name, last_name, username, email, password_hash, date_of_birth, about_me, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.FirstName, user.LastName, user.Username, user.Email,
		user.Password, user.DateOfBirth, user.AboutMe, user.AvatarUrl)
	if err != nil {
		return err
	}
	return transaction.Commit()
}

func (repo *UserRepository) ImageReferenced(filename string) (bool, error) {
	var referenced bool
	err := repo.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM users WHERE avatar_url = ?
			UNION ALL SELECT 1 FROM posts WHERE image_path = ?
			UNION ALL SELECT 1 FROM comments WHERE image_path = ?
		)`, filename, filename, filename).Scan(&referenced)
	return referenced, err
}

// CanViewImage reports whether the viewer may load an uploaded file:
// avatars are public, post and comment images follow the visibility of their post.
func (repo *UserRepository) CanViewImage(filename string, viewerID int64) (bool, error) {
	var allowed bool
	err := repo.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM users WHERE avatar_url = $1
			UNION ALL
			SELECT 1 FROM posts p WHERE p.image_path = $1 AND `+postVisibleTo("$2")+`
			UNION ALL
			SELECT 1 FROM comments c JOIN posts p ON p.id = c.post_id
			WHERE c.image_path = $1 AND `+postVisibleTo("$2")+`
		)`, filename, viewerID).Scan(&allowed)
	return allowed, err
}

// ErrInvalidCredentials is returned when the login or the password is wrong.
var ErrInvalidCredentials = errors.New("invalid email/username or password")

// LoginUser finds the user by email or username and checks the password
// against the stored bcrypt hash.
func (repo *UserRepository) LoginUser(user models.LoginRequest) (*models.User, error) {
	var userData models.User
	err := repo.db.QueryRow(`SELECT id, password_hash FROM users WHERE email = $1 OR username = $1`,
		user.Email).Scan(&userData.Id, &userData.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(user.Password)) != nil {
		return nil, ErrInvalidCredentials
	}
	userData.Password = ""
	return &userData, nil
}
func (repo *UserRepository) GetUserBySession(sessionID string) (models.User, error) {

	var user models.User

	query := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.username,
			u.email
		FROM users u
		JOIN sessions s ON s.user_id = u.id
		WHERE s.session_id = $1
	`

	err := repo.db.QueryRow(
		query,
		sessionID,
	).Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
	)

	if err != nil {
		fmt.Println("error getting user:", err)
		return user, err
	}

	return user, nil
}
func (repo *UserRepository) DeleteSession(sessionID string) error {
	query := `
		DELETE FROM sessions
		WHERE token = $1
	`
	_, err := repo.db.Exec(query, sessionID)
	if err != nil {
		fmt.Println("error deleting session:", err)
		return err
	}

	return nil
}

func (repo *UserRepository) GetUserByUsername(user_id int64, username string) (models.User, error) {

	var user models.User
	var requested bool
	query := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.username,
			u.email,
			u.avatar_url,
			COALESCE(u.date_of_birth, ''),
			u.about_me,
			COALESCE(u.created_at, ''),
			u.private,
			EXISTS (
				SELECT 1 FROM followers f
				WHERE f.follower_id = $1 AND f.following_target_id = u.id
			) AS followed_by_me,
			EXISTS (
				SELECT 1 FROM follow_requests fr
				WHERE fr.requester_id = $1 AND fr.target_id = u.id
			) AS requested_by_me,
			(SELECT COUNT(*) FROM followers f WHERE f.following_target_id = u.id) AS followers_count,
			(SELECT COUNT(*) FROM followers f WHERE f.follower_id = u.id) AS following_count
		FROM users u
		WHERE u.username = $2
	`

	err := repo.db.QueryRow(
		query,
		user_id,
		username,
	).Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.AvatarUrl,
		&user.DateOfBirth,
		&user.AboutMe,
		&user.CreatedAt,
		&user.IsPrivate,
		&user.FollowedByMe,
		&requested,
		&user.Followers_count,
		&user.Following_count,
	)

	if err != nil {
		fmt.Println("error getting user:", err)
		return user, err
	}
	user.CanView = !user.IsPrivate || user.FollowedByMe || user.Id == user_id
	if !user.CanView {
		// a private profile only shows its information to followers
		user.Email = ""
		user.DateOfBirth = ""
		user.AboutMe = ""
		user.CreatedAt = ""
	}
	user.FollowStatus = "none"
	if user.FollowedByMe {
		user.FollowStatus = "following"
	} else if requested {
		user.FollowStatus = "requested"
	}

	return user, nil
}

func (repo *UserRepository) GetMyUserData(userID int64) (models.User, error) {

	var user models.User

	query := `
		SELECT u.id, u.first_name, u.last_name, u.username, u.email, u.avatar_url, u.date_of_birth, u.about_me, u.created_at, u.private,
		(
			SELECT COUNT(*)
			FROM followers f
			WHERE f.following_target_id = u.id
		) AS followers_count,
		 (
			SELECT COUNT(*)
			FROM followers f
			WHERE f.follower_id = u.id
		) AS following_count
		FROM users u
		WHERE id = $1
	`
	err := repo.db.QueryRow(
		query,
		userID,
	).Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.AvatarUrl,
		&user.DateOfBirth,
		&user.AboutMe,
		&user.CreatedAt,
		&user.IsPrivate,
		&user.Followers_count,
		&user.Following_count,
	)

	if err != nil {
		fmt.Println("error getting user:", err)
		return user, err
	}

	user.CanView = true

	return user, nil
}

var (
	ErrFollowSelf      = errors.New("you cannot follow yourself")
	ErrNoFollowRequest = errors.New("no pending follow request from this user")
)

// FollowUser is the follow button: it unfollows when already following,
// cancels a pending request, sends a request to a private profile, or follows
// a public profile directly. It returns the new status: "following", "requested" or "none".
func (repo *UserRepository) FollowUser(followerID int64, targetID int64) (string, error) {
	if followerID == targetID {
		return "", ErrFollowSelf
	}
	result, err := repo.db.Exec(`DELETE FROM followers WHERE follower_id = ?1 AND following_target_id = ?2`,
		followerID, targetID)
	if err != nil {
		return "", err
	}
	if affected, _ := result.RowsAffected(); affected > 0 {
		return "none", nil // was following -> unfollowed
	}

	result, err = repo.db.Exec(`DELETE FROM follow_requests WHERE requester_id = ?1 AND target_id = ?2`,
		followerID, targetID)
	if err != nil {
		return "", err
	}
	if affected, _ := result.RowsAffected(); affected > 0 {
		// request cancelled: drop the notification the target received for it
		_, err = repo.db.Exec(`DELETE FROM notifications WHERE user_id = ?1 AND actor_id = ?2 AND type = 'follow_request'`,
			targetID, followerID)
		return "none", err
	}

	var private bool
	if err := repo.db.QueryRow(`SELECT private FROM users WHERE id = ?1`, targetID).Scan(&private); err != nil {
		return "", err
	}
	if private {
		_, err = repo.db.Exec(`INSERT INTO follow_requests (requester_id, target_id) VALUES (?1, ?2)`,
			followerID, targetID)
		return "requested", err
	}
	_, err = repo.db.Exec(`INSERT INTO followers (follower_id, following_target_id) VALUES (?1, ?2)`,
		followerID, targetID)
	return "following", err
}

// RespondFollowRequest accepts or declines the request requesterID sent to targetID.
func (repo *UserRepository) RespondFollowRequest(targetID, requesterID int64, accept bool) error {
	transaction, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	result, err := transaction.Exec(`DELETE FROM follow_requests WHERE requester_id = ?1 AND target_id = ?2`,
		requesterID, targetID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNoFollowRequest
	}
	if accept {
		if _, err := transaction.Exec(`INSERT OR IGNORE INTO followers (follower_id, following_target_id) VALUES (?1, ?2)`,
			requesterID, targetID); err != nil {
			return err
		}
	}
	// the request is answered: its notification has no purpose anymore
	if _, err := transaction.Exec(`DELETE FROM notifications WHERE user_id = ?1 AND actor_id = ?2 AND type = 'follow_request'`,
		targetID, requesterID); err != nil {
		return err
	}
	return transaction.Commit()
}

func (repo *UserRepository) Followers(follower_id int64) ([]models.ProfileConnection, error) {
	var followers []models.ProfileConnection
	query := `
		SELECT f.follower_id,u.username,u.avatar_url
		FROM followers f
		JOIN users u ON f.follower_id=u.id
		WHERE f.following_target_id=$1
	`

	rows, err := repo.db.Query(query, follower_id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var follower models.ProfileConnection
		err := rows.Scan(&follower.User_id, &follower.Username, &follower.AvatarUrl)
		if err != nil {
			return nil, err
		}
		followers = append(followers, follower)
	}
	return followers, nil
}

func (repo *UserRepository) Following(follower_id int64) ([]models.ProfileConnection, error) {
	var followers []models.ProfileConnection
	query := `
		SELECT f.following_target_id,u.username,u.avatar_url
		FROM followers f
		JOIN users u ON f.following_target_id=u.id
		WHERE f.follower_id=$1
	`
	rows, err := repo.db.Query(query, follower_id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var follower models.ProfileConnection
		err := rows.Scan(&follower.User_id, &follower.Username, &follower.AvatarUrl)
		if err != nil {
			return nil, err
		}
		followers = append(followers, follower)
	}
	return followers, nil
}

func (repo *UserRepository) SearchUser(name_sh string) ([]models.User, error) {
	query := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.username,
			u.avatar_url,
			u.about_me,
			(SELECT COUNT(*) FROM followers f WHERE f.following_target_id = u.id) AS followers_count,
			(SELECT COUNT(*) FROM followers f WHERE f.follower_id = u.id) AS following_count
		FROM users u
		WHERE
			u.first_name LIKE '%' || $1 || '%'
			OR u.last_name LIKE '%' || $1 || '%'
			OR u.username LIKE '%' || $1 || '%'
			OR (u.first_name || ' ' || u.last_name) LIKE '%' || $1 || '%'
	`
	rows, err := repo.db.Query(query, name_sh)
	if err != nil {
		fmt.Println("error database:", err)
		return nil, err
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.Id,
			&user.FirstName,
			&user.LastName,
			&user.Username,
			&user.AvatarUrl,
			&user.AboutMe,
			&user.Followers_count,
			&user.Following_count,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// CanViewProfile reports whether viewerID may see targetID's profile details
// (followers, following, posts...). Public profiles are visible to everyone;
// private ones only to the owner and their followers.
func (repo *UserRepository) CanViewProfile(viewerID, targetID int64) (bool, error) {
	if viewerID == targetID {
		return true, nil
	}
	var ok bool
	err := repo.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM users u
			WHERE u.id = ?2
			  AND (
				u.private = 0
				OR EXISTS (
					SELECT 1 FROM followers f
					WHERE f.follower_id = ?1 AND f.following_target_id = u.id
				)
			  )
		)
	`, viewerID, targetID).Scan(&ok)
	return ok, err
}

// SetPrivacy makes the user's profile private (true) or public (false).
func (repo *UserRepository) SetPrivacy(userID int64, private bool) error {
	transaction, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if _, err := transaction.Exec(`UPDATE users SET private = ?1 WHERE id = ?2`, private, userID); err != nil {
		return err
	}
	if !private {
		// a public profile needs no approval: pending requests become follows
		if _, err := transaction.Exec(`
			INSERT OR IGNORE INTO followers (follower_id, following_target_id)
			SELECT requester_id, target_id FROM follow_requests WHERE target_id = ?1`, userID); err != nil {
			return err
		}
		if _, err := transaction.Exec(`DELETE FROM follow_requests WHERE target_id = ?1`, userID); err != nil {
			return err
		}
		if _, err := transaction.Exec(`DELETE FROM notifications WHERE user_id = ?1 AND type = 'follow_request'`, userID); err != nil {
			return err
		}
	}
	return transaction.Commit()
}

// UsernameTaken reports whether a username (nickname) is already used.
func (repo *UserRepository) UsernameTaken(username string) (bool, error) {
	var taken bool
	err := repo.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM users WHERE lower(username) = lower(?1))`, username).Scan(&taken)
	return taken, err
}

// uniqueUsername builds a free username from the first and last name, keeping
// only letters, digits and underscores, and adds a number if it is taken.
func uniqueUsername(tx *sql.Tx, firstName, lastName string) (string, error) {
	var b strings.Builder
	for _, r := range strings.ToLower(firstName + "_" + lastName) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	base := strings.Trim(b.String(), "_")
	if len(base) < 3 {
		base = "user"
	}
	if len(base) > 24 {
		base = base[:24]
	}
	candidate := base
	for i := 2; ; i++ {
		var taken bool
		err := tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM users WHERE lower(username) = $1 OR lower(email) = $1)`,
			candidate).Scan(&taken)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
		candidate = base + strconv.Itoa(i)
	}
}
