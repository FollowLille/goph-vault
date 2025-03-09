package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/FollowLille/goph-vault/internal/client"
	"github.com/FollowLille/goph-vault/internal/service"
)

// NewGetSecretsCmd создает команду для получения секретов
func NewGetSecretsCmd(client *client.Client) *cobra.Command {
	return &cobra.Command{
		Use:   "get-secrets",
		Short: "Get all secrets",
		Run: func(cmd *cobra.Command, args []string) {
			// Вызываем сервисный метод для получения секретов
			secrets, err := service.GetSecrets(client)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			// Выводим локальные секреты
			fmt.Println("Local secrets:")
			for name, secret := range secrets {
				fmt.Printf("Secret: Name=%s, Type=%s, Synced=%v\n", name, secret.Type, secret.Synced)
			}
		},
	}
}

// NewGetSecretsByNameCmd создает команду для поиска секрета по имени
func NewGetSecretsByNameCmd(client *client.Client) *cobra.Command {
	var getSecretsByNameCmd = &cobra.Command{
		Use:   "get-secrets-by-name",
		Short: "Get secret by name",
		Run: func(cmd *cobra.Command, args []string) {
			secretName, _ := cmd.Flags().GetString("name")

			// Вызываем сервисный метод для поиска секрета
			secret, found, err := service.GetSecretByName(client, secretName)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			if found {
				fmt.Printf("Secret: Name=%s, Type=%s, Synced=%v\n", secretName, secret.Type, secret.Synced)
			} else {
				fmt.Printf("Secret not found: Name=%s\n", secretName)
			}
		},
	}

	getSecretsByNameCmd.Flags().StringP("name", "n", "", "Secret name")
	getSecretsByNameCmd.MarkFlagRequired("name")
	return getSecretsByNameCmd
}

// NewAddSecretCmd создает команду для добавления секрета
func NewAddSecretCmd(client *client.Client) *cobra.Command {
	var addSecretCmd = &cobra.Command{
		Use:   "add-secret",
		Short: "Add a new secret",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			secretType, _ := cmd.Flags().GetString("type")
			metadata, _ := cmd.Flags().GetString("metadata")
			data, _ := cmd.Flags().GetString("data")

			// Вызываем сервисный метод для добавления секрета
			err := service.AddSecret(client, name, secretType, metadata, data)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
		},
	}

	addSecretCmd.Flags().StringP("name", "n", "", "Name of the secret")
	addSecretCmd.Flags().StringP("type", "t", "", "Type of the secret")
	addSecretCmd.Flags().StringP("metadata", "m", "", "Metadata for the secret")
	addSecretCmd.Flags().StringP("data", "d", "", "Data for the secret")
	addSecretCmd.MarkFlagRequired("name")
	addSecretCmd.MarkFlagRequired("type")
	addSecretCmd.MarkFlagRequired("data")

	return addSecretCmd
}

// NewUpdateSecretCmd создает команду для обновления секрета
func NewUpdateSecretCmd(client *client.Client) *cobra.Command {
	var updateSecretCmd = &cobra.Command{
		Use:   "update-secret",
		Short: "Update an existing secret",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			secretType, _ := cmd.Flags().GetString("type")
			metadata, _ := cmd.Flags().GetString("metadata")
			data, _ := cmd.Flags().GetString("data")

			// Вызываем сервисный метод для обновления секрета
			err := service.UpdateSecret(client, name, secretType, metadata, data)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
		},
	}

	// Добавляем флаги
	updateSecretCmd.Flags().StringP("name", "n", "", "Name of the secret")
	updateSecretCmd.Flags().StringP("type", "t", "", "New type for the secret")
	updateSecretCmd.Flags().StringP("metadata", "m", "", "New metadata for the secret")
	updateSecretCmd.Flags().StringP("data", "d", "", "New data for the secret")
	updateSecretCmd.MarkFlagRequired("name")
	updateSecretCmd.MarkFlagRequired("data")

	return updateSecretCmd
}

// NewDeleteSecretCmd создает команду для удаления секрета
func NewDeleteSecretCmd(client *client.Client) *cobra.Command {
	var deleteSecretCmd = &cobra.Command{
		Use:   "delete-secret",
		Short: "Delete an existing secret",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")

			// Вызываем сервисный метод для удаления секрета
			err := service.DeleteSecret(client, name)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
		},
	}

	deleteSecretCmd.Flags().StringP("name", "n", "", "Name of the secret")
	deleteSecretCmd.MarkFlagRequired("name")

	return deleteSecretCmd
}
