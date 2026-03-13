package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kaikai20040827/graduation/internal/config"
	"github.com/spf13/viper"
)

func TestGenerateJWTSecretLength(t *testing.T) {
	secret, err := config.GenerateJWTSecret(32)
	if err != nil {
		t.Fatalf("GenerateJWTSecret: %v", err)
	}
	if len(secret) < 32 {
		t.Fatalf("secret too short: %d", len(secret))
	}
}

func TestEnsureJWTSecretWritesConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("jwt:\n  secret: PLEASE_CHANGE_ME_32_CHARS_MINIMUM\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	v := viper.New()
	v.SetConfigFile(cfgPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("read config: %v", err)
	}

	if err := config.EnsureJWTSecret(v); err != nil {
		t.Fatalf("EnsureJWTSecret: %v", err)
	}

	b, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config after: %v", err)
	}
	if strings.Contains(string(b), "PLEASE_CHANGE_ME_32_CHARS_MINIMUM") {
		t.Fatalf("secret not replaced")
	}
}

func TestEnsureFileCryptoKeyWritesConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("file_crypto:\n  key: PLEASE_CHANGE_ME_32_CHARS_MINIMUM\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	v := viper.New()
	v.SetConfigFile(cfgPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("read config: %v", err)
	}

	if err := config.EnsureFileCryptoKey(v); err != nil {
		t.Fatalf("EnsureFileCryptoKey: %v", err)
	}

	b, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config after: %v", err)
	}
	if strings.Contains(string(b), "PLEASE_CHANGE_ME_32_CHARS_MINIMUM") {
		t.Fatalf("key not replaced")
	}
}
