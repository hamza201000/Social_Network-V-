package repository

import (
	"backend/internal/models"
	"database/sql"
	"errors"
)

var (
	ErrAlreadyMember    = errors.New("already a member of the group")
	ErrAlreadyRequested = errors.New("join request already sent")
	ErrNoInvitation     = errors.New("no pending invitation")
	ErrNoJoinRequest    = errors.New("no pending join request")
)

// Join request outcomes.
const (
	JoinRequested = "requested" // a pending request was created
	JoinAccepted  = "joined"    // the user had an invitation, so they joined directly
)

type execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// addMember makes user_id a member and clears their invitations/requests for the group.
func addMember(db execer, group_id, user_id int64) error {
	if _, err := db.Exec(`
		INSERT INTO members_gp (group_id, user_id)
		SELECT $1, $2
		WHERE NOT EXISTS (SELECT 1 FROM members_gp WHERE group_id = $1 AND user_id = $2)
	`, group_id, user_id); err != nil {
		return err
	}
	return clearStatus(db, group_id, user_id)
}

func clearStatus(db execer, group_id, user_id int64) error {
	_, err := db.Exec(`DELETE FROM members_status WHERE group_id = $1 AND user_id = $2`, group_id, user_id)
	return err
}

// GroupAdminID returns the creator of the group (sql.ErrNoRows if it doesn't exist).
func (repo *UserRepository) GroupAdminID(group_id int64) (int64, error) {
	var admin int64
	err := repo.db.QueryRow(`SELECT admin_id FROM groups WHERE id = $1`, group_id).Scan(&admin)
	return admin, err
}

// IsGroupMember reports whether the user is the group admin or a member.
func (repo *UserRepository) IsGroupMember(user_id, group_id int64) (bool, error) {
	var ok bool
	err := repo.db.QueryRow(`
		SELECT EXISTS (SELECT 1 FROM groups WHERE id = $1 AND admin_id = $2)
		    OR EXISTS (SELECT 1 FROM members_gp WHERE group_id = $1 AND user_id = $2)
	`, group_id, user_id).Scan(&ok)
	return ok, err
}

