package service

import (
	"fmt"

	"github.com/FollowLille/goph-vault/internal/client"
	"github.com/FollowLille/goph-vault/internal/config"
)

// UpdateConfig обновляет параметр конфигурации
func UpdateConfig(client *client.Client, key, value string) error {
	switch key {
	case "server-address":
		client.ServerAddress = value
		fmt.Printf("Updated server address to: %s\n", value)
	case "allow-http":
		client.AllowHTTP = value == "true"
		fmt.Printf("Updated allow-http to: %v\n", client.AllowHTTP)
	case "allow-insecure-tls":
		client.AllowInsecureTLS = value == "true"
		fmt.Printf("Updated allow-insecure-tls to: %v\n", client.AllowInsecureTLS)
	case "use-gzip":
		client.UseGzip = value == "true"
		fmt.Printf("Updated use-gzip to: %v\n", client.UseGzip)
	case "encryption-key":
		client.EncryptionKey = value
		fmt.Printf("Updated encryption-key to: %s\n", value)
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	// Сохраняем обновленную конфигурацию в файл
	cfg := &config.ClientConfig{
		ServerAddress:    client.ServerAddress,
		AllowHTTP:        client.AllowHTTP,
		AllowInsecureTLS: client.AllowInsecureTLS,
		UseGzip:          client.UseGzip,
		EncryptionKey:    client.EncryptionKey,
	}

	if err := config.SaveClientConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}
