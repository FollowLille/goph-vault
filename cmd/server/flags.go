// Путь: goph-vault/cmd/server/flags.go

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/FollowLille/goph-vault/internal/config"
	"github.com/FollowLille/goph-vault/internal/logger"
)

var ConfigFilePath string

// ParseFlags парсит флаги, переменные окружения и файл конфигурации
func ParseFlags() (*config.ServerConfig, error) {
	var cfg config.ServerConfig

	// Определение флагов
	pflag.StringVarP(&cfg.JWTKey, "jwt-key", "k", "", "JWT secret key")
	pflag.StringVarP(&cfg.Address, "address", "a", "localhost:8080", "Server address")
	pflag.StringVarP(&cfg.AddressHTTPS, "address-https", "A", "localhost:8443", "HTTPS server address")
	pflag.StringVarP(&cfg.LogLevel, "log-level", "l", "info", "Log level")
	pflag.StringVarP(&cfg.Database.Host, "db-host", "H", "localhost", "Database host")
	pflag.IntVarP(&cfg.Database.Port, "db-port", "P", 5432, "Database port")
	pflag.StringVarP(&cfg.Database.User, "db-user", "U", "postgres", "Database user")
	pflag.StringVarP(&cfg.Database.Password, "db-password", "W", "", "Database password")
	pflag.StringVarP(&cfg.Database.DBName, "db-name", "N", "gophvault", "Database name")
	pflag.StringVarP(&ConfigFilePath, "config", "c", "", "Path to config file")
	pflag.StringVarP(&cfg.TLSCertFile, "tls-cert-file", "t", "", "Path to TLS certificate file")
	pflag.StringVarP(&cfg.TLSKeyFile, "tls-key-file", "T", "", "Path to TLS key file")
	pflag.Parse()

	// Переопределение из переменных окружения
	if envJWTKey := os.Getenv("JWT_KEY"); envJWTKey != "" {
		cfg.JWTKey = envJWTKey
	}
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		cfg.Address = envAddress
	}
	if envAddressHttps := os.Getenv("ADDRESS_HTTPS"); envAddressHttps != "" {
		cfg.AddressHTTPS = envAddressHttps
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if envDBHost := os.Getenv("DB_HOST"); envDBHost != "" {
		cfg.Database.Host = envDBHost
	}
	if envDBPort := os.Getenv("DB_PORT"); envDBPort != "" {
		port, err := strconv.Atoi(envDBPort)
		if err != nil {
			logger.Log.Error("Invalid database port", zap.Error(err))
			return nil, err
		}
		cfg.Database.Port = port
	}
	if envDBUser := os.Getenv("DB_USER"); envDBUser != "" {
		cfg.Database.User = envDBUser
	}
	if envDBPassword := os.Getenv("DB_PASSWORD"); envDBPassword != "" {
		cfg.Database.Password = envDBPassword
	}
	if envDBName := os.Getenv("DB_NAME"); envDBName != "" {
		cfg.Database.DBName = envDBName
	}

	if envTLSCertFile := os.Getenv("TLS_CERT_FILE"); envTLSCertFile != "" {
		cfg.TLSCertFile = envTLSCertFile
	}
	if envTLSKeyFile := os.Getenv("TLS_KEY_FILE"); envTLSKeyFile != "" {
		cfg.TLSKeyFile = envTLSKeyFile
	}

	// Переопределение из файла конфигурации
	if ConfigFilePath != "" {
		file, err := os.Open(ConfigFilePath)
		if err != nil {
			logger.Log.Error("Failed to open config file", zap.Error(err))
			return nil, err
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&cfg); err != nil {
			logger.Log.Error("Failed to decode config file", zap.Error(err))
			return nil, err
		}
	}

	// Логирование конфигурации
	logger.Log.Info("Configuration loaded",
		zap.String("jwt_key", cfg.JWTKey),
		zap.String("address", cfg.Address),
		zap.String("address_https", cfg.AddressHTTPS),
		zap.String("log_level", cfg.LogLevel),
		zap.String("db_host", cfg.Database.Host),
		zap.Int("db_port", cfg.Database.Port),
		zap.String("db_user", cfg.Database.User),
		zap.String("db_name", cfg.Database.DBName),
	)
	fmt.Println("Flags: ", cfg)

	return &cfg, nil
}
