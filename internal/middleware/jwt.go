package middleware

import (
	"strings"
	"time"

	"github.com/Kaikai20040827/graduation/internal/config"
	"github.com/Kaikai20040827/graduation/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(cfg *config.JWTConfig, user_id uint) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(cfg.ExpiryMinutes) * time.Minute)

	claims := JWTClaims{
		UserID: user_id,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Audience:  jwt.ClaimStrings{cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(cfg.Secret))
}

func JWTAuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(context *gin.Context) {
		if context.Request.URL.Path == "/login" ||
           context.Request.URL.Path == "/register" ||
           context.Request.URL.Path == "/" {
            context.Next()
            return
        }

		auth := context.GetHeader("authorization")
		if auth == "" {
			pToken, _ := context.Cookie("token")
			auth = pToken
		}
		if auth == "" {
			pkg.JSONError(context, 401, "missing token")
			context.Abort()
			return
		}

		//去掉前缀 "Bearer "
		auth = strings.TrimPrefix(auth, "Bearer ")

		token, err := jwt.ParseWithClaims(
			auth,
			&JWTClaims{},
			func(t *jwt.Token) (interface{}, error) {
				return []byte(cfg.Secret), nil
			})

		if err != nil || !token.Valid {
			pkg.JSONError(context, 401, "invalid token")
			context.Abort()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)

		if !ok {
			pkg.JSONError(context, 401, "invalid token claim")
			context.Abort()
			return
		}
		context.Set("user_id", claims.UserID)
		context.Next()
	}
}

// JWTAuthOrRedirect is for browser HTML routes: redirect to /login if token missing/invalid.
func JWTAuthOrRedirect(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(context *gin.Context) {
		auth := context.GetHeader("authorization")
		if auth == "" {
			pToken, _ := context.Cookie("token")
			auth = pToken
		}
		if auth == "" {
			context.Redirect(302, "/login")
			context.Abort()
			return
		}

		// 去掉前缀 "Bearer "
		auth = strings.TrimPrefix(auth, "Bearer ")

		token, err := jwt.ParseWithClaims(
			auth,
			&JWTClaims{},
			func(t *jwt.Token) (interface{}, error) {
				return []byte(cfg.Secret), nil
			},
		)

		if err != nil || !token.Valid {
			context.Redirect(302, "/login")
			context.Abort()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			context.Redirect(302, "/login")
			context.Abort()
			return
		}

		context.Set("user_id", claims.UserID)
		context.Next()
	}
}
