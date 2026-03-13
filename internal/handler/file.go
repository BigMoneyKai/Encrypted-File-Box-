package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Kaikai20040827/graduation/internal/pkg"
	"github.com/Kaikai20040827/graduation/internal/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type FileHandler struct {
	fileSrv *service.FileService
}

func NewFileHandler(fs *service.FileService) *FileHandler {
	fmt.Println("✓ Creating a new file handler done")
	return &FileHandler{fileSrv: fs}
}

// 以下代码可能存在漏洞，需要检查
// Upload
func (h *FileHandler) UploadFile(c *gin.Context) {
	uidv, ok := c.Get("user_id")
	if !ok {
		pkg.JSONError(c, 401, "unauthorized")
		return
	}
	uid, ok := uidv.(uint)
	if !ok {
		pkg.JSONError(c, 401, "invalid user token")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		pkg.JSONError(c, 40001, "file required")
		return
	}

	desc := c.PostForm("description")
	// save
	out, err := h.fileSrv.UploadValidatedMultipart(c.Request.Context(), fileHeader, uid, desc)
	if err != nil {
		pkg.JSONError(c, 50002, err.Error())
		return
	}

	pkg.JSONOK(c, gin.H{
		"file_id":  out.ID,
		"filename": out.Filename,
		"size":     out.Size,
		"url":      "/api/v1/files/download/" + strconv.FormatUint(uint64(out.ID), 10),
	})
}

// UploadFilePublic allows anonymous/public uploads (no JWT required).
func (h *FileHandler) UploadFilePublic(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		pkg.JSONError(c, 40001, "file required")
		return
	}

	desc := c.PostForm("description")
	// use uploader id 0 for public uploads
	out, err := h.fileSrv.UploadValidatedMultipart(c.Request.Context(), fileHeader, 0, desc)
	if err != nil {
		pkg.JSONError(c, 50002, err.Error())
		return
	}
	pkg.JSONOK(c, gin.H{
		"file_id":  out.ID,
		"filename": out.Filename,
		"size":     out.Size,
		"url":      "/api/v1/files/download/" + strconv.FormatUint(uint64(out.ID), 10),
	})
}

// UpdateFile replaces file content and/or updates description.
func (h *FileHandler) UpdateFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		pkg.JSONError(c, 40001, "invalid file id")
		return
	}

	var fileReader io.Reader
	var fileCloser io.Closer
	var filenamePtr *string
	var mimePtr *string
	var tmpPath string
	fileHeader, err := c.FormFile("file")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			pkg.JSONError(c, 40001, "file required")
			return
		}
	} else {
		tmp, filename, mimeType, err := h.fileSrv.PrepareValidatedTemp(c.Request.Context(), fileHeader)
		if err != nil {
			pkg.JSONError(c, 40001, err.Error())
			return
		}
		tmpPath = tmp
		f, err := os.Open(tmpPath)
		if err != nil {
			pkg.JSONError(c, 50001, "open file failed")
			return
		}
		fileReader = f
		fileCloser = f
		filenamePtr = &filename
		mimePtr = &mimeType
	}
	if fileCloser != nil {
		defer fileCloser.Close()
		defer os.Remove(tmpPath)
	}

	desc, descOK := c.GetPostForm("description")
	var descPtr *string
	if descOK {
		descPtr = &desc
	}
	if fileReader == nil && descPtr == nil {
		pkg.JSONError(c, 40001, "nothing to update")
		return
	}

	out, err := h.fileSrv.UpdateFile(uint(id), fileReader, filenamePtr, descPtr, mimePtr)
	if err != nil {
		pkg.JSONError(c, 50002, err.Error())
		return
	}

	pkg.JSONOK(c, gin.H{
		"file_id":  out.ID,
		"filename": out.Filename,
		"size":     out.Size,
		"url":      "/api/v1/files/download/" + strconv.FormatUint(uint64(out.ID), 10),
	})
}

// List
func (h *FileHandler) ListFiles(c *gin.Context) {
	page, size := pkg.GetPageParams(c)
	total, files, err := h.fileSrv.ListFiles(page, size)
	if err != nil {
		pkg.JSONError(c, 50001, err.Error())
		return
	}
	pkg.JSONOK(c, gin.H{"total": total, "items": files})
}

// Download
func (h *FileHandler) DownloadFile(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	f, err := h.fileSrv.GetFileByID(uint(id))
	if err != nil {
		pkg.JSONError(c, 404, "file not found")
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+f.Filename+"\"")
	if err := h.fileSrv.DecryptToWriter(c.Writer, f.StoragePath); err != nil {
		pkg.JSONError(c, 50002, err.Error())
		return
	}
}

