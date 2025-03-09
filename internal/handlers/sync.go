package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/crypto"
	"github.com/FollowLille/goph-vault/internal/database"
	"github.com/FollowLille/goph-vault/internal/logger"
	"github.com/FollowLille/goph-vault/internal/models"
)

// SyncSecretsHandler синхронизирует секреты пользователя.
// @Summary Синхронизация секретов
// @Description Получает секреты, измененные после указанной версии синхронизации
// @Tags Secrets
// @Accept json
// @Produce json
// @Security BearerAuth
//
//	@Param input body struct{
//	    LastSyncVersion int `json:"last_sync_version"`
//	} true "Последняя версия синхронизации"
//
// @Success 200 {object} map[string]interface{} "Секреты успешно синхронизированы"
// @Failure 400 {object} map[string]string "Некорректные входные данные"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /sync-secrets [post]
func SyncSecretsHandler(c *gin.Context) {
	var input struct {
		LastSyncVersion int `json:"last_sync_version"`
	}

	// Парсинг JSON-тела запроса
	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error("Invalid input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID := c.MustGet("userID").(uint)
	key := c.MustGet("encryptionKey").([]byte)

	// Поиск пользователя в базе данных
	var user models.User
	if err := database.GetDB().Where("id = ?", userID).First(&user).Error; err != nil {
		logger.Log.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Поиск секретов в базе данных
	var secrets []models.Secret
	if err := database.GetDB().Where("user_id = ? AND version > ?", userID, input.LastSyncVersion).Find(&secrets).Error; err != nil {
		logger.Log.Error("Failed to get secrets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get secrets"})
		return
	}

	// Декриптация секретов
	for i := range secrets {
		decrypedData, err := crypto.Decrypt([]byte(secrets[i].Data), key)
		if err != nil {
			logger.Log.Error("Failed to encrypt data", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt data"})
			return
		}
		secrets[i].Data = decrypedData
	}

	user.LastSyncVersion = input.LastSyncVersion
	if err := database.GetDB().Save(&user).Error; err != nil {
		logger.Log.Error("Failed to update sync version", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sync version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secrets synced successfully", "secrets": secrets, "last_sync_version": user.LastSyncVersion})

}
