package repository

import (
	"backend/internal/models"
	"database/sql"
	"fmt"
)

func (repo *UserRepository) RegisterUser(user models.RegisterRequest) error {

	qeury := `INSERT INTO users
	(first_name,last_name,username,email,password_hash,date_of_birth,about_me)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	result, err := repo.db.Exec(qeury, user.FirstName, user.LastName, user.Username, user.Email, user.Password, user.DateOfBirth, user.AboutMe)
	if err != nil {
		fmt.Println("err database", err)
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Println("err in result database", err)
		return err
	}
	fmt.Println("userid:", id)
	return nil
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

	query := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.username,
			u.email,
			EXISTS (
          			SELECT 1
          			FROM followers f
          			WHERE f.follower_id = $2
            		AND f.following_target_id = u.id
          		) AS followed_by_me
		FROM users u
		WHERE u.username = $1
	`

	err := repo.db.QueryRow(
		query,
		username,
		user_id,
	).Scan(
		&user.Id,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
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
		SELECT id, first_name, last_name, username, email, avatar_url, date_of_birth, about_me, created_at
		FROM users
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
	)

	if err != nil {
		fmt.Println("error getting user:", err)
		return user, err
	}

	return user, nil
}
func (repo *UserRepository) FollowUser(follower_id int64, username string) error {
	
	query := `
		DELETE FROM followers
WHERE follower_id = $2
  AND following_target_id = (
      SELECT id
      FROM users
      WHERE username = $1
  )`
	_,err := repo.db.Exec(query, username, follower_id)
	if err == sql.ErrNoRows {
		_, err = repo.db.Exec(`INSERT INTO followers (follower_id, following_target_id) 
								SELECT $1, id FROM users WHERE username = $2`, follower_id, username)
	}
	if err != nil {
		fmt.Println("error following user:", err)
		return err
	}
	return nil
}