// Delete
func (h *FileHandler) DeleteFile(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	if err := h.fileSrv.DeleteFile(uint(id)); err != nil {
		pkg.JSONError(c, 50001, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// Batch upload
func (h *FileHandler) UploadFileBatch(c *gin.Context) {
	uidv, ok := c.Get("user_id")
	if !ok {
		pkg.JSONError(c, 401, "unauthorized")
		return
	}
	uid, ok := uidv.(uint)
	if !ok {
		pkg.JSONError(c, 401, "invalid user token")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		pkg.JSONError(c, 40001, "invalid multipart form")
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		pkg.JSONError(c, 40001, "files required")
		return
	}
	descs := form.Value["descriptions"]
	results := make([]gin.H, 0, len(files))
	for i, fh := range files {
		desc := ""
		if i < len(descs) {
			desc = descs[i]
		}
		out, err := h.fileSrv.UploadValidatedMultipart(c.Request.Context(), fh, uid, desc)
		if err != nil {
			results = append(results, gin.H{
				"filename": filepath.Base(fh.Filename),
				"error":    err.Error(),
			})
			continue
		}
		results = append(results, gin.H{
			"file_id":  out.ID,
			"filename": out.Filename,
			"size":     out.Size,
			"mime":     out.Mime,
		})
	}
	pkg.JSONOK(c, gin.H{"items": results})
}

// Batch delete
func (h *FileHandler) DeleteFileBatch(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		pkg.JSONError(c, 40001, "ids required")
		return
	}
	results := make([]gin.H, 0, len(req.IDs))
	for _, id := range req.IDs {
		if err := h.fileSrv.DeleteFile(id); err != nil {
			results = append(results, gin.H{"id": id, "status": "failed", "error": err.Error()})
			continue
		}
		results = append(results, gin.H{"id": id, "status": "deleted"})
	}
	pkg.JSONOK(c, gin.H{"items": results})
}

// Resumable upload: init
func (h *FileHandler) InitResumable(c *gin.Context) {
	uidv, ok := c.Get("user_id")
	if !ok {
		pkg.JSONError(c, 401, "unauthorized")
		return
	}
	uid, ok := uidv.(uint)
	if !ok {
		pkg.JSONError(c, 401, "invalid user token")
		return
	}

	var req struct {
		Filename    string `json:"filename"`
		TotalSize   int64  `json:"total_size"`
		ChunkSize   int64  `json:"chunk_size"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.JSONError(c, 40001, "invalid request")
		return
	}
	session, err := h.fileSrv.InitResumableUpload(req.Filename, req.TotalSize, req.ChunkSize, uid, req.Description)
	if err != nil {
		pkg.JSONError(c, 40001, err.Error())
		return
	}
	pkg.JSONOK(c, session)
}

// Resumable upload: status
func (h *FileHandler) GetResumableStatus(c *gin.Context) {
	uploadID := c.Param("upload_id")
	session, err := h.fileSrv.GetUploadSession(uploadID)
	if err != nil {
		pkg.JSONError(c, 404, err.Error())
		return
	}
	uploaded, err := h.fileSrv.ListUploadedChunks(uploadID)
	if err != nil {
		pkg.JSONError(c, 50001, err.Error())
		return
	}
	pkg.JSONOK(c, gin.H{
		"upload_id":    session.UploadID,
		"total_size":   session.TotalSize,
		"chunk_size":   session.ChunkSize,
		"total_chunks": session.TotalChunks,
		"uploaded":     uploaded,
		"status":       session.Status,
	})
}

// Resumable upload: chunk
func (h *FileHandler) UploadChunk(c *gin.Context) {
	uploadID := c.Param("upload_id")
	indexStr := c.PostForm("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		pkg.JSONError(c, 40001, "invalid chunk index")
		return
	}
	fileHeader, err := c.FormFile("chunk")
	if err != nil {
		pkg.JSONError(c, 40001, "chunk required")
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		pkg.JSONError(c, 50001, "open chunk failed")
		return
	}
	defer f.Close()

	if err := h.fileSrv.SaveUploadChunk(c.Request.Context(), uploadID, index, f); err != nil {
		pkg.JSONError(c, 50001, err.Error())
		return
	}
	pkg.JSONOK(c, gin.H{"upload_id": uploadID, "index": index, "status": "ok"})
}

// Resumable upload: complete
func (h *FileHandler) CompleteResumable(c *gin.Context) {
	uploadID := c.Param("upload_id")
	file, err := h.fileSrv.CompleteUpload(c.Request.Context(), uploadID)
	if err != nil {
		pkg.JSONError(c, 50002, err.Error())
		return
	}
	pkg.JSONOK(c, gin.H{
		"file_id":  file.ID,
		"filename": file.Filename,
		"size":     file.Size,
		"mime":     file.Mime,
	})
}

// Resumable upload: abort
func (h *FileHandler) AbortResumable(c *gin.Context) {
	uploadID := c.Param("upload_id")
	if err := h.fileSrv.AbortUpload(uploadID); err != nil {
		pkg.JSONError(c, 50001, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// Preview
func (h *FileHandler) PreviewFile(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	f, err := h.fileSrv.GetFileByID(uint(id))
	if err != nil {
		pkg.JSONError(c, 404, "file not found")
		return
	}
	ext := service.NormalizeExt(f.Filename)
	if !isPreviewable(ext) {
		pkg.JSONError(c, 40001, "preview not supported")
		return
	}

	if isTextPreviewable(ext) {
		data, err := h.fileSrv.DecryptToBytesLimit(f.StoragePath, 2*1024*1024)
		if err != nil {
			pkg.JSONError(c, 50002, err.Error())
			return
		}
		text := decodeTextPreview(data)
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Header("Content-Disposition", "inline; filename=\""+safeInlineName(f.Filename)+"\"")
		c.String(http.StatusOK, text)
		return
	}

	if isOfficePreviewable(ext) {
		pdfPath, cleanup, err := h.convertOfficeToPDF(f.StoragePath, f.Filename)
		if err != nil {
			pkg.JSONError(c, 40001, err.Error())
			return
		}
		defer cleanup()
		pdfName := strings.TrimSuffix(safeInlineName(f.Filename), filepath.Ext(f.Filename)) + ".pdf"
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "inline; filename=\""+pdfName+"\"")
		c.File(pdfPath)
		return
	}

	mimeType := f.Mime
	if mimeType == "" {
		mimeType = service.GuessMimeByExtension(f.Filename)
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", "inline; filename=\""+f.Filename+"\"")
	if err := h.fileSrv.DecryptToWriter(c.Writer, f.StoragePath); err != nil {
		pkg.JSONError(c, 50002, err.Error())
		return
	}
}

func isPreviewable(ext string) bool {
	switch ext {
	case "jpg", "jpeg", "png", "gif", "webp",
		"txt", "md", "json", "log", "csv",
		"pdf",
		"doc", "docx", "xls", "xlsx", "ppt", "pptx":
		return true
	default:
		return false
	}
}

func isTextPreviewable(ext string) bool {
	switch ext {
	case "txt", "md", "json", "log", "csv":
		return true
	default:
		return false
	}
}

func isOfficePreviewable(ext string) bool {
	switch ext {
	case "doc", "docx", "xls", "xlsx", "ppt", "pptx":
		return true
	default:
		return false
	}
}

func safeInlineName(name string) string {
	base := filepath.Base(name)
	if base == "." || base == "/" || base == "" {
		return "preview"
	}
	return base
}

func decodeTextPreview(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	if bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		return string(data[3:])
	}
	if bytes.HasPrefix(data, []byte{0xFF, 0xFE}) || bytes.HasPrefix(data, []byte{0xFE, 0xFF}) {
		enc := unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM)
		if bytes.HasPrefix(data, []byte{0xFE, 0xFF}) {
			enc = unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM)
		}
		if decoded, err := decodeWith(enc, data); err == nil {
			return decoded
		}
	}
	if utf8.Valid(data) {
		return string(data)
	}
	if decoded, err := decodeWith(simplifiedchinese.GB18030, data); err == nil {
		return decoded
	}
	if decoded, err := decodeWith(charmap.ISO8859_1, data); err == nil {
		return decoded
	}
	return string(data)
}

func decodeWith(enc encoding.Encoding, data []byte) (string, error) {
	r := transform.NewReader(bytes.NewReader(data), enc.NewDecoder())
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (h *FileHandler) convertOfficeToPDF(storagePath, filename string) (string, func(), error) {
	bin := lookupOfficeConverter()
	if bin == "" {
		return "", nil, errors.New("office preview unavailable (libreoffice not installed)")
	}
	tmpDir, err := os.MkdirTemp("", "sfb_preview_")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	baseName := safeInlineName(filename)
	if filepath.Ext(baseName) == "" {
		baseName = baseName + ".bin"
	}
	inputPath := filepath.Join(tmpDir, baseName)
	outFile, err := os.Create(inputPath)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	if err := h.fileSrv.DecryptToWriter(outFile, storagePath); err != nil {
		outFile.Close()
		cleanup()
		return "", nil, err
	}
	_ = outFile.Close()

	cmd := exec.Command(bin, "--headless", "--convert-to", "pdf", "--outdir", tmpDir, inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("office conversion failed: %s", strings.TrimSpace(string(output)))
	}

	pdfName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath)) + ".pdf"
	pdfPath := filepath.Join(tmpDir, pdfName)
	if _, err := os.Stat(pdfPath); err != nil {
		cleanup()
		return "", nil, errors.New("office conversion failed: pdf not generated")
	}
	return pdfPath, cleanup, nil
}

func lookupOfficeConverter() string {
	if p, _ := exec.LookPath("libreoffice"); p != "" {
		return p
	}
	if p, _ := exec.LookPath("soffice"); p != "" {
		return p
	}
	return ""
}
