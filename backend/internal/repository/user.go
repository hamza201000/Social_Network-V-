package repository

import (
	"backend/internal/models"
	"database/sql"
	"fmt"
)

func (repo *UserRepository) RegisterUser(user models.RegisterRequest) error {
	transaction, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

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

func (repo *UserRepository) CanViewImage(filename string, viewerID int64) (bool, error) {
	var allowed bool
	err := repo.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM users WHERE avatar_url = ?
			UNION ALL
			SELECT 1 FROM posts p
			WHERE p.image_path = ? AND (
				p.group_id != 0 OR p.privacy = 'public' OR p.user_id = ? OR
				(p.privacy = 'almost_private' AND EXISTS (
					SELECT 1 FROM followers f WHERE f.follower_id = ? AND f.following_target_id = p.user_id
				)) OR
				(p.privacy = 'private' AND EXISTS (
					SELECT 1 FROM post_allowed_users a WHERE a.post_id = p.id AND a.user_id = ?
				))
			)
			UNION ALL
			SELECT 1 FROM comments c
			JOIN posts p ON p.id = c.post_id
			WHERE c.image_path = ? AND (
				p.privacy = 'public' OR p.user_id = ? OR
				(p.privacy = 'almost_private' AND EXISTS (
					SELECT 1 FROM followers f WHERE f.follower_id = ? AND f.following_target_id = p.user_id
				)) OR
				(p.privacy = 'private' AND EXISTS (
					SELECT 1 FROM post_allowed_users a WHERE a.post_id = p.id AND a.user_id = ?
				))
			)
		)`,
		filename,
		filename, viewerID, viewerID, viewerID,
		filename, viewerID, viewerID, viewerID,
	).Scan(&allowed)
	return allowed, err
}

func (repo *UserRepository) LoginUser(user models.LoginRequest) (*models.User, error) {

	var userData models.User
	qeury := `SELECT id,password_hash
	FROM users
	WHERE (email=$1 or username=$1)
	`
	err := repo.db.QueryRow(qeury, user.Email).Scan(&userData.Id, &userData.Password)
	if err == sql.ErrNoRows {
		fmt.Println("err database", err)
		return nil, err
	}

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
		WHERE session_id = $1
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
	fmt.Println("user_id", user_id)
	query := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.username,
			u.email,
			u.avatar_url,
			EXISTS (
          			SELECT 1
          			FROM followers f
          			WHERE f.follower_id = $1
            		AND f.following_target_id = u.id
          		) AS followed_by_me
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
		&user.FollowedByMe,
	)

	if err != nil {
		fmt.Println("error getting user:", err)
		return user, err
	}

	return user, nil
}

func (repo *UserRepository) GetMyUserData(userID int64) (models.User, error) {

	var user models.User

	query := `
		SELECT u.id, u.first_name, u.last_name, u.username, u.email, u.avatar_url, u.date_of_birth, u.about_me, u.created_at,
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
		&user.Followers_count,
		&user.Following_count,
	)

	if err != nil {
		fmt.Println("error getting user:", err)
		return user, err
	}

	return user, nil
}

func (repo *UserRepository) FollowUser(followerID int64, followingTargetID int64) error {
	query := `
		DELETE FROM followers
		WHERE follower_id = $1 AND following_target_id = $2;
	`

	result, err := repo.db.Exec(query, followerID, followingTargetID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		_, err = repo.db.Exec(`
			INSERT INTO followers (follower_id, following_target_id)
			VALUES ($1, $2)
		`, followerID, followingTargetID)

		if err != nil {
			return err
		}
	}

	return nil
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
