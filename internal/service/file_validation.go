package service

import (
	"bytes"
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var (
	allowedExtensions = map[string]struct{}{
		"jpg": {}, "jpeg": {}, "png": {}, "gif": {}, "webp": {},
		"txt": {}, "md": {}, "json": {}, "log": {}, "csv": {},
		"pdf": {},
		"doc": {}, "docx": {},
		"xls": {}, "xlsx": {},
		"ppt": {}, "pptx": {},
	}
	blockedExtensions = map[string]struct{}{
		"exe": {}, "dll": {}, "so": {}, "bin": {}, "sh": {}, "bat": {},
		"apk": {}, "dmg": {}, "iso": {}, "msi": {}, "com": {}, "scr": {},
	}
	extToMime = map[string][]string{
		"jpg":  {"image/jpeg"},
		"jpeg": {"image/jpeg"},
		"png":  {"image/png"},
		"gif":  {"image/gif"},
		"webp": {"image/webp"},
		"pdf":  {"application/pdf"},
		"doc":  {"application/msword", "application/vnd.ms-office", "application/octet-stream"},
		"docx": {"application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/zip"},
		"xls":  {"application/vnd.ms-excel", "application/vnd.ms-office", "application/octet-stream"},
		"xlsx": {"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "application/zip"},
		"ppt":  {"application/vnd.ms-powerpoint", "application/vnd.ms-office", "application/octet-stream"},
		"pptx": {"application/vnd.openxmlformats-officedocument.presentationml.presentation", "application/zip"},
		"txt":  {"text/plain"},
		"md":   {"text/plain", "text/markdown"},
		"log":  {"text/plain"},
		"csv":  {"text/plain", "text/csv"},
		"json": {"application/json", "text/plain"},
	}
)

func NormalizeExt(filename string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	return ext
}

func ValidateExtension(filename string) (string, error) {
	ext := NormalizeExt(filename)
	if ext == "" {
		return "", errors.New("file extension required")
	}
	if _, blocked := blockedExtensions[ext]; blocked {
		return "", errors.New("file type blocked")
	}
	if _, ok := allowedExtensions[ext]; !ok {
		return "", errors.New("file type not allowed")
	}
	return ext, nil
}

func ValidateContent(filename string, sample []byte) (string, error) {
	ext, err := ValidateExtension(filename)
	if err != nil {
		return "", err
	}
	if len(sample) == 0 {
		return "", errors.New("empty file")
	}

	mimeType := http.DetectContentType(sample)
	mimeType = normalizeMime(mimeType)

	if !matchesAllowedMime(ext, mimeType) {
		return "", errors.New("content type mismatch")
	}

	if isBinaryType(ext) {
		if !matchesMagic(ext, sample) {
			return "", errors.New("file header mismatch")
		}
		return mimeType, nil
	}

	if !looksLikeText(sample) {
		return "", errors.New("file content is not text")
	}
	return mimeType, nil
}

func normalizeMime(m string) string {
	if m == "" {
		return ""
	}
	parts := strings.Split(m, ";")
	return strings.TrimSpace(parts[0])
}

func matchesAllowedMime(ext string, mimeType string) bool {
	allowed, ok := extToMime[ext]
	if !ok || mimeType == "" {
		return false
	}
	for _, m := range allowed {
		if strings.EqualFold(m, mimeType) {
			return true
		}
	}
	return false
}

func matchesMagic(ext string, sample []byte) bool {
	switch ext {
	case "jpg", "jpeg":
		return len(sample) > 3 && sample[0] == 0xFF && sample[1] == 0xD8 && sample[2] == 0xFF
	case "png":
		return bytes.HasPrefix(sample, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	case "gif":
		return bytes.HasPrefix(sample, []byte("GIF87a")) || bytes.HasPrefix(sample, []byte("GIF89a"))
	case "webp":
		return len(sample) >= 12 && bytes.HasPrefix(sample, []byte("RIFF")) && bytes.HasPrefix(sample[8:], []byte("WEBP"))
	case "pdf":
		return bytes.HasPrefix(sample, []byte("%PDF-"))
	case "docx":
		// DOCX is a ZIP container
		return len(sample) >= 4 && bytes.HasPrefix(sample, []byte("PK\x03\x04"))
	case "xlsx":
		// XLSX is a ZIP container
		return len(sample) >= 4 && bytes.HasPrefix(sample, []byte("PK\x03\x04"))
	case "pptx":
		// PPTX is a ZIP container
		return len(sample) >= 4 && bytes.HasPrefix(sample, []byte("PK\x03\x04"))
	case "doc", "xls", "ppt":
		// Legacy Office formats use OLE Compound File Binary (CFB)
		return len(sample) >= 8 && bytes.HasPrefix(sample, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})
	default:
		return true
	}
}

func looksLikeText(sample []byte) bool {
	if len(sample) == 0 {
		return false
	}
	if bytes.Contains(sample, []byte{0x00}) {
		return false
	}
	if !utf8.Valid(sample) {
		return false
	}
	return true
}

func isBinaryType(ext string) bool {
	switch ext {
	case "jpg", "jpeg", "png", "gif", "webp", "pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx":
		return true
	default:
		return false
	}
}

func GuessMimeByExtension(filename string) string {
	m := mime.TypeByExtension(filepath.Ext(filename))
	return normalizeMime(m)
}
