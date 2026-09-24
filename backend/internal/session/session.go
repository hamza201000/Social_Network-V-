package session

import (
	"backend/internal/models"
	"errors"
	"sync"
	"uuid"
)

var (
	sessions = make(map[uuid.UUID]*models.Session)
	mu       sync.RWMutex
)

func CreateSession(userID int) (*models.Session, error) {
	sessionID := uuid.New()
	session := &models.Session{
		ID:     sessionID,
		UserID: userID,
	}
	return session, nil
}

func Get(sessionID uuid.UUID) (*models.Session, error) {
	mu.RLock()
	session, exists := sessions[sessionID]
	mu.RUnlock()

	if !exists {
		return nil, errors.New("session not found")
	}

	return session, nil
}

func Delete(sessionID uuid.UUID) error {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := sessions[sessionID]; !exists {
		return errors.New("session not found")
	}

	delete(sessions, sessionID)

	return nil
}
