package service

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Kaikai20040827/graduation/internal/model"
	"gorm.io/gorm"
)

type FileService struct {
	db      *gorm.DB
	dirpath string
	fileKey []byte
	metaKey []byte
	fileGCM cipher.AEAD
	metaGCM cipher.AEAD
}

const (
	fileMagicV2     = "SFB2"
	fileNoncePrefix = 8
	fileNonceSize   = 12
	metaNonceSize   = 12
	chunkSize       = 32 * 1024
)

func NewFileService(db *gorm.DB, storagePath string, base64Key string) *FileService {
	_ = os.MkdirAll(storagePath, 0755)
	fmt.Println("✓ Creating a new file service done")

	fileKey, metaKey := deriveKeys(base64Key)
	var fileGCM cipher.AEAD
	var metaGCM cipher.AEAD

	if len(fileKey) == 32 {
		if block, err := aes.NewCipher(fileKey); err == nil {
			fileGCM, _ = cipher.NewGCM(block)
		}
	}
	if len(metaKey) == 32 {
		if block, err := aes.NewCipher(metaKey); err == nil {
			metaGCM, _ = cipher.NewGCM(block)
		}
	}

	return &FileService{db: db, dirpath: storagePath, fileKey: fileKey, metaKey: metaKey, fileGCM: fileGCM, metaGCM: metaGCM}
}

func (f *FileService) UploadFile(fileReader io.Reader, filename string, uploaderID uint, description string) (*model.File, error) {
	storedName, err := randomStorageName()
	if err != nil {
		return nil, err
	}
	dst := filepath.Join(f.dirpath, storedName)
	size, err := f.encryptToFile(fileReader, dst)
	if err != nil {
		return nil, err
	}

	file := &model.File{
		Filename:    filename,
		StoragePath: dst,
		Size:        size,
		Description: description,
		UploaderID:  fmt.Sprintf("%d", uploaderID),
		CreatedAt:   time.Now(),
	}
	if err := f.encryptFileMetadata(file); err != nil {
		_ = os.Remove(dst)
		return nil, err
	}

	if err := f.db.Create(file).Error; err != nil {
		_ = os.Remove(dst)
		return nil, err
	}
	return file, nil
}

func (f *FileService) SaveUserAvatar(fileReader io.Reader, filename string, userID uint) (string, int64, error) {
	storedName, err := randomStorageName()
	if err != nil {
		return "", 0, err
	}
	dst := filepath.Join(f.dirpath, "avatar_"+storedName)
	size, err := f.encryptToFile(fileReader, dst)
	if err != nil {
		return "", 0, err
	}
	_ = filename
	_ = userID
	return dst, size, nil
}

