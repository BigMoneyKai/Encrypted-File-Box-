package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/Kaikai20040827/graduation/internal/pkg"
	"github.com/Kaikai20040827/graduation/internal/service"
	"github.com/gin-gonic/gin"
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
	f, err := fileHeader.Open()
	if err != nil {
		pkg.JSONError(c, 50001, "open file failed")
		return
	}
	defer f.Close()

	desc := c.PostForm("description")
	// save
	out, err := h.fileSrv.UploadFile(f, filepath.Base(fileHeader.Filename), uid, desc)
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
	f, err := fileHeader.Open()
	if err != nil {
		pkg.JSONError(c, 50001, "open file failed")
		return
	}
	defer f.Close()

	desc := c.PostForm("description")
	// use uploader id 0 for public uploads
	out, err := h.fileSrv.UploadFile(f, filepath.Base(fileHeader.Filename), 0, desc)
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
	fileHeader, err := c.FormFile("file")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			pkg.JSONError(c, 40001, "file required")
			return
		}
	} else {
		f, err := fileHeader.Open()
		if err != nil {
			pkg.JSONError(c, 50001, "open file failed")
			return
		}
		fileReader = f
		fileCloser = f
		filename := filepath.Base(fileHeader.Filename)
		filenamePtr = &filename
	}
	if fileCloser != nil {
		defer fileCloser.Close()
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

	out, err := h.fileSrv.UpdateFile(uint(id), fileReader, filenamePtr, descPtr)
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
