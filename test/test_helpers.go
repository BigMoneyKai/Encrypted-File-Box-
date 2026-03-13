package test

import (
	"bytes"
	"encoding/base64"
	"path/filepath"
	"testing"

	"github.com/Kaikai20040827/graduation/internal/model"
	"github.com/Kaikai20040827/graduation/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.File{},
		&model.UploadSession{},
		&model.UploadChunk{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func testBase64Key() string {
	raw := bytes.Repeat([]byte{0x11}, 32)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func newTestFileService(t *testing.T, db *gorm.DB) *service.FileService {
	t.Helper()
	return service.NewFileService(
		db,
		t.TempDir(),
		testBase64Key(),
		service.ScanOptions{Enabled: false},
	)
}
