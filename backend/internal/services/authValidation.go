package services

import (
	"backend/internal/models"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

func RegisterRequest(userData models.RegisterRequest) error {
	if strings.TrimSpace(userData.FirstName) == "" {
		return errors.New("first name is required")
	}

	if strings.TrimSpace(userData.LastName) == "" {
		return errors.New("last name is required")
	}

	username := strings.TrimSpace(userData.Username)

	if username == "" {
		return errors.New("username is required")
	}

	if len(username) < 3 {
		return errors.New("username must be at least 3 characters")
	}

	if len(username) > 30 {
		return errors.New("username must not exceed 30 characters")
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(username) {
		return errors.New("username can only contain letters, numbers and underscores")
	}

	email := strings.TrimSpace(userData.Email)

	if email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(
		`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
	)

	if !emailRegex.MatchString(email) {
		return errors.New("invalid email address")
	}

	if userData.Password == "" {
		return errors.New("password is required")
	}

	if len(userData.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if len(userData.Password) > 72 {
		return errors.New("password must not exceed 72 characters")
	}

	if userData.DateOfBirth != "" {
		dateOfBirth, err := time.Parse("2006-01-02", userData.DateOfBirth)
		if err != nil {
			return errors.New("date of birth must be a valid date")
		}
		if dateOfBirth.After(time.Now()) {
			return errors.New("date of birth cannot be in the future")
		}
	}

	if utf8.RuneCountInString(userData.AboutMe) > 500 {
		return errors.New("about me must not exceed 500 characters")
	}
	return nil
}
