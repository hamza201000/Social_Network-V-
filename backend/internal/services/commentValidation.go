package services

import (
	"backend/internal/models"
	"errors"
	"strings"
)

func ValidateCommentRequest(comment models.CommentRequest) error {
	if strings.TrimSpace(comment.Content) == "" && comment.Image == "" {
		return errors.New("write something or add an image")
	}
	if len(comment.Content) > 2000 {
		return errors.New("comment must not exceed 2000 characters")
	}
	return nil
}
