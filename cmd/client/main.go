package main

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/auth"
	"github.com/FollowLille/goph-vault/internal/client"
	"github.com/FollowLille/goph-vault/internal/commands"
	"github.com/FollowLille/goph-vault/internal/config"
	"github.com/FollowLille/goph-vault/internal/logger"
)

func main() {
	// Загрузка конфига
	cfg, err := config.LoadClientConfig()
	if err != nil {
		logger.Log.Fatal("Failed to load config", zap.Error(err))
		fmt.Println(err)
	}
	// Инициализация логгера
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		logger.Log.Fatal("Failed to initialize logger", zap.Error(err))
	}
	// Инициализация сервисов
	authService := auth.NewAuthService(cfg.EncryptionKey)
	httpClient := client.NewClient(cfg.ServerAddress, cfg.AllowHTTP, cfg.AllowInsecureTLS, cfg.UseGzip, cfg.EncryptionKey, authService)

	// Создание корневой команды
	cli := commands.NewRootCmd(httpClient)

	// Выполнение команды
	if err := cli.Execute(); err != nil {
		logger.Log.Fatal("CLI error", zap.Error(err))
	}
}
