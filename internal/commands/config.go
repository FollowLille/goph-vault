package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/FollowLille/goph-vault/internal/client"
	"github.com/FollowLille/goph-vault/internal/service"
)

// NewConfigCmd создает команду для управления конфигурацией
func NewConfigCmd(client *client.Client) *cobra.Command {
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Use subcommands to manage configuration.")
		},
	}

	configCmd.AddCommand(NewUpdateConfigCmd(client))

	return configCmd
}

// NewUpdateConfigCmd создает команду для обновления конфигурации
func NewUpdateConfigCmd(client *client.Client) *cobra.Command {
	var updateConfigCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a configuration parameter",
		Run: func(cmd *cobra.Command, args []string) {
			key, _ := cmd.Flags().GetString("key")
			value, _ := cmd.Flags().GetString("value")

			if key == "" || value == "" {
				fmt.Println("Both --key and --value flags are required.")
				return
			}

			// Вызываем сервисный метод для обновления конфигурации
			err := service.UpdateConfig(client, key, value)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("Configuration updated successfully.")
		},
	}

	updateConfigCmd.Flags().StringP("key", "k", "", "Configuration key to update (e.g., server-address)")
	updateConfigCmd.Flags().StringP("value", "v", "", "New value for the configuration key")
	updateConfigCmd.MarkFlagRequired("key")
	updateConfigCmd.MarkFlagRequired("value")

	return updateConfigCmd
}
