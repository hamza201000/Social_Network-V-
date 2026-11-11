package repository

import (
	"backend/internal/models"
	"fmt"
)

// globalQuery selects groups with their member count (admin included) and the
// viewer's status in each group. $1 must be the viewer's user id.
const globalQuery = `
		SELECT g.id, g.title, g.description_gp, g.admin_id, u.username,
			COUNT(DISTINCT m.user_id) + 1 AS member_count,
			CASE
				WHEN g.admin_id = $1 THEN 'admin'
				WHEN EXISTS (SELECT 1 FROM members_gp mm WHERE mm.group_id = g.id AND mm.user_id = $1) THEN 'member'
				ELSE COALESCE((
					SELECT s.member_status FROM members_status s
					WHERE s.group_id = g.id AND s.user_id = $1 AND s.member_status IN ('invited', 'pending')
					ORDER BY s.id DESC LIMIT 1
				), 'none')
			END AS my_status
		FROM groups g
		JOIN users u ON g.admin_id = u.id
		LEFT JOIN members_gp m ON g.id = m.group_id AND m.user_id != g.admin_id
		`

const groupBy = ` GROUP BY g.id, g.title, g.description_gp, g.admin_id, u.username`

func scanGroup(row rowScanner, viewerID int64) (models.Group, error) {
	var group models.Group
	err := row.Scan(&group.Id, &group.Title, &group.Description, &group.AdminId,
		&group.AdminName, &group.MemberCount, &group.MyStatus)
	if err == nil && group.AdminId == viewerID {
		group.AdminName = "You"
	}
	return group, err
}

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
	rows, err := repo.db.Query(globalQuery+groupBy+` ORDER BY g.created_at DESC`, user_id)
	if err != nil {
		fmt.Println("Error retrieving groups from database:", err)
		return nil, err
	}
	defer rows.Close()
	groups := []models.Group{}
	for rows.Next() {
		group, err := scanGroup(rows, user_id)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (repo *UserRepository) SearchGroup(user_id int64, name_sh string) ([]models.Group, error) {
	rows, err := repo.db.Query(globalQuery+`WHERE g.title LIKE '%' || $2 || '%'`+groupBy, user_id, name_sh)
	if err != nil {
		fmt.Println("error database:", err)
		return nil, err
	}
	defer rows.Close()
	groups := []models.Group{}
	for rows.Next() {
		group, err := scanGroup(rows, user_id)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (repo *UserRepository) GetGroup(user_id, group_id int64) (models.Group, error) {
	return scanGroup(repo.db.QueryRow(globalQuery+`WHERE g.id = $2`+groupBy, user_id, group_id), user_id)
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