func (f *FileService) RemoveStoredFile(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (f *FileService) UpdateFile(id uint, fileReader io.Reader, filename *string, description *string) (*model.File, error) {
	file, err := f.GetFileByID(id)
	if err != nil {
		return nil, err
	}

	if fileReader != nil {
		tmpPath := fmt.Sprintf("%s.tmp.%d", file.StoragePath, time.Now().UnixNano())
		size, err := f.encryptToFile(fileReader, tmpPath)
		if err != nil {
			_ = os.Remove(tmpPath)
			return nil, err
		}
		if err := os.Rename(tmpPath, file.StoragePath); err != nil {
			_ = os.Remove(tmpPath)
			return nil, err
		}
		file.Size = size
		if filename != nil && *filename != "" {
			file.Filename = *filename
		}
	}

	if description != nil {
		file.Description = *description
	}

	if err := f.encryptFileMetadata(file); err != nil {
		return nil, err
	}
	if err := f.db.Model(&model.File{}).Where("id = ?", id).Updates(map[string]interface{}{
		"enc_filename":     file.EncFilename,
		"enc_storage_path": file.EncStoragePath,
		"enc_size":         file.EncSize,
		"enc_description":  file.EncDescription,
		"enc_uploader_id":  file.EncUploaderID,
		"filename":         "",
		"storage_path":     "",
		"size":             0,
		"description":      "",
		"uploader_id":      "",
	}).Error; err != nil {
		return nil, err
	}

	return file, nil
}

func (f *FileService) DeleteFile(id uint) error {
	file, err := f.GetFileByID(id)
	if err != nil {
		return err
	}

	if err := os.Remove(file.StoragePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return f.db.Delete(&model.File{}, id).Error
}

func (f *FileService) GetFileByID(id uint) (*model.File, error) {
	var file model.File
	if err := f.db.First(&file, id).Error; err != nil {
		return nil, err
	}
	if err := f.decryptFileMetadata(&file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *FileService) ListFiles(page, size int) (total int64, files []model.File, err error) {
	offset := (page - 1) * size
	if err = s.db.Model(&model.File{}).Count(&total).Error; err != nil {
		return
	}
	err = s.db.Order("created_at desc").Limit(size).Offset(offset).Find(&files).Error
	if err != nil {
		return
	}
	filtered := make([]model.File, 0, len(files))
	for i := range files {
		if derr := s.decryptFileMetadata(&files[i]); derr != nil {
			// Skip records encrypted with a different key or corrupted metadata,
			// so one bad row doesn't break the whole listing API.
			continue
		}
		filtered = append(filtered, files[i])
	}
	files = filtered
	return
}

func deriveKeys(base64Key string) ([]byte, []byte) {
	raw, err := base64.RawURLEncoding.DecodeString(base64Key)
	if err != nil || len(raw) < 32 {
		return nil, nil
	}

	fileKeyMAC := hmac.New(sha256.New, raw)
	fileKeyMAC.Write([]byte("file-gcm-aes256"))
	metaKeyMAC := hmac.New(sha256.New, raw)
	metaKeyMAC.Write([]byte("db-meta-gcm-aes256"))

	return fileKeyMAC.Sum(nil), metaKeyMAC.Sum(nil)
}

func (f *FileService) encryptToFile(src io.Reader, dstPath string) (int64, error) {
	if f.fileGCM == nil {
		return 0, errors.New("file crypto key not configured")
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	prefix := make([]byte, fileNoncePrefix)
	if _, err := rand.Read(prefix); err != nil {
		return 0, err
	}

	header := append([]byte(fileMagicV2), prefix...)
	if _, err := out.Write(header); err != nil {
		return 0, err
	}

	writer := bufio.NewWriterSize(out, chunkSize*2)
	buf := make([]byte, chunkSize)
	var total int64
	var counter uint32

	for {
		n, rerr := src.Read(buf)
		if n > 0 {
			total += int64(n)
			nonce := makeChunkNonce(prefix, counter)
			aad := makeChunkAAD(counter)
			sealed := f.fileGCM.Seal(nil, nonce, buf[:n], aad)

			if err := binary.Write(writer, binary.BigEndian, uint32(len(sealed))); err != nil {
				return 0, err
			}
			if _, err := writer.Write(sealed); err != nil {
				return 0, err
			}
			counter++
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return 0, rerr
		}
	}

	if err := writer.Flush(); err != nil {
		return 0, err
	}
	return total, nil
}

func (f *FileService) DecryptToWriter(w io.Writer, srcPath string) error {
	if f.fileGCM == nil {
		return errors.New("file crypto key not configured")
	}

	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()

	header := make([]byte, len(fileMagicV2)+fileNoncePrefix)
	if _, err := io.ReadFull(in, header); err != nil {
		return err
	}
	if string(header[:len(fileMagicV2)]) != fileMagicV2 {
		return errors.New("invalid file magic")
	}
	prefix := header[len(fileMagicV2):]

	reader := bufio.NewReaderSize(in, chunkSize*2)
	var counter uint32
	for {
		var n uint32
		err := binary.Read(reader, binary.BigEndian, &n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return errors.New("invalid encrypted chunk length")
		}

		sealed := make([]byte, n)
		if _, err := io.ReadFull(reader, sealed); err != nil {
			return err
		}

		nonce := makeChunkNonce(prefix, counter)
		aad := makeChunkAAD(counter)
		plain, err := f.fileGCM.Open(nil, nonce, sealed, aad)
		if err != nil {
			return errors.New("file integrity check failed")
		}
		if _, err := w.Write(plain); err != nil {
			return err
		}
		counter++
	}

	return nil
}

func makeChunkNonce(prefix []byte, counter uint32) []byte {
	nonce := make([]byte, fileNonceSize)
	copy(nonce, prefix)
	binary.BigEndian.PutUint32(nonce[fileNoncePrefix:], counter)
	return nonce
}

func makeChunkAAD(counter uint32) []byte {
	aad := make([]byte, 4)
	binary.BigEndian.PutUint32(aad, counter)
	return aad
}

func randomStorageName() (string, error) {
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randBytes) + ".bin", nil
}

func (f *FileService) encryptFileMetadata(file *model.File) error {
	if f.metaGCM == nil {
		return errors.New("file crypto key not configured")
	}
	var err error

	if file.EncFilename, err = f.encryptString(file.Filename); err != nil {
		return err
	}
	if file.EncStoragePath, err = f.encryptString(file.StoragePath); err != nil {
		return err
	}
	if file.EncSize, err = f.encryptString(strconv.FormatInt(file.Size, 10)); err != nil {
		return err
	}
	if file.EncDescription, err = f.encryptString(file.Description); err != nil {
		return err
	}
	if file.EncUploaderID, err = f.encryptString(file.UploaderID); err != nil {
		return err
	}
	file.LegacyFilename = ""
	file.LegacyPath = ""
	file.LegacySize = 0
	file.LegacyDesc = ""
	file.LegacyUploader = ""
	return nil
}

func (f *FileService) decryptFileMetadata(file *model.File) error {
	if f.metaGCM == nil {
		return errors.New("file crypto key not configured")
	}
	if file.EncFilename == "" && file.EncStoragePath == "" && file.EncSize == "" && file.EncDescription == "" && file.EncUploaderID == "" {
		file.Filename = file.LegacyFilename
		file.StoragePath = file.LegacyPath
		file.Size = file.LegacySize
		file.Description = file.LegacyDesc
		file.UploaderID = file.LegacyUploader
		return nil
	}
	var err error

	if file.Filename, err = f.decryptString(file.EncFilename); err != nil {
		return err
	}
	if file.StoragePath, err = f.decryptString(file.EncStoragePath); err != nil {
		return err
	}
	sizeText, err := f.decryptString(file.EncSize)
	if err != nil {
		return err
	}
	if sizeText == "" {
		file.Size = 0
	} else {
		size, convErr := strconv.ParseInt(sizeText, 10, 64)
		if convErr != nil {
			return convErr
		}
		file.Size = size
	}
	if file.Description, err = f.decryptString(file.EncDescription); err != nil {
		return err
	}
	if file.UploaderID, err = f.decryptString(file.EncUploaderID); err != nil {
		return err
	}
	return nil
}

func (f *FileService) encryptString(plain string) (string, error) {
	nonce := make([]byte, metaNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := f.metaGCM.Seal(nil, nonce, []byte(plain), nil)
	payload := append(nonce, sealed...)
	return "v1:" + base64.RawURLEncoding.EncodeToString(payload), nil
}

func (f *FileService) decryptString(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	if len(ciphertext) < 3 || ciphertext[:3] != "v1:" {
		return "", errors.New("invalid encrypted metadata format")
	}
	blob, err := base64.RawURLEncoding.DecodeString(ciphertext[3:])
	if err != nil {
		return "", err
	}
	if len(blob) < metaNonceSize {
		return "", errors.New("invalid encrypted metadata payload")
	}
	nonce := blob[:metaNonceSize]
	sealed := blob[metaNonceSize:]
	plain, err := f.metaGCM.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", errors.New("metadata integrity check failed")
	}
	return string(plain), nil
}
