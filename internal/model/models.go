package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	Email           string         `gorm:"uniqueIndex;size:255" json:"email"`
	Username        string         `gorm:"size:100" json:"username"`
	Password        string         `gorm:"size:255" json:"-"`
	AvatarPath      string         `gorm:"size:1024" json:"-"`
	AvatarMime      string         `gorm:"size:128" json:"-"`
	AvatarUpdatedAt *time.Time     `json:"avatar_updated_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type File struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	EncFilename    string         `gorm:"column:enc_filename;type:text" json:"-"`
	EncStoragePath string         `gorm:"column:enc_storage_path;type:text" json:"-"`
	EncSize        string         `gorm:"column:enc_size;type:text" json:"-"`
	EncDescription string         `gorm:"column:enc_description;type:text" json:"-"`
	EncUploaderID  string         `gorm:"column:enc_uploader_id;type:text" json:"-"`
	LegacyFilename string         `gorm:"column:filename" json:"-"`
	LegacyPath     string         `gorm:"column:storage_path" json:"-"`
	LegacySize     int64          `gorm:"column:size" json:"-"`
	LegacyDesc     string         `gorm:"column:description" json:"-"`
	LegacyUploader string         `gorm:"column:uploader_id" json:"-"`
	Filename       string         `gorm:"-" json:"filename"`
	StoragePath    string         `gorm:"-" json:"-"`
	Size           int64          `gorm:"-" json:"size"`
	Description    string         `gorm:"-" json:"description"`
	UploaderID     string         `gorm:"-" json:"uploader_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
