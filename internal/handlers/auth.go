package handlers

import (
	"errors"
	"gorm.io/gorm"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/FollowLille/goph-vault/internal/auth"
	"github.com/FollowLille/goph-vault/internal/database"
	"github.com/FollowLille/goph-vault/internal/logger"
	"github.com/FollowLille/goph-vault/internal/models"
)

// AuthHandlers обрабатывает запросы аутентификации
type AuthHandlers struct {
	authService *auth.AuthService
	db          *database.DB
	jwtKey      []byte
}

// NewAuthHandlers создает новый обработчик
func NewAuthHandlers(authService auth.AuthService, db *database.DB, jwtKey []byte) *AuthHandlers {
	return &AuthHandlers{
		authService: &authService,
		db:          db,
		jwtKey:      jwtKey,
	}
}

// RegisterHandler регистрирует нового пользователя.
// @Summary Регистрация пользователя
// @Description Регистрирует нового пользователя с указанным именем и паролем
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body struct{Username string; Password string} true "Данные для регистрации"
// @Success 200 {object} map[string]string "Пользователь успешно зарегистрирован"
// @Failure 400 {object} map[string]string "Некорректные входные данные"
// @Failure 409 {object} map[string]string "Имя пользователя уже занято"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /register [post]
func (h *AuthHandlers) RegisterHandler(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error("Invalid input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Проверка существования пользователя
	var existingUser models.User
	err := h.db.Where("username = ?", input.Username).First(&existingUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Пользователь не найден, можно продолжить регистрацию
			logger.Log.Info("User not found, continue registration", zap.Error(err))
		} else {
			// Другая ошибка базы данных
			logger.Log.Error("Database error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	} else {
		// Пользователь уже существует
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Создание пользователя
	user := models.User{
		Username: input.Username,
		Password: string(hashedPassword),
	}
	if err := h.db.Create(&user).Error; err != nil {
		logger.Log.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered"})
}

// LoginHandler выполняет вход пользователя.
// @Summary Вход пользователя
// @Description Аутентифицирует пользователя и возвращает JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body struct{Username string; Password string} true "Данные для входа"
// @Success 200 {object} map[string]string "JWT токен"
// @Failure 400 {object} map[string]string "Некорректные входные данные"
// @Failure 401 {object} map[string]string "Неверные учетные данные"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /login [post]
func (h *AuthHandlers) LoginHandler(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error("Invalid input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var user models.User
	if err := h.db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Генерация JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, _ := token.SignedString(h.jwtKey)

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
