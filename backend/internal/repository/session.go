package repository

import (
	"backend/internal/models"

	"fmt"
	"time"
	"uuid"
)

// SessionDuration is how long a login lasts. Users stay logged in until they
// log out; this only cleans up sessions that were abandoned long ago.
const SessionDuration = 30 * 24 * time.Hour

func (r *UserRepository) CreateSession(userID int64) (uuid.UUID, error) {
	sessionID := uuid.New()

	expiresAt := time.Now().UTC().Add(SessionDuration)

	_, err := r.db.Exec(`
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, sessionID, userID, expiresAt)

	if err != nil {
		fmt.Println("error in create session", err)
		return uuid.Nil(), err
	}
	return sessionID, nil
}

func (repo *UserRepository) GetSession(token string) (*models.Session, error) {
	query := `
		SELECT id, user_id, expires_at
		FROM sessions 
		WHERE token = ?
	`

	var session models.Session

	err := repo.db.QueryRow(query, token).Scan(
		&session.ID,
		&session.User_id,
		&session.ExpiresAt,
	)

	if err != nil {
		fmt.Println("error", err)
		return nil, err
	}

	return &session, nil
}
