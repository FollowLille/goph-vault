package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/client"
	"github.com/FollowLille/goph-vault/internal/local"
	"github.com/FollowLille/goph-vault/internal/logger"
	"github.com/FollowLille/goph-vault/internal/models"
)

// GetSecrets выполняет синхронизацию и загружает локальные секреты
func GetSecrets(client *client.Client) (map[string]models.Secret, error) {
	// Выполняем синхронизацию
	if err := client.SyncService.Sync(); err != nil {
		logger.Log.Error("Failed to synchronize secrets", zap.Error(err))
		fmt.Println("Failed to synchronize secrets. Falling back to local secrets...")
	}

	// Загружаем локальные секреты
	secrets, err := local.LoadSecrets()
	if err != nil {
		logger.Log.Error("Failed to load local secrets", zap.Error(err))
		return nil, fmt.Errorf("failed to load local secrets: %w", err)
	}

	return secrets, nil
}

// GetSecretByName находит секрет по имени
func GetSecretByName(client *client.Client, secretName string) (models.Secret, bool, error) {
	// Выполняем синхронизацию
	if err := client.SyncService.Sync(); err != nil {
		fmt.Println("Failed to synchronize secrets. Falling back to local secrets...")
	}

	// Загружаем локальные секреты
	secrets, err := local.LoadSecrets()
	if err != nil {
		return models.Secret{}, false, fmt.Errorf("failed to load local secrets: %w", err)
	}

	// Ищем секрет по имени
	if secret, exists := secrets[secretName]; exists {
		return secret, true, nil
	}

	return models.Secret{}, false, nil
}

// AddSecret добавляет новый секрет
func AddSecret(client *client.Client, name, secretType, metadata, data string) error {
	// Создаем тело запроса для отправки на сервер
	requestBody, err := json.Marshal(map[string]string{
		"name":     name,
		"type":     secretType,
		"metadata": metadata,
		"data":     data,
	})
	if err != nil {
		return fmt.Errorf("failed to prepare request body: %w", err)
	}

	// Отправляем запрос на сервер
	resp, err := client.SendRequest("POST", "/api/secrets", requestBody, false)
	if err != nil {
		logger.Log.Error("Failed to send request to server", zap.Error(err))
	} else {
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			logger.Log.Error("Failed to add secret on server", zap.Int("status", resp.StatusCode))
		} else {
			fmt.Println("Secret added successfully on server")
		}
	}

	// Загружаем локальные секреты
	secrets, err := local.LoadSecrets()
	if err != nil {
		return fmt.Errorf("failed to load local secrets: %w", err)
	}

	// Добавляем новый секрет
	newSecret := models.Secret{
		Name:           name,
		Type:           models.SecretType(secretType),
		Metadata:       metadata,
		Data:           []byte(data),
		Version:        1, // Начальная версия
		LocalUpdatedAt: time.Now(),
		Synced:         false, // Помечаем как несинхронизированный
	}

	secrets[name] = newSecret

	// Сохраняем изменения
	if err := local.SaveSecrets(secrets); err != nil {
		return fmt.Errorf("failed to save local secrets: %w", err)
	}

	fmt.Println("Secret saved locally")
	return nil
}

// UpdateSecret обновляет существующий секрет
func UpdateSecret(client *client.Client, name, secretType, metadata, data string) error {
	// Загружаем локальные секреты
	secrets, err := local.LoadSecrets()
	if err != nil {
		return fmt.Errorf("failed to load local secrets: %w", err)
	}

	// Проверяем, существует ли секрет локально
	secret, exists := secrets[name]
	if !exists {
		logger.Log.Warn("Secret not found locally", zap.String("name", name))
		fmt.Println("Secret not found locally")
	}

	// Подготовка данных для отправки на сервер
	requestBody, err := json.Marshal(map[string]interface{}{
		"type":     secretType,
		"metadata": metadata,
		"data":     data,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	// Попытка обновить секрет на сервере
	resp, err := client.SendRequest("PUT", "/api/secrets/"+name, requestBody, false)
	if err != nil {
		logger.Log.Warn("Failed to send request to server", zap.Error(err))
	} else {
		defer resp.Body.Close()

		// Проверяем статус ответа
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Secret updated successfully on server")
		} else if resp.StatusCode == http.StatusNotFound {
			if !exists {
				fmt.Println("Secret not found locally and not on server, you need to add it manually")
				return nil
			}
			logger.Log.Warn("Secret not found on server, attempting to add it", zap.String("name", name))

			// Обновляем данные секрета локально
			secret.Type = models.SecretType(secretType)
			secret.Metadata = metadata
			secret.Data = []byte(data)
			secret.LocalUpdatedAt = time.Now()
			secret.Synced = false // Секрет больше не синхронизирован
			secrets[name] = secret

			// Сохраняем изменения локально
			if err := local.SaveSecrets(secrets); err != nil {
				return fmt.Errorf("failed to save local secrets: %w", err)
			}

			// Если секрет не найден на сервере, но существует локально, добавляем его
			addResp, addErr := client.SendRequest("POST", "/api/secrets", requestBody, false)
			if addErr != nil {
				logger.Log.Error("Failed to add secret to server", zap.Error(addErr))
				return fmt.Errorf("failed to add secret to server: %w", addErr)
			}
			defer addResp.Body.Close()

			if addResp.StatusCode != http.StatusOK {
				logger.Log.Error("Failed to add secret to server", zap.Int("status", addResp.StatusCode))
				return fmt.Errorf("failed to add secret to server: status %d", addResp.StatusCode)
			}

			fmt.Println("Secret added successfully on server")
		} else {
			logger.Log.Warn("Failed to update secret on server", zap.Int("status", resp.StatusCode))
			return fmt.Errorf("failed to update secret on server: status %d", resp.StatusCode)
		}
	}

	fmt.Println("Secret updated locally")
	return nil
}

// DeleteSecret удаляет секрет
func DeleteSecret(client *client.Client, name string) error {
	// Загружаем локальные секреты
	secrets, err := local.LoadSecrets()
	if err != nil {
		return fmt.Errorf("failed to load local secrets: %w", err)
	}

	// Проверяем, существует ли секрет локально
	secret, exists := secrets[name]
	if !exists {
		logger.Log.Warn("Secret not found locally", zap.String("name", name))
		fmt.Println("Secret not found locally")
	}

	// Попытка удалить секрет с сервера
	resp, err := client.SendRequest("DELETE", "/api/secrets/"+name, nil, true)
	if err != nil {
		logger.Log.Warn("Failed to send request to server, marking secret as deleted locally", zap.Error(err))
	} else {
		defer resp.Body.Close()

		// Проверяем статус ответа
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Secret deleted successfully on server")
			if exists {
				// Удаляем секрет локально
				delete(secrets, name)
			}
			return nil
		} else if resp.StatusCode == http.StatusNotFound {
			if !exists {
				fmt.Println("Secret not found locally and not on server")
				return nil
			}
			logger.Log.Warn("Secret not found on server, marking secret as deleted locally", zap.String("name", name))
		}
	}

	// Помечаем секрет как удалённый локально
	secret.Deleted = true
	secret.Synced = false
	secrets[name] = secret

	// Сохраняем изменения локально
	if err := local.SaveSecrets(secrets); err != nil {
		return fmt.Errorf("failed to save local secrets: %w", err)
	}

	fmt.Println("Secret marked as deleted locally")
	return nil
}
