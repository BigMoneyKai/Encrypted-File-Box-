package test

import (
	"net/http/httptest"
	"testing"

	"github.com/Kaikai20040827/graduation/internal/pkg"
	"github.com/gin-gonic/gin"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := pkg.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if err := pkg.CheckPassword(hash, "secret"); err != nil {
		t.Fatalf("CheckPassword: %v", err)
	}
	if err := pkg.CheckPassword(hash, "wrong"); err == nil {
		t.Fatalf("expected wrong password error")
	}
}

func TestGetPageParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?page=0&size=200", nil)
	page, size := pkg.GetPageParams(c)
	if page != 1 || size != 20 {
		t.Fatalf("expected defaults, got page=%d size=%d", page, size)
	}
}
