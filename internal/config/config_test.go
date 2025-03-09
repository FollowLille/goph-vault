package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/FollowLille/goph-vault/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadServerConfig(t *testing.T) {
	// Создаем временный файл с тестовой конфигурацией
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "server_config.json")
	configData := `{
        "jwt_key": "test-jwt-key",
        "encryption_key": "test-encryption-key",
        "address": "localhost:8080",
        "database": {
            "host": "localhost",
            "port": 5432,
            "user": "test-user",
            "password": "test-password",
            "dbname": "test-db"
        },
        "tls_key_file": "key.pem",
        "tls_cert_file": "cert.pem",
        "log_level": "debug",
        "address_https": "localhost:8443"
    }`
	err := os.WriteFile(configPath, []byte(configData), 0600)
	assert.NoError(t, err)

	// Загружаем конфигурацию
	cfg, err := config.LoadServerConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Проверяем значения
	assert.Equal(t, "test-jwt-key", cfg.JWTKey)
	assert.Equal(t, "test-encryption-key", cfg.EncryptionKey)
	assert.Equal(t, "localhost:8080", cfg.Address)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "test-user", cfg.Database.User)
	assert.Equal(t, "test-password", cfg.Database.Password)
	assert.Equal(t, "test-db", cfg.Database.DBName)
	assert.Equal(t, "key.pem", cfg.TLSKeyFile)
	assert.Equal(t, "cert.pem", cfg.TLSCertFile)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "localhost:8443", cfg.AddressHTTPS)
}

func TestLoadClientConfig_ExistingFile(t *testing.T) {
	// Создаем временный файл с тестовой конфигурацией
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	configPath := filepath.Join(tempDir, ".gophvault", "client_config.json")
	configData := `{
        "server_address": "http://example.com",
        "server_address_https": "https://example.com",
        "allow_http": true,
        "allow_insecure_tls": false,
        "use_gzip": true,
        "encryption_key": "abcdef1234567890abcdef1234567890",
        "log_level": "info"
    }`
	err := os.MkdirAll(filepath.Dir(configPath), 0700)
	assert.NoError(t, err)
	err = os.WriteFile(configPath, []byte(configData), 0600)
	assert.NoError(t, err)

	// Загружаем конфигурацию
	cfg, err := config.LoadClientConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Проверяем значения
	assert.Equal(t, "http://example.com", cfg.ServerAddress)
	assert.Equal(t, "https://example.com", cfg.ServerAddressHttps)
	assert.True(t, cfg.AllowHTTP)
	assert.False(t, cfg.AllowInsecureTLS)
	assert.True(t, cfg.UseGzip)
	assert.Equal(t, "abcdef1234567890abcdef1234567890", cfg.EncryptionKey)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLoadClientConfig_DefaultFile(t *testing.T) {
	// Устанавливаем временную директорию HOME
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Убедимся, что файла конфигурации нет
	configPath := filepath.Join(tempDir, ".gophvault", "client_config.json")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("Config file already exists: %s", configPath)
	}

	// Загружаем конфигурацию (должна быть создана по умолчанию)
	cfg, err := config.LoadClientConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Проверяем значения по умолчанию
	assert.Equal(t, "http://localhost:8080", cfg.ServerAddress)
	assert.Equal(t, "https://localhost:8443", cfg.ServerAddressHttps)
	assert.True(t, cfg.AllowHTTP)
	assert.False(t, cfg.AllowInsecureTLS)
	assert.False(t, cfg.UseGzip)
	assert.Equal(t, "1234567890abcdef1234567890abcdef", cfg.EncryptionKey)
	assert.Equal(t, "info", cfg.LogLevel)

	// Проверяем, что файл был создан
	_, err = os.Stat(configPath)
	assert.NoError(t, err)
}
