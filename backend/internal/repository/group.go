package repository

import (
	"backend/internal/models"
	"fmt"
)

const globalQuery = `
		SELECT g.id, g.title, g.description_gp, g.admin_id, u.username, COUNT(m.user_id) AS member_count
		FROM groups g
		JOIN users u ON g.admin_id = u.id
		LEFT JOIN members_gp m ON g.id = m.group_id
		
		`

func (repo *UserRepository) CreateGroup(Admin_id int64, newGroup models.NewGroupDetails) (int64, error) {

	query := `
		INSERT INTO groups (title, description_gp, admin_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var groupID int64
	err := repo.db.QueryRow(query, newGroup.Title, newGroup.Description, Admin_id).Scan(&groupID)
	if err != nil {
		fmt.Println("Error creating group in database:", err)
		return 0, err
	}
	return groupID, nil
}

func (repo *UserRepository) GetGroups(user_id int64) ([]models.Group, error) {
	var groups []models.Group
	// query := `
	// 	SELECT g.id, g.title, g.description_gp, g.admin_id, u.username, COUNT(m.user_id) AS member_count
	// 	FROM groups g
	// 	JOIN users u ON g.admin_id = u.id
	// 	LEFT JOIN members_gp m ON g.id = m.group_id
	// 	GROUP BY g.id, g.title, g.description_gp, g.admin_id, u.username
	// 	`
	//WHERE (g.admin_id = $1 OR g.id = (SELECT group_id FROM members_gp WHERE user_id = $1 ) OR )
	rows, err := repo.db.Query(globalQuery+"GROUP BY g.id, g.title, g.description_gp, g.admin_id, u.username", user_id)
	if err != nil {
		fmt.Println("Error retrieving groups from database:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var group models.Group
		err := rows.Scan(&group.Id, &group.Title, &group.Description, &group.AdminId, &group.AdminName, &group.MemberCount)
		if err != nil {
			fmt.Println("Error retrieving group row from database:", err)
			return nil, err
		}

		if group.AdminId == user_id {
			group.AdminName = "You"
		}
		groups = append(groups, group)
	}
	return groups, nil
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
				WHEN EXISTS (SELECT 1 FROM groups g WHERE g.id = $2 AND g.admin_id = u.id)
				  OR EXISTS (SELECT 1 FROM members_gp m WHERE m.group_id = $2 AND m.user_id = u.id)
					THEN 'member'
				ELSE COALESCE((
					SELECT s.member_status FROM members_status s
					WHERE s.group_id = $2 AND s.user_id = u.id
					ORDER BY s.created_at DESC, s.id DESC LIMIT 1
				), 'none')
			END AS status
		FROM users u
		WHERE u.id != $1 AND u.id IN (
			SELECT follower_id FROM followers WHERE following_target_id = $1
			UNION
			SELECT following_target_id FROM followers WHERE follower_id = $1
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

// InviteUsrGp records an invitation for user_id to join group_id.
// It returns false if the user is already a member or already has an
// invitation or join request.
func (repo *UserRepository) InviteUsrGp(user_id, group_id int64) (bool, error) {
	result, err := repo.db.Exec(`
		INSERT INTO members_status (group_id, user_id, member_status)
		SELECT $1, $2, 'invited'
		WHERE NOT EXISTS (SELECT 1 FROM groups WHERE id = $1 AND admin_id = $2)
		  AND NOT EXISTS (SELECT 1 FROM members_gp WHERE group_id = $1 AND user_id = $2)
		  AND NOT EXISTS (
			SELECT 1 FROM members_status
			WHERE group_id = $1 AND user_id = $2 AND member_status IN ('invited', 'pending')
		  )
	`, group_id, user_id)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}

func (repo *UserRepository) JoinUsrGroup(user_id, group_id int64) error {
	_, err := repo.db.Exec(`
		INSERT INTO members_status (group_id, user_id)
		VALUES ($1, $2)
	`, group_id, user_id)
	return err
}

func (repo *UserRepository) SearchGroup(name_sh string) ([]models.Group, error) {

	query := globalQuery + `
		WHERE title LIKE '%' || $1 || '%'
		GROUP BY g.id, g.title, g.description_gp, g.admin_id, u.username
	`
	rows, err := repo.db.Query(query, name_sh)
	if err != nil {
		fmt.Println("error database:", err)
		return nil, err
	}
	var groups []models.Group
	for rows.Next() {
		var group models.Group
		err := rows.Scan(&group.Id, &group.Title, &group.Description, &group.AdminId, &group.AdminName, &group.MemberCount)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (repo *UserRepository) GetGroup(group_id int64) (models.Group, error) {
	var group models.Group
	// query := `
	// 	SELECT g.id, g.title, g.description_gp, g.admin_id, u.username, COUNT(m.user_id) AS member_count
	// 	FROM groups	g
	// 	JOIN users u ON g.admin_id = u.id
	// 	LEFT JOIN members_gp m ON g.id = m.group_id
	// 	GROUP BY g.id, g.title, g.description_gp, g.admin_id, u.username
	// 	`
	// 	WHERE (g.admin_id = $1 OR g.id = (SELECT group_id FROM members_gp WHERE user_id = $1 ) OR )
	err := repo.db.QueryRow(globalQuery+"WHERE g.id = $1 GROUP BY g.id, g.title, g.description_gp, g.admin_id, u.username",
		group_id).Scan(&group.Id, &group.Title, &group.Description, &group.AdminId, &group.AdminName, &group.MemberCount)
	if err != nil {
		fmt.Println("Error retrieving groups from database:", err)
		return group, err
	}
	return group, nil
}

func (repo *UserRepository) CreatePostGp(group_id, User_id int64, post models.PostRequest) (int64, error) {
	transaction, err := repo.db.Begin()
	if err != nil {
		return 0, err
	}
	defer transaction.Rollback()
	query := `
		INSERT INTO posts (user_id,content,image_path,group_id)
		VALUES ($1,$2,$3,$4)
		`
	// 	WHERE (g.admin_id = $1 OR g.id = (SELECT group_id FROM members_gp WHERE user_id = $1 ) OR )
	result, err := transaction.Exec(query, User_id, post.Content, post.Image, group_id)
	if err != nil {
		fmt.Println("Error retrieving groups from database:", err)
		return 0, err
	}
	postId, err := result.LastInsertId()
	if err != nil {
		fmt.Println("Error retrieving groups from database:", err)
		return 0, err
	}
	if err := transaction.Commit(); err != nil {
		return 0, err
	}
	return postId, nil
}

func (repo *UserRepository) GetPostsGp(group_id int64) ([]models.PostData, error) {

	quey := `SELECT p.id, p.user_id, u.username, p.content, p.image_path, p.group_id, p.created_at
	FROM posts p JOIN users u ON u.id = p.user_id
	WHERE p.group_id=$1
	ORDER BY p.created_at DESC
	`
	rows, err := repo.db.Query(quey, group_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []models.PostData{}
	for rows.Next() {
		var post models.PostData
		if err := rows.Scan(&post.Id, &post.User_id, &post.Username, &post.Content, &post.Image, &post.Group_id, &post.Created_at); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}
