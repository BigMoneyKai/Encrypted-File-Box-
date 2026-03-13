package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kaikai20040827/graduation/internal/config"
	"github.com/Kaikai20040827/graduation/internal/middleware"
	"github.com/gin-gonic/gin"
)

func TestJWTAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.JWTConfig{
		Secret:        "0123456789abcdef0123456789abcdef",
		Issuer:        "iss",
		Audience:      "aud",
		ExpiryMinutes: 10,
	}
	token, err := middleware.GenerateToken(cfg, 42)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	router := gin.New()
	router.Use(middleware.JWTAuthMiddleware(cfg))
	router.GET("/protected", func(c *gin.Context) {
		uid, _ := c.Get("user_id")
		c.JSON(200, gin.H{"uid": uid})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != 401 {
		t.Fatalf("expected 401, got %d", w2.Code)
	}
}
