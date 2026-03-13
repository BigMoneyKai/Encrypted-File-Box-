package test

import (
	"testing"

	"github.com/Kaikai20040827/graduation/internal/service"
)

func TestValidateExtension(t *testing.T) {
	if _, err := service.ValidateExtension("report.pdf"); err != nil {
		t.Fatalf("expected pdf allowed: %v", err)
	}
	if _, err := service.ValidateExtension("script.sh"); err == nil {
		t.Fatalf("expected blocked extension error")
	}
	if _, err := service.ValidateExtension("file.unknown"); err == nil {
		t.Fatalf("expected not allowed error")
	}
}

func TestValidateContentText(t *testing.T) {
	sample := []byte("hello world\n")
	mimeType, err := service.ValidateContent("note.txt", sample)
	if err != nil {
		t.Fatalf("ValidateContent: %v", err)
	}
	if mimeType == "" {
		t.Fatalf("mimeType empty")
	}
}

func TestValidateContentBinaryMismatch(t *testing.T) {
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	if _, err := service.ValidateContent("note.txt", jpeg); err == nil {
		t.Fatalf("expected content type mismatch")
	}
}

func TestValidateContentEmpty(t *testing.T) {
	if _, err := service.ValidateContent("note.txt", nil); err == nil {
		t.Fatalf("expected empty file error")
	}
}
