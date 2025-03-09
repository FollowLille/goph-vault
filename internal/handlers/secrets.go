package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/database"
	"github.com/FollowLille/goph-vault/internal/logger"
	"github.com/FollowLille/goph-vault/internal/models"
)

// SecretType определяет тип секрета
type SecretType string

const (
	TypeText SecretType = "text" // Значение по умолчанию для Type
)

// SecretHandlers обрабатывает запросы секретов
type SecretHandlers struct {
	db *database.DB
}

// NewSecretHandlers создает новый обработчик
func NewSecretHandlers(db *database.DB) *SecretHandlers {
	return &SecretHandlers{db: db}
}

// GetSecretsHandler возвращает список секретов пользователя.
// @Summary Получить список секретов
// @Description Возвращает все секреты, принадлежащие текущему пользователю
// @Tags Secrets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Secret "Список секретов"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /secrets [get]
func (h *SecretHandlers) GetSecretsHandler(c *gin.Context) {
	var secrets []models.Secret
	// Поиск секретов в базе данных
	if err := h.db.Where("user_id = ?", c.GetUint("userID")).Find(&secrets).Error; err != nil {
		logger.Log.Error("Failed to get secrets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get secrets"})
		return
	}

	c.JSON(http.StatusOK, secrets)
}

// AddSecretHandler добавляет новый секрет.
// @Summary Добавить секрет
// @Description Добавляет новый секрет для текущего пользователя
// @Tags Secrets
// @Accept json
// @Produce json
// @Security BearerAuth
//
//	@Param input body struct{
//	    Name string `json:"name" validate:"required"`
//	    Type models.SecretType `json:"type"`
//	    Metadata string `json:"metadata"`
//	    Data string `json:"data"`
//	} true "Данные нового секрета"
//
// @Success 200 {object} map[string]string "Секрет успешно добавлен"
// @Failure 400 {object} map[string]string "Некорректные входные данные"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /secrets [post]
func (h *SecretHandlers) AddSecretHandler(c *gin.Context) {
	var input struct {
		Name     string            `json:"name" validate:"required"`
		Type     models.SecretType `json:"type"`
		Metadata string            `json:"metadata"`
		Data     string            `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error("Invalid input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	secret := models.Secret{
		UserID:   c.GetUint("userID"),
		Name:     input.Name,
		Type:     input.Type,
		Metadata: input.Metadata,
		Data:     []byte(input.Data),
	}

	// Добавление секрета
	if err := h.db.Create(&secret).Error; err != nil {
		logger.Log.Error("Failed to create secret", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create secret"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secret added successfully"})
}

// UpdateSecretHandler обновляет существующий секрет.
// @Summary Обновить секрет
// @Description Обновляет данные существующего секрета по имени
// @Tags Secrets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Имя секрета"
//
//	@Param input body struct{
//	    Type models.SecretType `json:"type"`
//	    Metadata string `json:"metadata"`
//	    Data string `json:"data"`
//	} true "Обновленные данные секрета"
//
// @Success 200 {object} map[string]string "Секрет успешно обновлен"
// @Failure 400 {object} map[string]string "Некорректные входные данные"
// @Failure 404 {object} map[string]string "Секрет не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /secrets/{name} [put]
func (h *SecretHandlers) UpdateSecretHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		logger.Log.Error("Invalid name")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid name"})
		return
	}

	var input struct {
		Type     models.SecretType `json:"type"`
		Metadata string            `json:"metadata"`
		Data     string            `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error("Invalid input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Поиск секрета
	var secret models.Secret
	if err := h.db.Where("name = ? AND user_id = ?", name, c.GetUint("userID")).First(&secret).Error; err != nil {
		logger.Log.Error("Failed to find secret", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
		return
	}

	secret.Type = input.Type
	secret.Metadata = input.Metadata
	secret.Data = []byte(input.Data)

	// Обновление секрета
	if err := h.db.Save(&secret).Error; err != nil {
		logger.Log.Error("Failed to update secret", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update secret"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secret updated successfully"})
}

// DeleteSecretHandler удаляет секрет.
// @Summary Удалить секрет
// @Description Удаляет секрет по имени
// @Tags Secrets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Имя секрета"
// @Success 200 {object} map[string]string "Секрет успешно удален"
// @Failure 404 {object} map[string]string "Секрет не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /secrets/{name} [delete]
func (h *SecretHandlers) DeleteSecretHandler(c *gin.Context) {
	name := c.Param("name")

	// Проверка имени секрета
	var secret models.Secret
	if err := h.db.Where("name = ? AND user_id = ?", name, c.GetUint("userID")).First(&secret).Error; err != nil {
		logger.Log.Error("Failed to find secret", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
		return
	}

	// Удаление секрета
	if err := h.db.Delete(&secret).Error; err != nil {
		logger.Log.Error("Failed to delete secret", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete secret"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secret deleted successfully"})
}
