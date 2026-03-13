package test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaikai20040827/graduation/internal/model"
)

func TestFileServiceUploadAndDecrypt(t *testing.T) {
	db := newTestDB(t)
	fs := newTestFileService(t, db)

	plain := []byte("hello secure file box")
	file, err := fs.UploadFile(bytes.NewReader(plain), "hello.txt", 123, "desc", "text/plain")
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if file.ID == 0 {
		t.Fatalf("expected file ID")
	}

	got, err := fs.GetFileByID(file.ID)
	if err != nil {
		t.Fatalf("GetFileByID: %v", err)
	}
	if got.Filename != "hello.txt" || got.Description != "desc" || got.UploaderID != "123" {
		t.Fatalf("metadata mismatch: %+v", got)
	}

	out, err := fs.DecryptToBytesLimit(got.StoragePath, 0)
	if err != nil {
		t.Fatalf("DecryptToBytesLimit: %v", err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("decrypt mismatch")
	}
}

func TestFileServiceDecryptLimit(t *testing.T) {
	db := newTestDB(t)
	fs := newTestFileService(t, db)

	plain := bytes.Repeat([]byte("a"), 1024)
	file, err := fs.UploadFile(bytes.NewReader(plain), "a.txt", 1, "", "text/plain")
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	var buf bytes.Buffer
	n, err := fs.DecryptToWriterLimit(&buf, file.StoragePath, 100)
	if err != nil {
		t.Fatalf("DecryptToWriterLimit: %v", err)
	}
	if n != 100 {
		t.Fatalf("expected 100 bytes, got %d", n)
	}
	if buf.Len() != 100 {
		t.Fatalf("buffer length mismatch: %d", buf.Len())
	}
}

func TestFileServiceRemoveStoredFile(t *testing.T) {
	db := newTestDB(t)
	fs := newTestFileService(t, db)

	path := os.TempDir() + "/sfb_test_remove.bin"
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := fs.RemoveStoredFile(path); err != nil {
		t.Fatalf("RemoveStoredFile: %v", err)
	}
	if err := fs.RemoveStoredFile(path); err != nil {
		t.Fatalf("RemoveStoredFile idempotent: %v", err)
	}
}

func TestFileServiceDecryptInvalidMagic(t *testing.T) {
	db := newTestDB(t)
	fs := newTestFileService(t, db)

	tmp := filepath.Join(t.TempDir(), "bad.bin")
	if err := os.WriteFile(tmp, []byte("BADMAGIC"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := fs.DecryptToWriter(io.Discard, tmp); err == nil {
		t.Fatalf("expected invalid magic error")
	}
}

func TestFileServiceLegacyMetadataFallback(t *testing.T) {
	db := newTestDB(t)
	fs := newTestFileService(t, db)

	f := model.File{
		LegacyFilename: "legacy.txt",
		LegacyPath:     "/tmp/legacy.bin",
		LegacySize:     42,
		LegacyDesc:     "legacy",
		LegacyUploader: "7",
	}
	if err := db.Create(&f).Error; err != nil {
		t.Fatalf("create legacy: %v", err)
	}
	got, err := fs.GetFileByID(f.ID)
	if err != nil {
		t.Fatalf("GetFileByID: %v", err)
	}
	if got.Filename != "legacy.txt" || got.Size != 42 || got.UploaderID != "7" {
		t.Fatalf("legacy metadata not loaded: %+v", got)
	}
}
