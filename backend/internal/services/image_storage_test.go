package services

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImageStorageSavesUniqueFilesAndRejectsUnsafeNames(t *testing.T) {
	storage := NewImageStorage(t.TempDir())

	first, err := storage.Save(strings.NewReader("image bytes"), "image/webp")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	second, err := storage.Save(strings.NewReader("image bytes"), "image/webp")
	if err != nil {
		t.Fatalf("Save() second error = %v", err)
	}
	if first == second {
		t.Fatalf("duplicate uploads returned the same filename %q", first)
	}
	if !ValidImageFilename(first) || filepath.Ext(first) != ".webp" {
		t.Fatalf("unexpected stored filename %q", first)
	}

	file, _, err := storage.Open(first)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	content, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		t.Fatalf("read stored file: %v", err)
	}
	if string(content) != "image bytes" {
		t.Fatalf("stored content = %q", content)
	}

	if _, _, err := storage.Open("../outside"); !errors.Is(err, os.ErrInvalid) {
		t.Fatalf("Open() unsafe name error = %v, want os.ErrInvalid", err)
	}
	if err := storage.Delete(first); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(storage.uploadDir, first)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("deleted file stat error = %v, want os.ErrNotExist", err)
	}
}
