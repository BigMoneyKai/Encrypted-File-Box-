package test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestResumableUploadFlow(t *testing.T) {
	db := newTestDB(t)
	fs := newTestFileService(t, db)

	content := bytes.Repeat([]byte("hello-"), 1000) // 6000 bytes
	session, err := fs.InitResumableUpload("chunk.txt", int64(len(content)), 1024, 9, "desc")
	if err != nil {
		t.Fatalf("InitResumableUpload: %v", err)
	}

	chunkSize := int(session.ChunkSize)
	for i := 0; i < session.TotalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(content) {
			end = len(content)
		}
		if err := fs.SaveUploadChunk(context.Background(), session.UploadID, i, bytes.NewReader(content[start:end])); err != nil {
			t.Fatalf("SaveUploadChunk %d: %v", i, err)
		}
	}

	list, err := fs.ListUploadedChunks(session.UploadID)
	if err != nil {
		t.Fatalf("ListUploadedChunks: %v", err)
	}
	if len(list) != session.TotalChunks {
		t.Fatalf("uploaded chunks mismatch: %d", len(list))
	}

	file, err := fs.CompleteUpload(context.Background(), session.UploadID)
	if err != nil {
		t.Fatalf("CompleteUpload: %v", err)
	}

	plain, err := fs.DecryptToBytesLimit(file.StoragePath, 0)
	if err != nil {
		t.Fatalf("DecryptToBytesLimit: %v", err)
	}
	if !bytes.Equal(plain, content) {
		t.Fatalf("merged content mismatch")
	}

	if _, err := os.Stat(filepath.Dir(file.StoragePath)); err != nil && os.IsNotExist(err) {
		t.Fatalf("storage dir missing")
	}
}

func TestResumableUploadMissingChunk(t *testing.T) {
	db := newTestDB(t)
	fs := newTestFileService(t, db)

	totalSize := int64(600000)
	session, err := fs.InitResumableUpload("chunk.txt", totalSize, 256, 1, "")
	if err != nil {
		t.Fatalf("InitResumableUpload: %v", err)
	}

	if err := fs.SaveUploadChunk(context.Background(), session.UploadID, 0, bytes.NewReader([]byte("partial"))); err != nil {
		t.Fatalf("SaveUploadChunk: %v", err)
	}

	if _, err := fs.CompleteUpload(context.Background(), session.UploadID); err == nil {
		t.Fatalf("expected missing chunks error")
	}
}

func TestResumableUploadInvalidChunkIndex(t *testing.T) {
	db := newTestDB(t)
	fs := newTestFileService(t, db)

	session, err := fs.InitResumableUpload("chunk.txt", 1024, 512, 1, "")
	if err != nil {
		t.Fatalf("InitResumableUpload: %v", err)
	}
	if err := fs.SaveUploadChunk(context.Background(), session.UploadID, session.TotalChunks+1, bytes.NewReader(nil)); err == nil {
		t.Fatalf("expected invalid chunk index")
	}
}