// IsConnected reports whether the two users follow each other in either direction.
func (repo *UserRepository) IsConnected(a, b int64) (bool, error) {
	var ok bool
	err := repo.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM followers
			WHERE (follower_id = $1 AND following_target_id = $2)
			   OR (follower_id = $2 AND following_target_id = $1)
		)
	`, a, b).Scan(&ok)
	return ok, err
}

// InvitableUsers lists the followers and followees of user_id together with
// their current status in the group.
func (repo *UserRepository) InvitableUsers(user_id, group_id int64) ([]models.InvitableUser, error) {
	rows, err := repo.db.Query(`
		SELECT u.id, u.first_name, u.last_name, u.username, u.avatar_url,
			CASE
				WHEN EXISTS (SELECT 1 FROM groups g WHERE g.id = ?2 AND g.admin_id = u.id)
				  OR EXISTS (SELECT 1 FROM members_gp m WHERE m.group_id = ?2 AND m.user_id = u.id)
					THEN 'member'
				ELSE COALESCE((
					SELECT s.member_status FROM members_status s
					WHERE s.group_id = ?2 AND s.user_id = u.id AND s.member_status IN ('invited', 'pending')
					ORDER BY s.id DESC LIMIT 1
				), 'none')
			END AS status
		FROM users u
		WHERE u.id != ?1 AND u.id IN (
			SELECT follower_id FROM followers WHERE following_target_id = ?1
			UNION
			SELECT following_target_id FROM followers WHERE follower_id = ?1
		)
		ORDER BY u.username
	`, user_id, group_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.InvitableUser{}
	for rows.Next() {
		var u models.InvitableUser
		if err := rows.Scan(&u.UserId, &u.FirstName, &u.LastName, &u.Username, &u.AvatarUrl, &u.Status); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// InviteUsrGp records an invitation from inviter_id for user_id to join group_id.
// It returns false if the user is already a member or already has an
// invitation or join request.
func (repo *UserRepository) InviteUsrGp(inviter_id, user_id, group_id int64) (bool, error) {
	result, err := repo.db.Exec(`
		INSERT INTO members_status (group_id, user_id, member_status, invited_by)
		SELECT $1, $2, 'invited', $3
		WHERE NOT EXISTS (SELECT 1 FROM groups WHERE id = $1 AND admin_id = $2)
		  AND NOT EXISTS (SELECT 1 FROM members_gp WHERE group_id = $1 AND user_id = $2)
		  AND NOT EXISTS (
			SELECT 1 FROM members_status
			WHERE group_id = $1 AND user_id = $2 AND member_status IN ('invited', 'pending')
		  )
	`, group_id, user_id, inviter_id)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}

// RequestToJoin creates a pending join request. If the user was already
// invited, the invitation is accepted instead and JoinAccepted is returned
// together with the inviter's id.
func (repo *UserRepository) RequestToJoin(user_id, group_id int64) (outcome string, inviterID int64, err error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return "", 0, err
	}
	defer tx.Rollback()

	var member bool
	if err := tx.QueryRow(`
		SELECT EXISTS (SELECT 1 FROM groups WHERE id = $1 AND admin_id = $2)
		    OR EXISTS (SELECT 1 FROM members_gp WHERE group_id = $1 AND user_id = $2)
	`, group_id, user_id).Scan(&member); err != nil {
		return "", 0, err
	}
	if member {
		return "", 0, ErrAlreadyMember
	}

	var status string
	var invitedBy sql.NullInt64
	err = tx.QueryRow(`
		SELECT member_status, invited_by FROM members_status
		WHERE group_id = $1 AND user_id = $2 AND member_status IN ('invited', 'pending')
		ORDER BY id DESC LIMIT 1
	`, group_id, user_id).Scan(&status, &invitedBy)
	switch {
	case err == nil && status == "pending":
		return "", 0, ErrAlreadyRequested
	case err == nil && status == "invited":
		if err := addMember(tx, group_id, user_id); err != nil {
			return "", 0, err
		}
		return JoinAccepted, invitedBy.Int64, tx.Commit()
	case err != sql.ErrNoRows:
		return "", 0, err
	}

	if _, err := tx.Exec(`
		INSERT INTO members_status (group_id, user_id, member_status) VALUES ($1, $2, 'pending')
	`, group_id, user_id); err != nil {
		return "", 0, err
	}
	return JoinRequested, 0, tx.Commit()
}

// RespondInvitation accepts or declines the invitation of user_id to group_id
// and returns who sent it (0 if unknown).
func (repo *UserRepository) RespondInvitation(user_id, group_id int64, accept bool) (int64, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var invitedBy sql.NullInt64
	err = tx.QueryRow(`
		SELECT invited_by FROM members_status
		WHERE group_id = $1 AND user_id = $2 AND member_status = 'invited'
		ORDER BY id DESC LIMIT 1
	`, group_id, user_id).Scan(&invitedBy)
	if err == sql.ErrNoRows {
		return 0, ErrNoInvitation
	}
	if err != nil {
		return 0, err
	}
	if accept {
		err = addMember(tx, group_id, user_id)
	} else {
		err = clearStatus(tx, group_id, user_id)
	}
	if err != nil {
		return 0, err
	}
	return invitedBy.Int64, tx.Commit()
}

// RespondJoinRequest accepts or refuses the pending join request of user_id.
func (repo *UserRepository) RespondJoinRequest(group_id, user_id int64, accept bool) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM members_status
			WHERE group_id = $1 AND user_id = $2 AND member_status = 'pending'
		)
	`, group_id, user_id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNoJoinRequest
	}
	if accept {
		err = addMember(tx, group_id, user_id)
	} else {
		err = clearStatus(tx, group_id, user_id)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GroupMembers lists the admin and all members of a group, admin first.
func (repo *UserRepository) GroupMembers(group_id int64) ([]models.GroupMember, error) {
	return repo.queryMembers(`
		SELECT u.id, u.first_name, u.last_name, u.username, u.avatar_url, 'admin', g.created_at
		FROM groups g JOIN users u ON u.id = g.admin_id
		WHERE g.id = $1
		UNION ALL
		SELECT u.id, u.first_name, u.last_name, u.username, u.avatar_url, 'member', m.created_at
		FROM members_gp m
		JOIN groups g ON g.id = m.group_id
		JOIN users u ON u.id = m.user_id
		WHERE m.group_id = $1 AND m.user_id != g.admin_id
		ORDER BY 6, 7
	`, group_id)
}

// JoinRequests lists the users waiting for the admin to accept them, oldest first.
func (repo *UserRepository) JoinRequests(group_id int64) ([]models.GroupMember, error) {
	return repo.queryMembers(`
		SELECT u.id, u.first_name, u.last_name, u.username, u.avatar_url, '', MIN(s.created_at)
		FROM members_status s JOIN users u ON u.id = s.user_id
		WHERE s.group_id = $1 AND s.member_status = 'pending'
		GROUP BY u.id
		ORDER BY 7
	`, group_id)
}

func (repo *UserRepository) queryMembers(query string, args ...any) ([]models.GroupMember, error) {
	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := []models.GroupMember{}
	for rows.Next() {
		var m models.GroupMember
		if err := rows.Scan(&m.UserId, &m.FirstName, &m.LastName, &m.Username, &m.AvatarUrl, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
