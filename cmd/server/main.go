// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/auth"
	"github.com/FollowLille/goph-vault/internal/config"
	"github.com/FollowLille/goph-vault/internal/database"
	"github.com/FollowLille/goph-vault/internal/handlers"
	"github.com/FollowLille/goph-vault/internal/logger"
	"github.com/FollowLille/goph-vault/internal/middleware"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadServerConfig("config.json")
	if err != nil {
		logger.Log.Fatal("Failed to load server config", zap.Error(err))
	}

	// Инициализация логгера
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		logger.Log.Fatal("Failed to initialize logger", zap.Error(err))
	}

	// Настройка базы данных
	if err := database.SetupDatabase(
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
	); err != nil {
		logger.Log.Fatal("Failed to setup database", zap.Error(err))
	}

	// Инициализация сервисов
	authService := auth.NewAuthService(cfg.EncryptionKey)
	authHandlers := handlers.NewAuthHandlers(*authService, database.GetDB(), []byte(cfg.JWTKey))
	secretHandlers := handlers.NewSecretHandlers(database.GetDB())

	// Инициализация роутера
	r := gin.Default()

	// Middleware
	r.Use(middleware.RequestLogger())
	r.Use(middleware.RateLimitMiddleware())

	// Публичные роуты
	public := r.Group("/api")
	{
		public.POST("/register", authHandlers.RegisterHandler)
		public.POST("/login", authHandlers.LoginHandler)
	}

	// Защищенные роуты
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware([]byte(cfg.JWTKey)))
	{
		protected.GET("/secrets", secretHandlers.GetSecretsHandler)
		protected.POST("/secrets", secretHandlers.AddSecretHandler)
		protected.DELETE("/secrets/:name", secretHandlers.DeleteSecretHandler)
		protected.PUT("/secrets/:name", secretHandlers.UpdateSecretHandler)
		protected.POST("/sync", handlers.SyncSecretsHandler)
	}

	// Создаем контекст для отмены
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск HTTP-сервера
	go func() {
		logger.Log.Info("Starting HTTP server", zap.String("address", cfg.Address))
		if err := r.Run(cfg.Address); err != nil {
			fmt.Println(err)
			logger.Log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Запуск HTTPS-сервера (если указаны сертификаты)
	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		go func() {
			logger.Log.Info("Starting HTTPS server", zap.String("address", cfg.AddressHTTPS))
			if err := r.RunTLS(cfg.AddressHTTPS, cfg.TLSCertFile, cfg.TLSKeyFile); err != nil {
				fmt.Println(err)
				logger.Log.Fatal("Failed to start HTTPS server", zap.Error(err))
			}
		}()
	}

	// Ожидание сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	logger.Log.Info("Server is running. Press Ctrl+C to stop.")

	<-sigChan // Блокируем выполнение до получения сигнала
	logger.Log.Info("Shutting down server...")
	cancel() // Отменяем контекст для завершения горутин
}
