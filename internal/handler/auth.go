package handler

import (
	"fmt"

	"github.com/Kaikai20040827/graduation/internal/config"
	"github.com/Kaikai20040827/graduation/internal/middleware"
	"github.com/Kaikai20040827/graduation/internal/pkg"
	"github.com/Kaikai20040827/graduation/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userSrv *service.UserService
	jwtCfg  *config.JWTConfig
}

func NewAuthHandler(usersrv *service.UserService, jwtcfg *config.JWTConfig) *AuthHandler {
	fmt.Println("✓ Creating a new authorization handler done")
	return &AuthHandler{
		userSrv: usersrv,
		jwtCfg:  jwtcfg,
	}
}

type RegisterReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Username string `json:"username" binding:"required"`
}

func (h *AuthHandler) Register(context *gin.Context) {
	var req RegisterReq
	if err := context.ShouldBindJSON(&req); err != nil {
		pkg.JSONError(context, 40001, "failed to bind")
		context.Abort()
		return
	}
	user, err := h.userSrv.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		pkg.JSONError(context, 40002, "failed to create user")
		context.Abort()
		return
	}
	pkg.JSONOK(context, user)
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.JSONError(c, 400, "invalid params")
		return
	}
	u, err := h.userSrv.Authenticate(req.Email, req.Password)
	if err != nil {
		pkg.JSONError(c, 401, "invalid credentials")
		return
	}
	user_id := u.ID
	token, err := middleware.GenerateToken(h.jwtCfg, uint(user_id))
	if err != nil {
		pkg.JSONError(c, 500, "token gen failed")
		return
	}

	setAuthTokenCookie(c, token, int(h.jwtCfg.ExpiryMinutes*60))

	pkg.JSONOK(c, gin.H{
		"token":   token,
		"expires": 0,
		"user":    u})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	setAuthTokenCookie(c, "", -1)
	pkg.JSONOK(c, gin.H{"message": "logged out"})
}

func setAuthTokenCookie(c *gin.Context, value string, maxAgeSeconds int) {
	c.SetCookie("token", value, maxAgeSeconds, "/", "", false, true)
}
