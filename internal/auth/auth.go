package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/crypto"
	"github.com/FollowLille/goph-vault/internal/logger"
)

// Интерфейс для сервиса аутентификации
type AuthProvider interface {
	SaveToken(token string, encryptionKey string) error
	LoadToken(encryptionKey string) (string, error)
	ValidatePassword(password string) error
	DeleteToken() error
}

// Сервис аутентификации
type AuthService struct {
	encryptionKey string
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(encryptionKey string) *AuthService {
	return &AuthService{encryptionKey: encryptionKey}
}

// SaveToken сохраняет токен в файл
func (a *AuthService) SaveToken(token string, encryptionKey string) error {
	// Удаляем старый токен
	if err := a.DeleteToken(); err != nil {
		logger.Log.Warn("Failed to delete old token", zap.Error(err))
	}

	// Шифруем новый токен
	encryptedToken, err := crypto.Encrypt([]byte(token), []byte(encryptionKey))
	if err != nil {
		logger.Log.Error("Failed to encrypt token", zap.Error(err))
		return err
	}

	// Преобразуем зашифрованный токен в Base64
	encodedToken := base64.StdEncoding.EncodeToString(encryptedToken)

	authData := struct {
		Token string `json:"token"`
	}{Token: string(encodedToken)}

	data, err := json.MarshalIndent(authData, "", "  ")
	if err != nil {
		return err
	}

	authFilePath := filepath.Join(os.Getenv("HOME"), ".gophvault", "auth.json")
	if err := os.MkdirAll(filepath.Dir(authFilePath), 0700); err != nil {
		return err
	}

	return os.WriteFile(authFilePath, data, 0600)
}

// LoadToken загружает токен из файла
func (a *AuthService) LoadToken(encryptionKey string) (string, error) {
	authFilePath := filepath.Join(os.Getenv("HOME"), ".gophvault", "auth.json")
	data, err := os.ReadFile(authFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Log.Warn("No token found. Proceeding without authentication.")
			return "", nil
		}
		logger.Log.Warn("Failed to read auth file", zap.Error(err))
		return "", err
	}
	var authData struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &authData); err != nil {
		return "", err
	}

	decodedToken, err := base64.StdEncoding.DecodeString(authData.Token)
	if err != nil {
		logger.Log.Error("Failed to decode token from Base64", zap.Error(err))
		return "", err
	}
	// Расшифровываем токен
	token, err := crypto.Decrypt(decodedToken, []byte(encryptionKey))
	if err != nil {
		logger.Log.Error("Failed to decrypt token", zap.Error(err))
		return "", err
	}

	return string(token), nil
}

// DeleteToken удаляет файл с токеном
func (a *AuthService) DeleteToken() error {
	authFilePath := filepath.Join(os.Getenv("HOME"), ".gophvault", "auth.json")
	if _, err := os.Stat(authFilePath); os.IsNotExist(err) {
		return nil // Файл не существует, ничего удалять не нужно
	}

	if err := os.Remove(authFilePath); err != nil {
		logger.Log.Error("Failed to delete auth file", zap.Error(err))
		return err
	}

	return nil
}

// ValidatePassword валидирует пароль
func (a *AuthService) ValidatePassword(password string) error {
	const minPasswordLength = 8
	var errors []string
	if len(password) < minPasswordLength {
		errors = append(errors, fmt.Sprintf("password must be at least %d characters long", minPasswordLength))
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		errors = append(errors, "password must contain at least one uppercase letter")
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		errors = append(errors, "password must contain at least one lowercase letter")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		errors = append(errors, "password must contain at least one number")
	}
	if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		errors = append(errors, "password must contain at least one special character")
	}
	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}
