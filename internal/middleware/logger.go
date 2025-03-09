package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/logger"
)

// RequestLogger отвечат за логирование запросов и ответов на них
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Сохраняем тело запроса
		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // Восстанавливаем тело запроса
		}

		// Логируем запрос
		logger.Log.Info("Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.ByteString("body", body), // Логируем тело запроса
		)

		// Создаем responseLogger с инициализированным буфером
		responseWriter := &responseLogger{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{}, // Инициализируем буфер
		}
		c.Writer = responseWriter

		// Продолжаем выполнение запроса
		c.Next()

		// Логируем ответ
		logger.Log.Info("Response",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

type responseLogger struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseLogger) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
