package repository

import (
	"backend/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"uuid"
)

func (r *UserRepository) CreateSession(userID int64) (uuid.UUID, error) {
	sessionID := uuid.New()

	expiresAt := time.Now().Add(24 * time.Minute)

	_, err := r.db.Exec(`
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, sessionID, userID, expiresAt)

	if err != nil {
		fmt.Println("error in create session",err)
		return uuid.Nil(), err
	}
	return sessionID, nil
}

func (repo *UserRepository) GetSession(token string) (*models.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at
		FROM sessions
		WHERE token = ?
		AND expires_at > CURRENT_TIMESTAMP
	`

	var session models.Session

	err := repo.db.QueryRow(query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, err
	}

	return &session, nil
}
