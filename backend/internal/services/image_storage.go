package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var imageFilenamePattern = regexp.MustCompile(`^[a-f0-9]{32}\.(jpg|png|webp|gif)$`)

type ImageStorage struct {
	uploadDir string
}

func NewImageStorage(backendRoot string) *ImageStorage {
	return &ImageStorage{uploadDir: filepath.Join(backendRoot, "database", "upload")}
}

func (storage *ImageStorage) Save(source io.Reader, contentType string) (string, error) {
	extensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
	}
	extension, ok := extensions[contentType]
	if !ok {
		return "", errors.New("unsupported image type")
	}
	if err := os.MkdirAll(storage.uploadDir, 0755); err != nil {
		return "", fmt.Errorf("create image directory: %w", err)
	}
	for attempt := 0; attempt < 5; attempt++ {
		randomID := make([]byte, 16)
		if _, err := rand.Read(randomID); err != nil {
			return "", fmt.Errorf("generate image identifier: %w", err)
		}
		filename := hex.EncodeToString(randomID) + extension
		path := filepath.Join(storage.uploadDir, filename)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("create image file: %w", err)
		}

		_, copyErr := io.Copy(file, source)
		closeErr := file.Close()
		if copyErr != nil {
			_ = os.Remove(path)
			return "", fmt.Errorf("write image file: %w", copyErr)
		}
		if closeErr != nil {
			_ = os.Remove(path)
			return "", fmt.Errorf("close image file: %w", closeErr)
		}
		return filename, nil
	}
	return "", errors.New("could not generate a unique image filename")
}

func (storage *ImageStorage) Open(filename string) (*os.File, os.FileInfo, error) {
	
	path := filepath.Join(storage.uploadDir, filename)
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		file.Close()
		return nil, nil, os.ErrNotExist
	}
	return file, info, nil
}

func (storage *ImageStorage) Delete(filename string) error {
	if !ValidImageFilename(filename) {
		return os.ErrInvalid
	}
	err := os.Remove(filepath.Join(storage.uploadDir, filename))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func ValidImageFilename(filename string) bool {
	
	return imageFilenamePattern.MatchString(filename)
}
