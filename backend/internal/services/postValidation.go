package services

import (
	"backend/internal/models"
	"errors"
	"strings"
)

func ValidatePostRequest(postData models.PostRequest) error {
	if strings.TrimSpace(postData.Content) == ""  {
		return errors.New("post content is required")
	}
	if postData.Privacy != "public" && postData.Privacy != "almost_private" && postData.Privacy != "private" {
		return errors.New("privacy must be public, almost_private or private")
	}
	return nil
}
