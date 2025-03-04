package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GzipMiddleware добавляет возможность работаать с архивированными данными
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем умеют ли применик работать со сжатием, если да - сжимаем
		if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Header("Content-Encoding", "gzip")
			gz := gzip.NewWriter(c.Writer)
			defer gz.Close()
			c.Writer = &gzipResponseWriter{w: gz, ResponseWriter: c.Writer}
		}

		// Проверяем данные на наличие сжатия
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			defer gz.Close()
			c.Request.Body = gz
		}

		c.Next()
	}
}

// gzipResponseWriter определяет тип для исползования сжатия
type gzipResponseWriter struct {
	gin.ResponseWriter
	w *gzip.Writer
}

// Write функцию записи
func (g gzipResponseWriter) Write(b []byte) (int, error) {
	return g.w.Write(b)
}
