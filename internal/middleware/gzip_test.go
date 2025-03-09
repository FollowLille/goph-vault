package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGzipMiddleware(t *testing.T) {
	// Настройка Gin в тестовом режиме
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		requestHeaders   map[string]string
		requestBody      string
		expectedStatus   int
		expectedEncoding string
	}{
		{
			name:             "No compression",
			requestHeaders:   map[string]string{},
			expectedStatus:   http.StatusOK,
			expectedEncoding: "",
		},
		{
			name: "Response compressed with gzip",
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip",
			},
			expectedStatus:   http.StatusOK,
			expectedEncoding: "gzip",
		},
		{
			name: "Request body compressed with gzip",
			requestHeaders: map[string]string{
				"Content-Encoding": "gzip",
			},
			requestBody:      compressString(`{"key":"value"}`),
			expectedStatus:   http.StatusOK,
			expectedEncoding: "",
		},
		{
			name: "Invalid gzip request body",
			requestHeaders: map[string]string{
				"Content-Encoding": "gzip",
			},
			requestBody:      "Invalid Gzip Data", // Некорректные gzip-данные
			expectedStatus:   http.StatusBadRequest,
			expectedEncoding: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый маршрутизатор Gin
			router := gin.Default()

			// Применяем middleware и добавляем фиктивный обработчик
			router.Use(GzipMiddleware())
			router.POST("/test", MockHandler)

			// Создаем HTTP-запрос
			var reqBody io.Reader
			if tt.requestBody != "" {
				reqBody = strings.NewReader(tt.requestBody)
			} else {
				reqBody = nil
			}

			req := httptest.NewRequest(http.MethodPost, "/test", reqBody)
			for key, value := range tt.requestHeaders {
				req.Header.Set(key, value)
			}

			resp := httptest.NewRecorder()

			// Выполняем запрос
			router.ServeHTTP(resp, req)

			// Проверяем статус ответа
			assert.Equal(t, tt.expectedStatus, resp.Code)

			// Проверяем заголовок Content-Encoding
			if tt.expectedEncoding == "gzip" {
				assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))
			} else {
				assert.Empty(t, resp.Header().Get("Content-Encoding"))
			}

			// Проверяем тело ответа (если статус успешный)
			if tt.expectedStatus == http.StatusOK {
				if tt.expectedEncoding == "gzip" {
					// Распаковываем gzip-ответ
					gz, err := gzip.NewReader(resp.Body)
					assert.NoError(t, err)
					defer gz.Close()

					body, err := io.ReadAll(gz)
					assert.NoError(t, err)
					assert.JSONEq(t, `{"message":"Success"}`, string(body))
				} else {
					// Проверяем несжатый ответ
					assert.JSONEq(t, `{"message":"Success"}`, resp.Body.String())
				}
			}
		})
	}
}

// compressString сжимает строку с использованием gzip.
func compressString(data string) string {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte(data))
	_ = gz.Close()
	return buf.String()
}
