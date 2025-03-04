package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/logger"
)

type rateLimiter struct {
	request  int
	lastTime time.Time
}

var (
	limit    = 10
	interval = time.Minute
	clients  = make(map[string]*rateLimiter)
	mutex    = &sync.Mutex{}
)

// RateLimitMiddleware обрабатывает запросы на частоту
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем IP-адрес
		ip := c.ClientIP()

		mutex.Lock()
		defer mutex.Unlock()

		client, ok := clients[ip]
		if !ok {
			// Если клиент новый, добавляем его в карту
			clients[ip] = &rateLimiter{
				request:  1,
				lastTime: time.Now(),
			}
		} else {
			// Если клиент уже существует
			if time.Since(client.lastTime) > interval {
				// Если прошло больше интервала, сбрасываем счетчик
				client.request = 1
				client.lastTime = time.Now()
			} else {
				// Увеличиваем счетчик
				client.request++
			}
		}

		// Проверяем лимит
		if clients[ip].request > limit {
			logger.Log.Warn("Too many requests", zap.String("ip", ip))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
