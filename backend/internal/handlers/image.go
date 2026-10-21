package handler

import (
	"backend/internal/repository"
	"backend/internal/services"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type ImageHandler struct {
	storage *services.ImageStorage
}

func NewImageHandler(storage *services.ImageStorage) *ImageHandler {
	return &ImageHandler{storage: storage}
}

func (imageHandler *ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	file, info, err := imageHandler.storage.Open(filename)
	if errors.Is(err, os.ErrNotExist) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("open image: %v", err)
		SendError(w, http.StatusInternalServerError, "could not retrieve image")
		return
	}
	defer file.Close()

	w.Header().Set("Cache-Control", "private, no-cache")
	w.Header().Set("Content-Type", imageContentType(filepath.Ext(filename)))
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, filename, info.ModTime(), file)
}

func imageContentType(extension string) string {
	switch extension {
	case ".jpg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

func (imageHandler *ImageHandler) GetAuthorizedImage(repo *repository.UserRepository, w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	filename := r.PathValue("filename")
	if (filename=="icon_user.png"){
		imageHandler.GetImage(w, r);
		return 
	}
	if !services.ValidImageFilename(filename) {
		fmt.Println("error double error :::",filename)
		SendError(w, http.StatusBadRequest, "invalid image filename")
		return
	}
	allowed, err := repo.CanViewImage(filename, userID)
	if err != nil {
		log.Printf("authorize image: %v", err)
		SendError(w, http.StatusInternalServerError, "could not retrieve image")
		return
	}
	if !allowed {
		http.NotFound(w, r)
		return
	}
	imageHandler.GetImage(w, r)
}
