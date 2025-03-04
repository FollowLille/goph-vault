package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/FollowLille/goph-vault/internal/logger"
	"github.com/FollowLille/goph-vault/internal/models"
)

type DB struct {
	*gorm.DB
}

var dbInstance *DB

// InitDB инициализирует подключение к базе данных
func InitDB(host, user, password, dbname string, port int) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		host, user, password, dbname, port,
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	dbInstance = &DB{gormDB}
	return dbInstance, nil
}

// SetupDatabase настраивает базу данных (создает, если не существует)
func SetupDatabase(host, user, password, dbname string, port int) error {
	// Подключаемся к postgres без указания базы данных
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable", host, port, user, password)
	adminDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Error("Failed to connect to database:", zap.Error(err))
		return err
	}

	// Проверяем, существует ли база данных
	var count int64
	if err := adminDB.Raw("SELECT count(*) FROM pg_catalog.pg_database WHERE datname = ?", dbname).Count(&count).Error; err != nil {
		logger.Log.Error("Failed to check database existence:", zap.Error(err))
		return err
	}

	// Создаем базу данных, если она не существует
	if count == 0 {
		if err := adminDB.Exec("CREATE DATABASE " + dbname).Error; err != nil {
			logger.Log.Error("Failed to create database:", zap.Error(err))
			return err
		}
		logger.Log.Info("Created database", zap.String("dbname", dbname))
	}

	// Подключаемся к созданной/существующей базе данных
	if _, err := InitDB(host, user, password, dbname, port); err != nil {
		logger.Log.Error("Failed to initialize database:", zap.Error(err))
		return err
	}

	// Выполняем миграции
	if err := dbInstance.AutoMigrate(&models.User{}, &models.Secret{}); err != nil {
		logger.Log.Error("Failed to run migrations:", zap.Error(err))
		return err
	}

	logger.Log.Info("Database setup completed successfully")
	return nil
}

// GetDB возвращает экземпляр базы данных
func GetDB() *DB {
	return dbInstance
}
