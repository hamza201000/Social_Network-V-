package repository

import (
	"backend/internal/models"
	"database/sql"
)

const notificationSelect = `
	SELECT n.id, n.user_id, n.actor_id, u.username, u.avatar_url,
		n.type, n.group_id, COALESCE(g.title, ''), n.event_id,
		n.message, n.is_read, n.created_at
	FROM notifications n
	JOIN users u ON u.id = n.actor_id
	LEFT JOIN groups g ON g.id = n.group_id
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanNotification(row rowScanner) (models.Notification, error) {
	var n models.Notification
	var groupID, eventID sql.NullInt64
	err := row.Scan(&n.ID, &n.UserID, &n.ActorID, &n.ActorUsername, &n.ActorAvatar,
		&n.Type, &groupID, &n.GroupTitle, &eventID,
		&n.Message, &n.IsRead, &n.CreatedAt)
	if groupID.Valid {
		n.GroupID = &groupID.Int64
	}
	if eventID.Valid {
		n.EventID = &eventID.Int64
	}
	return n, err
}

func nullID(id int64) sql.NullInt64 {
	return sql.NullInt64{Int64: id, Valid: id > 0}
}

// CreateNotification stores a notification and returns it with actor and group details.
// groupID and eventID can be 0 when they don't apply.
func (repo *UserRepository) CreateNotification(userID, actorID int64, notifType string, groupID, eventID int64, message string) (models.Notification, error) {
	result, err := repo.db.Exec(`
		INSERT INTO notifications (user_id, actor_id, type, group_id, event_id, message)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, actorID, notifType, nullID(groupID), nullID(eventID), message)
	if err != nil {
		return models.Notification{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return models.Notification{}, err
	}
	return scanNotification(repo.db.QueryRow(notificationSelect+` WHERE n.id = $1`, id))
}

// GetNotifications returns the latest notifications of a user and how many are unread.
func (repo *UserRepository) GetNotifications(userID int64, limit int) (models.NotificationList, error) {
	list := models.NotificationList{Notifications: []models.Notification{}}

	rows, err := repo.db.Query(notificationSelect+`
		WHERE n.user_id = $1
		ORDER BY n.created_at DESC, n.id DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return list, err
		}
		list.Notifications = append(list.Notifications, n)
	}
	if err := rows.Err(); err != nil {
		return list, err
	}

	err = repo.db.QueryRow(`
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = 0
	`, userID).Scan(&list.UnreadCount)
	return list, err
}

// MarkNotificationRead marks one notification as read. Returns sql.ErrNoRows
// if it doesn't exist or belongs to another user.
func (repo *UserRepository) MarkNotificationRead(notificationID, userID int64) error {
	result, err := repo.db.Exec(`
		UPDATE notifications SET is_read = 1 WHERE id = $1 AND user_id = $2
	`, notificationID, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *UserRepository) MarkAllNotificationsRead(userID int64) error {
	_, err := repo.db.Exec(`
		UPDATE notifications SET is_read = 1 WHERE user_id = $1 AND is_read = 0
	`, userID)
	return err
}

// GroupMemberIDs returns the admin and all members of a group.
func (repo *UserRepository) GroupMemberIDs(groupID int64) ([]int64, error) {
	rows, err := repo.db.Query(`
		SELECT admin_id FROM groups WHERE id = $1
		UNION
		SELECT user_id FROM members_gp WHERE group_id = $1
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
