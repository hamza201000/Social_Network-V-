package repository

import (
	"backend/internal/models"
	"database/sql"
	"errors"
	"time"
	"uuid"
)

func (r *UserRepository) CreateSession(userID int64) (uuid.UUID, error) {
	sessionID := uuid.New()

	expiresAt := time.Now().Add(24 * time.Hour)

	_, err := r.db.Exec(`
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, sessionID, userID, expiresAt)

	if err != nil {
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
func (repo *UserRepository) DeleteSession(token string) error {
	query := `
		DELETE FROM sessions
		WHERE token = ?
	`

	_, err := repo.db.Exec(query, token)
	return err
}
