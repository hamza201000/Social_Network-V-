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
func (repo *UserRepository) InviteUsrGp(user_id, group_id int64) (int64, error) {

	return 0, nil
}

func (repo *UserRepository) JoinUsrGroup(user_id, group_id int64) error {

	query := `
		INSERT INTO members_status (group_id,user_id)
		VALUES(1$,2$)
	`
	_, err := repo.db.Exec(query, user_id, group_id)
	if err != nil {
		return err
	}
	return nil
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
