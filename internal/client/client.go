package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/auth"
	"github.com/FollowLille/goph-vault/internal/local"
	"github.com/FollowLille/goph-vault/internal/logger"
	"github.com/FollowLille/goph-vault/internal/models"
	"github.com/FollowLille/goph-vault/internal/sync"
)

// Client представляет собой HTTP-клиент для взаимодействия с сервером
type Client struct {
	ServerAddress    string
	AllowHTTP        bool
	AllowInsecureTLS bool
	UseGzip          bool
	EncryptionKey    string
	AuthProvider     auth.AuthProvider
	httpClient       *http.Client
	SyncService      *sync.SyncService
}

// NewClient создает новый экземпляр Client
func NewClient(
	serverAddress string,
	allowHTTP bool,
	allowInsecureTLS bool,
	useGzip bool,
	encryptionKey string,
	authProvider auth.AuthProvider,
) *Client {
	client := &Client{
		ServerAddress:    serverAddress,
		AllowHTTP:        allowHTTP,
		AllowInsecureTLS: allowInsecureTLS,
		UseGzip:          useGzip,
		EncryptionKey:    encryptionKey,
		AuthProvider:     authProvider,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: allowInsecureTLS,
				},
			},
			Timeout: time.Second * 5,
		},
	}

	// Инициализация SyncService
	client.SyncService = sync.NewSyncService(client)

	return client
}

// SendRequest отправляет HTTP-запрос на сервер
func (c *Client) SendRequest(method, url string, data []byte, skipSync bool) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Таймаут 5 секунд
	defer cancel()
	// Загружаем токен
	token, err := c.AuthProvider.LoadToken(c.EncryptionKey)
	if err != nil {
		// Если токен отсутствует, это нормально для новых пользователей
		logger.Log.Warn("No token found. Proceeding without authentication.", zap.Error(err))
		token = ""
	}

	// Синхронизация локальных секретов перед запросом (если не пропущена)
	if !skipSync {
		if err := c.SyncService.Sync(); err != nil {
			fmt.Println("Failed to sync local secrets, working offline")
			logger.Log.Warn("Failed to sync local secrets, working offline", zap.Error(err))
		}
	}

	// Сжимаем данные, если необходимо
	if c.UseGzip && len(data) > 0 {
		data, err = compressData(data)
		if err != nil {
			logger.Log.Error("Failed to compress data", zap.Error(err))
			return nil, err
		}
	}

	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, method, c.ServerAddress+url, bytes.NewBuffer(data))
	if err != nil {
		logger.Log.Error("Failed to create request", zap.Error(err))
		return nil, err
	}

	// Добавляем заголовки
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.UseGzip {
		req.Header.Set("Content-Encoding", "gzip")
	}

	// Отправляем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("Failed to send request", zap.Error(err))
		return nil, err
	}

	// Проверяем статус ответа
	if resp.StatusCode == http.StatusUnauthorized {
		logger.Log.Warn("Token is invalid or expired. Need to relogin")
		if err := c.AuthProvider.DeleteToken(); err != nil {
			logger.Log.Error("Failed to delete token", zap.Error(err))
		}
		return nil, fmt.Errorf("token is invalid or expired. Please log in again")
	}

	return resp, nil
}

// SendAddRequest отправляет запрос на добавление секрета
func (c *Client) SendAddRequest(secret models.Secret, skipSync bool) error {
	requestBody, err := json.Marshal(secret)
	if err != nil {
		return err
	}

	resp, err := c.SendRequest("POST", "/api/secrets", requestBody, skipSync)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add secret: status %d", resp.StatusCode)
	}
	return nil
}

// SendDeleteRequest отправляет запрос на удаление секрета
func (c *Client) SendDeleteRequest(name string, skipSync bool) error {
	resp, err := c.SendRequest("DELETE", "/api/secrets/"+name, nil, skipSync)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Если статус 404, секрета уже нет на сервере - это не ошибка
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	// Если статус не 200 OK, возвращаем ошибку
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete secret: status %d", resp.StatusCode)
	}
	return nil
}

// SendUpdateRequest отправляет запрос на обновление секрета
func (c *Client) SendUpdateRequest(secret models.Secret, skipSync bool) error {
	requestBody, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	resp, err := c.SendRequest("PUT", "/api/secrets/"+secret.Name, requestBody, skipSync)
	if err != nil {
		return fmt.Errorf("failed to send update request: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode == http.StatusOK {
		return nil // Успешное обновление
	} else if resp.StatusCode == http.StatusNotFound {
		// Если секрет не найден, пытаемся добавить его
		return c.SendAddRequest(secret, skipSync)
	}

	// Возвращаем ошибку для других статусов
	return fmt.Errorf("failed to update secret: status %d", resp.StatusCode)
}

// SyncLocalSecrets синхронизирует локальные секреты с сервером
func (c *Client) SyncLocalSecrets() error {
	// Загружаем локальные секреты
	secrets, err := local.LoadSecrets()
	if err != nil {
		logger.Log.Error("Failed to load local secrets", zap.Error(err))
		return err
	}

	// Синхронизируем каждый секрет
	for name, secret := range secrets {
		if secret.Synced {
			continue // Пропускаем уже синхронизированные секреты
		}

		if secret.Deleted {
			// Удаление секрета
			if err := c.SendDeleteRequest(name, true); err != nil {
				logger.Log.Warn("Failed to delete secret", zap.String("name", name), zap.Error(err))
				continue
			}
		} else {
			// Добавление/обновление секрета
			if err := c.SendUpdateRequest(secret, true); err != nil {
				logger.Log.Warn("Failed to update secret", zap.String("name", name), zap.Error(err))
				continue
			}
		}

		// Помечаем секрет как синхронизированный
		secret.Synced = true
		secrets[name] = secret
	}

	// Сохраняем обновленные секреты
	if err := local.SaveSecrets(secrets); err != nil {
		logger.Log.Error("Failed to save local secrets", zap.Error(err))
		return err
	}

	logger.Log.Info("Sync completed successfully")
	return nil
}

// compressData сжимает данные с использованием gzip
func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
