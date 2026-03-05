package handler

import (
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Kaikai20040827/graduation/internal/model"
	"github.com/Kaikai20040827/graduation/internal/pkg"
	"github.com/Kaikai20040827/graduation/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSrv *service.UserService
	fileSrv *service.FileService
}

func NewUserHandler(us *service.UserService, fs *service.FileService) *UserHandler {
	fmt.Println("✓ Creating a new user handler done")
	return &UserHandler{
		userSrv: us,
		fileSrv: fs,
	}
}

func (uh *UserHandler) GetProfile(context *gin.Context) {
	uidv, ok := context.Get("user_id")
	if !ok {
		pkg.JSONError(context, 401, "unauthorized")
		context.Abort()
		return
	}
	uid := uidv.(uint)

	user, err := uh.userSrv.GetByID(uid)
	if err != nil {
		pkg.JSONError(context, 404, "cannot find user")
		context.Abort()
		return
	}
	pkg.JSONOK(context, uh.profileResponse(user))
}

type UpdateProfileReq struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"`
}

func (uh *UserHandler) UpdateProfile(context *gin.Context) {
	var req UpdateProfileReq
	if err := context.ShouldBindBodyWithJSON(&req); err != nil {
		pkg.JSONError(context, 40001, "invalid params")
		return
	}

	uidv, _ := context.Get("user_id")
	uid := uidv.(uint)
	u, err := uh.userSrv.UpdateProfile(uid, req.Username, req.Email)
	if err != nil {
		pkg.JSONError(context, 50001, err.Error())
		return
	}
	pkg.JSONOK(context, uh.profileResponse(u))
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (uh *UserHandler) ChangePassword(context *gin.Context) {
	var req ChangePasswordReq
	if err := context.ShouldBindBodyWithJSON(&req); err != nil {
		pkg.JSONError(context, 40001, "invalid params")
		return
	}

	uidv, ok := context.Get("user_id")
	if !ok {
		pkg.JSONError(context, 401, "unauthorized")
		return
	}
	uid := uidv.(uint)

	if err := uh.userSrv.ChangePassword(uid, req.OldPassword, req.NewPassword); err != nil {
		pkg.JSONError(context, 40002, err.Error())
		return
	}
	context.Status(http.StatusNoContent)
}

const maxAvatarSize = 5 * 1024 * 1024

type ProfileResp struct {
	ID              uint       `json:"id"`
	Email           string     `json:"email"`
	Username        string     `json:"username"`
	AvatarURL       string     `json:"avatar_url,omitempty"`
	AvatarUpdatedAt *time.Time `json:"avatar_updated_at,omitempty"`
}

func (uh *UserHandler) profileResponse(u *model.User) ProfileResp {
	resp := ProfileResp{
		ID:       u.ID,
		Email:    u.Email,
		Username: u.Username,
	}
	if u.AvatarPath != "" {
		resp.AvatarURL = "/api/v1/user/avatar"
		if u.AvatarUpdatedAt != nil {
			resp.AvatarUpdatedAt = u.AvatarUpdatedAt
			resp.AvatarURL = resp.AvatarURL + "?ts=" + strconv.FormatInt(u.AvatarUpdatedAt.Unix(), 10)
		}
	}
	return resp
}

func (uh *UserHandler) UpdateAvatar(context *gin.Context) {
	uidv, ok := context.Get("user_id")
	if !ok {
		pkg.JSONError(context, 401, "unauthorized")
		return
	}
	uid, ok := uidv.(uint)
	if !ok {
		pkg.JSONError(context, 401, "invalid user token")
		return
	}

	fileHeader, err := context.FormFile("avatar")
	if err != nil {
		fileHeader, err = context.FormFile("file")
	}
	if err != nil {
		pkg.JSONError(context, 40001, "avatar file required")
		return
	}
	if fileHeader.Size > maxAvatarSize {
		pkg.JSONError(context, 40001, "avatar file too large")
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(fileHeader.Filename))
	}
	if contentType == "" || !strings.HasPrefix(contentType, "image/") {
		pkg.JSONError(context, 40001, "only image avatars are supported")
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		pkg.JSONError(context, 50001, "open file failed")
		return
	}
	defer src.Close()

	user, err := uh.userSrv.GetByID(uid)
	if err != nil {
		pkg.JSONError(context, 404, "cannot find user")
		return
	}

	storedPath, _, err := uh.fileSrv.SaveUserAvatar(src, filepath.Base(fileHeader.Filename), uid)
	if err != nil {
		pkg.JSONError(context, 50002, err.Error())
		return
	}

	if err := uh.fileSrv.RemoveStoredFile(user.AvatarPath); err != nil {
		_ = uh.fileSrv.RemoveStoredFile(storedPath)
		pkg.JSONError(context, 50002, err.Error())
		return
	}

	u, err := uh.userSrv.UpdateAvatar(uid, storedPath, contentType)
	if err != nil {
		_ = uh.fileSrv.RemoveStoredFile(storedPath)
		pkg.JSONError(context, 50002, err.Error())
		return
	}
	pkg.JSONOK(context, uh.profileResponse(u))
}

func (uh *UserHandler) GetAvatar(context *gin.Context) {
	uidv, ok := context.Get("user_id")
	if !ok {
		pkg.JSONError(context, 401, "unauthorized")
		return
	}
	uid, ok := uidv.(uint)
	if !ok {
		pkg.JSONError(context, 401, "invalid user token")
		return
	}

	user, err := uh.userSrv.GetByID(uid)
	if err != nil {
		pkg.JSONError(context, 404, "cannot find user")
		return
	}
	if user.AvatarPath == "" {
		pkg.JSONError(context, 404, "avatar not found")
		return
	}

	if user.AvatarMime != "" {
		context.Header("Content-Type", user.AvatarMime)
	}
	context.Header("Cache-Control", "no-store")
	if err := uh.fileSrv.DecryptToWriter(context.Writer, user.AvatarPath); err != nil {
		pkg.JSONError(context, 50002, err.Error())
		return
	}
}

type DeleteUserReq struct {
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

func (uh *UserHandler) DeleteUser(context *gin.Context) {
	var req DeleteUserReq
	if err := context.ShouldBindBodyWithJSON(&req); err != nil {
		pkg.JSONError(context, 40001, "invalid params")
		return
	}

	passwordv, _ := context.Get("password")
	password := passwordv.(string)

	emailv, _ := context.Get("email")
	email := emailv.(string)

	if err := uh.userSrv.DeleteUser(email, password); err != nil {
		pkg.JSONError(context, 40002, err.Error())
		return
	}
	context.Status(http.StatusNoContent) //后续开发跳转网页
}
