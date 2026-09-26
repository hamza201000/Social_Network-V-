package repository

import (
	"backend/internal/models"
)

func (repo *UserRepository) CreateComment(userID, postID int64, comment models.CommentRequest) (int64, error) {
	result, err := repo.db.Exec("INSERT INTO comments (user_id, post_id, content, image_path) VALUES (?, ?, ?, ?)", userID, postID, comment.Content, comment.Image)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (repo *UserRepository) GetComments(postID int64) ([]models.CommentData, error) {
	rows, err := repo.db.Query(`SELECT c.id, c.user_id, u.username, c.post_id, c.content, c.image_path, c.created_at
		FROM comments c JOIN users u ON u.id = c.user_id WHERE c.post_id = ? ORDER BY c.created_at ASC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []models.CommentData{}
	for rows.Next() {
		var comment models.CommentData
		if err := rows.Scan(&comment.Id, &comment.User_id, &comment.Username, &comment.Post_id, &comment.Content, &comment.Image, &comment.Created_at); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}
