package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/client"
	"github.com/FollowLille/goph-vault/internal/logger"
)

// Login выполняет авторизацию пользователя
func Login(client *client.Client, username, password string) error {
	// Подготавливаем тело запроса
	requestBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		logger.Log.Error("Failed to marshal request body", zap.Error(err))
		return fmt.Errorf("failed to prepare login request")
	}

	// Отправляем запрос на сервер
	resp, err := client.SendRequest("POST", "/api/login", requestBody, true)
	if err != nil {
		logger.Log.Error("Failed to send request", zap.Error(err))
		return fmt.Errorf("failed to send login request")
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("Failed to read response body", zap.Error(err))
		return fmt.Errorf("failed to read server response")
	}

	// Обрабатываем успешный ответ
	if resp.StatusCode == http.StatusOK {
		var tokenResponse struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(body, &tokenResponse); err != nil {
			logger.Log.Error("Failed to parse server response", zap.Error(err))
			return fmt.Errorf("failed to parse server response")
		}

		// Сохраняем токен
		if err := client.AuthProvider.SaveToken(tokenResponse.Token, client.EncryptionKey); err != nil {
			logger.Log.Error("Failed to save token", zap.Error(err))
			return fmt.Errorf("failed to save token")
		}

		return nil // Успешная авторизация
	}

	// Обрабатываем ошибки
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("invalid credentials")
	default:
		return fmt.Errorf("failed to log in: %s", string(body))
	}
}

// Register выполняет регистрацию нового пользователя
func Register(client *client.Client, username, password string) error {
	// Подготавливаем тело запроса
	requestBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return fmt.Errorf("failed to prepare registration request: %w", err)
	}

	// Отправляем запрос на сервер
	resp, err := client.SendRequest("POST", "/api/register", requestBody, true)
	if err != nil {
		return fmt.Errorf("failed to send registration request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read server response: %w", err)
	}

	// Обрабатываем успешный ответ
	if resp.StatusCode == http.StatusOK {
		return nil // Успешная регистрация
	}

	// Обрабатываем ошибки
	switch resp.StatusCode {
	case http.StatusConflict:
		return fmt.Errorf("username already exists")
	default:
		return fmt.Errorf("failed to register user: %s", string(body))
	}
}
