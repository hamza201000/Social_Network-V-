package repository

import (
	"backend/internal/models"
	"database/sql"
	"errors"
)

var ErrMessageNotAllowed = errors.New("users are not connected by a follow relationship")

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (repo *MessageRepository) CanMessage(senderID, receiverID int64) (bool, error) {
	var allowed bool
	err := repo.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM followers
			WHERE (follower_id = $1 AND following_target_id = $2)
			   OR (follower_id = $2 AND following_target_id = $1)
		)
	`, senderID, receiverID).Scan(&allowed)
	return allowed, err
}

func (repo *MessageRepository) InsertMessage(senderID, receiverID int64, text string) (models.DataMessage, error) {
	result, err := repo.db.Exec(`
		INSERT INTO messages (sender_id, receiver_id, message)
		VALUES ($1, $2, $3)
	`, senderID, receiverID, text)
	if err != nil {
		return models.DataMessage{}, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return models.DataMessage{}, err
	}
	return repo.getMessageByID(messageID)
}

func (repo *MessageRepository) GetMessagesBetweenUsers(userID, otherUserID int64) ([]models.DataMessage, error) {
	rows, err := repo.db.Query(`
		SELECT m.id, m.sender_id, m.receiver_id, m.message, m.seen, m.created_at, u.username
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE (m.sender_id = $1 AND m.receiver_id = $2)
		   OR (m.sender_id = $2 AND m.receiver_id = $1)
		ORDER BY m.created_at ASC, m.id ASC
	`, userID, otherUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]models.DataMessage, 0)
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (repo *MessageRepository) GetLatestMessage(userID, otherUserID int64) (models.DataMessage, error) {
	return scanMessage(repo.db.QueryRow(`
		SELECT m.id, m.sender_id, m.receiver_id, m.message, m.seen, m.created_at, u.username
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE (m.sender_id = $1 AND m.receiver_id = $2)
		   OR (m.sender_id = $2 AND m.receiver_id = $1)
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT 1
	`, userID, otherUserID))
}

func (repo *MessageRepository) GetChatUsers(userID int64) ([]models.Client, error) {
	rows, err := repo.db.Query(`
		WITH chat_users AS (
			SELECT following_target_id AS other_user_id
			FROM followers
			WHERE follower_id = $1
			UNION
			SELECT follower_id AS other_user_id
			FROM followers
			WHERE following_target_id = $1
			UNION
			SELECT CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END AS other_user_id
			FROM messages
			WHERE sender_id = $1 OR receiver_id = $1
		),
		latest_messages AS (
			SELECT
				CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END AS other_user_id,
				MAX(id) AS message_id
			FROM messages
			WHERE sender_id = $1 OR receiver_id = $1
			GROUP BY other_user_id
		)
		SELECT u.id, u.username,
			m.id, m.sender_id, m.receiver_id, m.message, m.seen, m.created_at, sender.username
		FROM chat_users chats
		JOIN users u ON u.id = chats.other_user_id
		LEFT JOIN latest_messages latest ON latest.other_user_id = chats.other_user_id
		LEFT JOIN messages m ON m.id = latest.message_id
		LEFT JOIN users sender ON sender.id = m.sender_id
		ORDER BY CASE WHEN m.id IS NULL THEN 1 ELSE 0 END, m.created_at DESC, m.id DESC, u.username
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chats := make([]models.Client, 0)
	for rows.Next() {
		var chat models.Client
		var messageID, senderID, receiverID sql.NullInt64
		var text, senderName sql.NullString
		var seen sql.NullBool
		var createdAt sql.NullTime
		if err := rows.Scan(
			&chat.UserID,
			&chat.Username,
			&messageID,
			&senderID,
			&receiverID,
			&text,
			&seen,
			&createdAt,
			&senderName,
		); err != nil {
			return nil, err
		}
		if messageID.Valid {
			chat.LastMsg = &models.DataMessage{
				ID:             messageID.Int64,
				Type:           "message",
				SenderID:       senderID.Int64,
				ReceiverID:     receiverID.Int64,
				Message:        text.String,
				Seen:           seen.Bool,
				CreatedAt:      createdAt.Time,
				UsernameSender: senderName.String,
			}
		}
		chats = append(chats, chat)
	}
	return chats, rows.Err()
}

func (repo *MessageRepository) MarkMessageSeen(messageID, receiverID int64) error {
	result, err := repo.db.Exec(`
		UPDATE messages SET seen = 1
		WHERE id = $1 AND receiver_id = $2
	`, messageID, receiverID)
	if err != nil {
		return err
	}

	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated > 0 {
		return nil
	}

	var exists bool
	err = repo.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM messages WHERE id = $1 AND receiver_id = $2
		)
	`, messageID, receiverID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *MessageRepository) getMessageByID(messageID int64) (models.DataMessage, error) {
	return scanMessage(repo.db.QueryRow(`
		SELECT m.id, m.sender_id, m.receiver_id, m.message, m.seen, m.created_at, u.username
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.id = $1
	`, messageID))
}

type messageScanner interface {
	Scan(dest ...any) error
}

func scanMessage(scanner messageScanner) (models.DataMessage, error) {
	var message models.DataMessage
	err := scanner.Scan(
		&message.ID,
		&message.SenderID,
		&message.ReceiverID,
		&message.Message,
		&message.Seen,
		&message.CreatedAt,
		&message.UsernameSender,
	)
	message.Type = "message"
	return message, err
}
