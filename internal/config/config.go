package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ServerConfig содержит настройки сервера
type ServerConfig struct {
	JWTKey        string         `json:"jwt_key"`
	EncryptionKey string         `json:"encryption_key"`
	Address       string         `json:"address"`
	Database      DatabaseConfig `json:"database"`
	TLSKeyFile    string         `json:"tls_key_file"`
	TLSCertFile   string         `json:"tls_cert_file"`
	LogLevel      string         `json:"log_level"`
	AddressHTTPS  string         `json:"address_https"`
}

// ClientConfig содержит настройки клиента
type ClientConfig struct {
	ServerAddress      string `json:"server_address"`
	ServerAddressHttps string `json:"server_address_https"`
	AllowHTTP          bool   `json:"allow_http"`
	AllowInsecureTLS   bool   `json:"allow_insecure_tls"`
	UseGzip            bool   `json:"use_gzip"`
	EncryptionKey      string `json:"encryption_key"`
	LogLevel           string `json:"log_level"`
}

// DatabaseConfig содержит настройки базы данных
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

// LoadServerConfig загружает конфигурацию сервера из файла
func LoadServerConfig(path string) (*ServerConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg ServerConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadClientConfig загружает конфигурацию клиента из файла
func LoadClientConfig() (*ClientConfig, error) {
	// Определяем путь к файлу конфигурации
	configPath := filepath.Join(os.Getenv("HOME"), ".gophvault", "client_config.json")

	// Проверяем, существует ли файл
	if _, err := os.Stat(configPath); os.IsNotExist(err) {

		// Если файл не существует, создаем его со стандартными значениями
		defaultConfig := &ClientConfig{
			ServerAddress:      "http://localhost:8080",
			ServerAddressHttps: "https://localhost:8443",
			AllowHTTP:          true,
			AllowInsecureTLS:   false,
			UseGzip:            false,
			EncryptionKey:      "1234567890abcdef1234567890abcdef",
			LogLevel:           "info",
		}
		// Создаем директорию, если она не существует
		if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}

		// Сохраняем стандартную конфигурацию в файл
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal default config: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return nil, fmt.Errorf("failed to write default config file: %w", err)
		}

		fmt.Println("Created default config file at:", configPath)
		return defaultConfig, nil
	}

	// Читаем файл конфигурации
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Парсим JSON
	var cfg ClientConfig
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return &cfg, nil
}

func SaveClientConfig(config *ClientConfig) error {
	// Определяем путь к файлу конфигурации
	configPath := filepath.Join(os.Getenv("HOME"), ".gophvault", "client_config.json")

	// Преобразуем конфигурацию в JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Записываем JSON в файл
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println("Configuration saved successfully.")
	return nil
}
