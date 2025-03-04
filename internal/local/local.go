package local

import (
	"encoding/json"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/logger"
	"github.com/FollowLille/goph-vault/internal/models"
)

// secretsFilePath - путь к файлу с локальными секретами
var secretsFilePath = filepath.Join(os.Getenv("HOME"), ".gophvault", "secrets.json")

// SaveSecrets сохраняет локальные секреты в файл
func SaveSecrets(secrets map[string]models.Secret) error {
	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		logger.Log.Error("Failed to marshal secrets", zap.Error(err))
		return err
	}

	if err := os.MkdirAll(filepath.Dir(secretsFilePath), 0700); err != nil {
		logger.Log.Error("Failed to create secrets directory", zap.Error(err))
		return err
	}

	return os.WriteFile(secretsFilePath, data, 0600)
}

// LoadSecrets загружает локальные секреты из файла
func LoadSecrets() (map[string]models.Secret, error) {
	file, err := os.ReadFile(secretsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]models.Secret), nil
		}
		logger.Log.Error("Failed to read secrets file", zap.Error(err))
		return nil, err
	}

	var secrets map[string]models.Secret
	if err := json.Unmarshal(file, &secrets); err != nil {
		logger.Log.Error("Failed to unmarshal secrets", zap.Error(err))
		return nil, err
	}

	return secrets, nil
}
