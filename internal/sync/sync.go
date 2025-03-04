package sync

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/FollowLille/goph-vault/internal/local"
	"github.com/FollowLille/goph-vault/internal/logger"
	"github.com/FollowLille/goph-vault/internal/models"
	"go.uber.org/zap"
	"net/http"
)

// SyncClient описывает методы, необходимые для синхронизации
type SyncClient interface {
	SendAddRequest(secret models.Secret, skipSync bool) error
	SendDeleteRequest(name string, skipSync bool) error
	SendUpdateRequest(secret models.Secret, skipSync bool) error
	SendRequest(method, url string, data []byte, skipSync bool) (*http.Response, error)
}

// SyncService отвечает за синхронизацию локальных секретов с сервером
type SyncService struct {
	client SyncClient
}

// NewSyncService создает новый экземпляр SyncService
func NewSyncService(client SyncClient) *SyncService {
	return &SyncService{client: client}
}

// Sync выполняет синхронизацию локальных секретов с сервером
func (s *SyncService) Sync() error {
	secrets, err := local.LoadSecrets()
	if err != nil {
		logger.Log.Error("Failed to load local secrets", zap.Error(err))
		return err
	}

	// Загружаем секреты с сервера
	resp, err := s.client.SendRequest("GET", "/api/secrets", nil, true)
	if err != nil {
		logger.Log.Warn("Failed to get secrets from server", zap.Error(err))
	} else {
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var serverSecrets []models.Secret
			if err := json.NewDecoder(resp.Body).Decode(&serverSecrets); err != nil {
				logger.Log.Error("Failed to decode server response", zap.Error(err))
			} else {
				// Обновляем локальные секреты данными с сервера
				for _, serverSecret := range serverSecrets {
					if localSecret, exists := secrets[serverSecret.Name]; exists {
						// Разрешаем конфликт: выбираем более новый секрет
						if serverSecret.UpdatedAt.After(localSecret.LocalUpdatedAt) {
							// Серверный секрет новее
							localSecret.Data = serverSecret.Data
							localSecret.Metadata = serverSecret.Metadata
							localSecret.UpdatedAt = serverSecret.UpdatedAt
						} else if localSecret.UpdatedAt.After(serverSecret.UpdatedAt) {
							// Локальный секрет новее
							localSecret.UpdatedAt = localSecret.LocalUpdatedAt
						}

						serverSecret.Synced = true
						secrets[serverSecret.Name] = localSecret
					} else {
						// Добавляем новый секрет
						serverSecret.Synced = true
						secrets[serverSecret.Name] = serverSecret
					}
				}
			}
		}
	}

	// Обрабатываем локальные секреты
	for name, secret := range secrets {
		if secret.Synced {
			if secret.Deleted {
				delete(secrets, name)
			}
			continue
		}
		// Удаляем секрет с сервера
		if secret.Deleted {
			if err := s.client.SendDeleteRequest(name, true); err != nil {
				logger.Log.Warn("Failed to delete secret", zap.String("name", name), zap.Error(err))
				continue
			}
			delete(secrets, name)
		} else {
			// Если секрет уже существует, обновляем его
			if err := s.client.SendUpdateRequest(secret, true); err != nil {
				logger.Log.Warn("Failed to update secret", zap.String("name", name), zap.Error(err))
				continue

			}
		}

		secret.Synced = true
		secrets[name] = secret
	}

	// Сохраняем изменения
	if err := local.SaveSecrets(secrets); err != nil {
		logger.Log.Error("Failed to save local secrets", zap.Error(err))
		return err
	}

	logger.Log.Info("Sync completed successfully")
	return nil
}

// compressData сжимает данные с использованием gzip
func compressData(data []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}
