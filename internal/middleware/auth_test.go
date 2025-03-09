package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// MockHandler — это фиктивный обработчик для тестирования middleware.
func MockHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

func TestAuthMiddleware(t *testing.T) {
	// Настройка Gin в тестовом режиме
	gin.SetMode(gin.TestMode)

	// Ключ для подписи JWT
	jwtKey := []byte("test-secret-key")

	tests := []struct {
		name           string
		authHeader     string
		setup          func() // Дополнительная настройка, если требуется
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No Authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Authorization header not found"}`,
		},
		{
			name:           "Invalid Authorization header format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid or expired token"}`,
		},
		{
			name:           "Expired token",
			authHeader:     generateTokenWithExpiration(-1, jwtKey), // Токен просрочен
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid or expired token"}`,
		},
		{
			name:           "Valid token",
			authHeader:     generateTokenWithExpiration(1, jwtKey), // Токен действителен
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый маршрутизатор Gin
			router := gin.Default()

			// Применяем middleware и добавляем фиктивный обработчик
			router.Use(AuthMiddleware(jwtKey))
			router.GET("/test", MockHandler)

			// Создаем HTTP-запрос
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authHeader)
			}
			resp := httptest.NewRecorder()

			// Выполняем запрос
			router.ServeHTTP(resp, req)

			// Проверяем статус ответа
			assert.Equal(t, tt.expectedStatus, resp.Code)

			// Проверяем тело ответа
			assert.JSONEq(t, tt.expectedBody, resp.Body.String())
		})
	}
}

// generateToken создает JWT-токен с указанным методом подписи.
func generateToken(signingMethod jwt.SigningMethod, key []byte) string {
	token := jwt.NewWithClaims(signingMethod, jwt.MapClaims{
		"username": "test-user",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})

	tokenString, _ := token.SignedString(key)
	return tokenString
}

// generateTokenWithExpiration создает JWT-токен с заданным сроком действия (в часах).
func generateTokenWithExpiration(expirationHours int, key []byte) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "test-user",
		"exp":      time.Now().Add(time.Duration(expirationHours) * time.Hour).Unix(),
	})

	tokenString, _ := token.SignedString(key)
	return tokenString
}
