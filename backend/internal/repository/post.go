package repository

import (
	"backend/internal/models"
	"fmt"
)

func (repo *UserRepository) CreatePost(userID int64, post models.PostRequest) (int64, error) {
	transaction, err := repo.db.Begin()
	if err != nil {
		return 0, err
	}
	defer transaction.Rollback()

	result, err := transaction.Exec(
		"INSERT INTO posts (user_id, content, privacy) VALUES (?, ?, ?)",
		userID, post.Content, post.Privacy,
	)
	if err != nil {
		return 0, err
	}
	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	for _, followerID := range post.FollowerIDs {
		_, err = transaction.Exec(
			`INSERT INTO post_allowed_users (post_id, user_id)
			 SELECT ?, ? WHERE EXISTS (
				SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ?
			 )`,
			postID, followerID, followerID, userID,
		)
		if err != nil {
			return 0, err
		}
	}

	if err := transaction.Commit(); err != nil {
		return 0, err
	}
	return postID, nil
}

func (repo *UserRepository) CanViewPost(postID, viewerID int64) (bool, error) {
	var allowed int
	err := repo.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM posts p
			WHERE p.id = ? AND (
				p.privacy = 'public' OR p.user_id = ? OR
				(p.privacy = 'almost_private' AND EXISTS (
					SELECT 1 FROM followers f WHERE f.follower_id = ? AND f.following_id = p.user_id
				)) OR
				(p.privacy = 'private' AND EXISTS (
					SELECT 1 FROM post_allowed_users a WHERE a.post_id = p.id AND a.user_id = ?
				))
			)
		)`, postID, viewerID, viewerID, viewerID).Scan(&allowed)
	return allowed == 1, err
}

func (repo *UserRepository) GetPosts(viewerID int64) ([]models.PostData, error) {
	rows, err := repo.db.Query(postQuery+` WHERE p.privacy = 'public' OR p.user_id = ? OR
		(p.privacy = 'almost_private' AND EXISTS (SELECT 1 FROM followers f WHERE f.follower_id = ? AND f.following_id = p.user_id)) OR
		(p.privacy = 'private' AND EXISTS (SELECT 1 FROM post_allowed_users a WHERE a.post_id = p.id AND a.user_id = ?))
		ORDER BY p.created_at DESC`, viewerID, viewerID, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []models.PostData{}
	for rows.Next() {
		var post models.PostData
		if err := rows.Scan(&post.Id, &post.User_id, &post.Username, &post.Content, &post.Privacy, &post.Created_at); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

func (repo *UserRepository) GetPostByUsername(username string, viewerID int64) ([]models.PostData, error) {
	var post models.PostData
	rows, err := repo.db.Query(postQuery+` WHERE u.username = $1 
	AND (p.privacy = 'public' 
		OR p.user_id = $2 
		OR
		(p.privacy = 'almost_private' AND EXISTS (
			SELECT 1 FROM followers f WHERE f.follower_id = $2 
			AND f.following_id = p.user_id)) 
		OR
		(p.privacy = 'private' AND EXISTS (
			SELECT 1 FROM post_allowed_users a WHERE a.post_id = p.id 
			AND a.user_id = $2))
		)`,
	username, viewerID)

	if err != nil {
		fmt.Println("error getpostbyusername database", err)
		return nil, err
	}
	defer rows.Close()
	var postList []models.PostData
	for rows.Next() {
		if err := rows.Scan(
			&post.Id, 
			&post.User_id, 
			&post.Username, 
			&post.Content, 
			&post.Privacy, 
			&post.Created_at);
			 err != nil {
			return nil, err
		}
		postList = append(postList, post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(postList) == 0 {
		return nil, fmt.Errorf("no posts found for username: %s", username)
	}
	return postList, nil
}

const postQuery = `SELECT p.id, p.user_id, u.username, p.content, p.privacy, p.created_at
	FROM posts p JOIN users u ON u.id = p.user_id`

func (repo *UserRepository) ValidatePrivateFollowers(userID int64, followerIDs []int64) error {
	for _, followerID := range followerIDs {
		var exists int
		err := repo.db.QueryRow("SELECT EXISTS (SELECT 1 FROM followers WHERE follower_id = $1 AND following_id = $2)", followerID, userID).Scan(&exists)
		if err != nil {
			return err
		}
		if exists != 1 {
			return fmt.Errorf("user %d is not following you", followerID)
		}
	}
	return nil
}


