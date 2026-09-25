package services

import (
	"backend/internal/models"
	"errors"
	"strings"
)

func ValidateCommentRequest(comment models.CommentRequest) error {
	if strings.TrimSpace(comment.Content) == ""  {
		return errors.New("comment content is required")
	}
	return nil
}
