package repository

import (
	"backend/internal/models"
	"database/sql"
	"errors"
)

var ErrNotGroupMember = errors.New("user is not a member of this group")

func (repo *UserRepository) isGroupMember(userID, groupID int64) (bool, error) {
	var isMember bool
	err := repo.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM groups g WHERE g.id = $1 AND g.admin_id = $2
			UNION ALL
			SELECT 1 FROM members_gp m WHERE m.group_id = $1 AND m.user_id = $2
		)
	`, groupID, userID).Scan(&isMember)
	return isMember, err
}

func (repo *UserRepository) CreateEvent(userID, groupID int64, event models.NewEvent) (int64, error) {
	member, err := repo.isGroupMember(userID, groupID)
	if err != nil {
		return 0, err
	}
	if !member {
		return 0, ErrNotGroupMember
	}
	result, err := repo.db.Exec(`
		INSERT INTO events (group_id, creator_id, title, description, event_at)
		VALUES ($1, $2, $3, $4, $5)
	`, groupID, userID, event.Title, event.Description, event.EventAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *UserRepository) GetGroupEvents(userID, groupID int64) ([]models.Event, error) {
	member, err := repo.isGroupMember(userID, groupID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, ErrNotGroupMember
	}
	rows, err := repo.db.Query(`
		SELECT e.id, e.group_id, e.creator_id, e.title, e.description, e.event_at, e.created_at,
			COALESCE(SUM(CASE WHEN er.response = 'Going' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN er.response = 'Not going' THEN 1 ELSE 0 END), 0),
			COALESCE(MAX(CASE WHEN er.user_id = $2 THEN er.response END), '')
		FROM events e
		LEFT JOIN event_responses er ON er.event_id = e.id
		WHERE e.group_id = $1
		GROUP BY e.id
		ORDER BY e.event_at ASC
	`, groupID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]models.Event, 0)
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(&event.Id, &event.GroupId, &event.CreatorId, &event.Title, &event.Description,
			&event.EventAt, &event.CreatedAt, &event.GoingCount, &event.NotGoingCount, &event.MyResponse); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (repo *UserRepository) VoteEvent(userID, groupID, eventID int64, response string) error {
	if response != "Going" && response != "Not going" {
		return errors.New("response must be Going or Not going")
	}
	member, err := repo.isGroupMember(userID, groupID)
	if err != nil {
		return err
	}
	if !member {
		return ErrNotGroupMember
	}
	result, err := repo.db.Exec(`
		INSERT INTO event_responses (event_id, user_id, response, updated_at)
		SELECT e.id, $1, $4, CURRENT_TIMESTAMP
		FROM events e
		WHERE e.id = $2 AND e.group_id = $3
		ON CONFLICT(event_id, user_id) DO UPDATE SET
			response = excluded.response,
			updated_at = CURRENT_TIMESTAMP
	`, userID, eventID, groupID, response)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}
