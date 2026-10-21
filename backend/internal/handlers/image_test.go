package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	handler "backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/services"
)

func TestGetImageServesContentAndRejectsUnsafeNames(t *testing.T) {
	storage := services.NewImageStorage(t.TempDir())
	filename, err := storage.Save(strings.NewReader("image bytes"), "image/png")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	imageHandler := handler.NewImageHandler(storage)

	request := httptest.NewRequest(http.MethodGet, "/api/images/"+filename, nil)
	request.SetPathValue("filename", filename)
	response := httptest.NewRecorder()
	imageHandler.GetImage(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body)
	}
	if got := response.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("Content-Type = %q, want image/png", got)
	}
	if got := response.Body.String(); got != "image bytes" {
		t.Fatalf("body = %q, want image bytes", got)
	}
	if got := response.Header().Get("Cache-Control"); got != "private, no-cache" {
		t.Fatalf("Cache-Control = %q, want private, no-cache", got)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/images/"+filename, nil)
	request.SetPathValue("filename", filename)
	request.Header.Set("If-Modified-Since", response.Header().Get("Last-Modified"))
	response = httptest.NewRecorder()
	imageHandler.GetImage(response, request)
	if response.Code != http.StatusNotModified {
		t.Fatalf("conditional request status = %d, want %d", response.Code, http.StatusNotModified)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/images/../secret", nil)
	request.SetPathValue("filename", "../secret")
	response = httptest.NewRecorder()
	imageHandler.GetImage(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("unsafe filename status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestGetImageReturnsNotFound(t *testing.T) {
	imageHandler := handler.NewImageHandler(services.NewImageStorage(t.TempDir()))
	filename := "00000000000000000000000000000000.png"
	request := httptest.NewRequest(http.MethodGet, "/api/images/"+filename, nil)
	request.SetPathValue("filename", filename)
	response := httptest.NewRecorder()

	imageHandler.GetImage(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestImageRouteRequiresAuthentication(t *testing.T) {
	imageHandler := handler.NewImageHandler(services.NewImageStorage(t.TempDir()))
	protected := middleware.Middleware(repository.NewUserRepository(nil), imageHandler.GetImage)
	request := httptest.NewRequest(http.MethodGet, "/api/images/image.png", nil)
	response := httptest.NewRecorder()

	protected.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
