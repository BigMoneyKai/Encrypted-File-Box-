package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Kaikai20040827/graduation/internal/model"
	"gorm.io/gorm"
)

const (
	maxChunkSize      = 20 * 1024 * 1024
	minChunkSize      = 256 * 1024
	sessionStatusNew  = "pending"
	sessionStatusDone = "completed"
)

func (f *FileService) InitResumableUpload(filename string, totalSize int64, chunkSize int64, uploaderID uint, description string) (*model.UploadSession, error) {
	if _, err := ValidateExtension(filename); err != nil {
		return nil, err
	}
	if totalSize <= 0 {
		return nil, errors.New("invalid file size")
	}
	if chunkSize < minChunkSize {
		chunkSize = minChunkSize
	}
	if chunkSize > maxChunkSize {
		chunkSize = maxChunkSize
	}

	totalChunks := int(math.Ceil(float64(totalSize) / float64(chunkSize)))
	uploadID, err := randomStorageName()
	if err != nil {
		return nil, err
	}
	uploadID = "u_" + strings.TrimSuffix(uploadID, ".bin")

	tempDir := filepath.Join(f.dirpath, tmpUploadDir, uploadID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, err
	}

	session := &model.UploadSession{
		UploadID:    uploadID,
		UploaderID:  uploaderID,
		Filename:    filename,
		Description: description,
		TotalSize:   totalSize,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		Status:      sessionStatusNew,
		TempDir:     tempDir,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := f.db.Create(session).Error; err != nil {
		return nil, err
	}
	return session, nil
}

func (f *FileService) SaveUploadChunk(ctx context.Context, uploadID string, index int, reader io.Reader) error {
	session, err := f.GetUploadSession(uploadID)
	if err != nil {
		return err
	}
	if session.Status != sessionStatusNew {
		return errors.New("upload session not active")
	}
	if index < 0 || index >= session.TotalChunks {
		return errors.New("invalid chunk index")
	}

	chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk_%06d.part", index))
	out, err := os.Create(chunkPath)
	if err != nil {
		return err
	}
	defer out.Close()

	n, err := io.Copy(out, reader)
	if err != nil {
		return err
	}

	chunk := &model.UploadChunk{
		UploadID:  uploadID,
		Index:     index,
		Size:      n,
		CreatedAt: time.Now(),
	}
	if err := f.db.Where("upload_id = ? AND `index` = ?", uploadID, index).Delete(&model.UploadChunk{}).Error; err != nil {
		return err
	}
	return f.db.Create(chunk).Error
}

func (f *FileService) ListUploadedChunks(uploadID string) ([]int, error) {
	var chunks []model.UploadChunk
	if err := f.db.Where("upload_id = ?", uploadID).Find(&chunks).Error; err != nil {
		return nil, err
	}
	indices := make([]int, 0, len(chunks))
	for _, c := range chunks {
		indices = append(indices, c.Index)
	}
	sort.Ints(indices)
	return indices, nil
}

func (f *FileService) CompleteUpload(ctx context.Context, uploadID string) (*model.File, error) {
	session, err := f.GetUploadSession(uploadID)
	if err != nil {
		return nil, err
	}
	if session.Status != sessionStatusNew {
		return nil, errors.New("upload session not active")
	}

	uploaded, err := f.ListUploadedChunks(uploadID)
	if err != nil {
		return nil, err
	}
	if len(uploaded) != session.TotalChunks {
		return nil, errors.New("missing chunks")
	}

	mergedPath := filepath.Join(session.TempDir, "merged.raw")
	out, err := os.Create(mergedPath)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk_%06d.part", i))
		in, err := os.Open(chunkPath)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(out, in); err != nil {
			in.Close()
			return nil, err
		}
		in.Close()
	}

	sample, err := readSample(mergedPath, 512)
	if err != nil {
		return nil, err
	}
	mimeType, err := ValidateContent(session.Filename, sample)
	if err != nil {
		return nil, err
	}
	if err := f.scanFile(ctx, mergedPath); err != nil {
		return nil, err
	}

	file, err := f.UploadFileFromPath(mergedPath, session.Filename, session.UploaderID, session.Description, mimeType)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session.Status = sessionStatusDone
	session.FileID = &file.ID
	session.CompletedAt = &now
	session.UpdatedAt = now
	if err := f.db.Save(session).Error; err != nil {
		return nil, err
	}

	_ = os.RemoveAll(session.TempDir)
	_ = f.db.Where("upload_id = ?", uploadID).Delete(&model.UploadChunk{}).Error
	return file, nil
}

func (f *FileService) AbortUpload(uploadID string) error {
	session, err := f.GetUploadSession(uploadID)
	if err != nil {
		return err
	}
	_ = os.RemoveAll(session.TempDir)
	_ = f.db.Where("upload_id = ?", uploadID).Delete(&model.UploadChunk{}).Error
	return f.db.Delete(session).Error
}

func (f *FileService) GetUploadSession(uploadID string) (*model.UploadSession, error) {
	var session model.UploadSession
	if err := f.db.Where("upload_id = ?", uploadID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("upload session not found")
		}
		return nil, err
	}
	return &session, nil
}

func readSample(path string, n int) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, n)
	read, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return buf[:read], nil
}
