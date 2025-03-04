package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/FollowLille/goph-vault/internal/client"
	"github.com/FollowLille/goph-vault/internal/service"
)

// NewRegisterCmd создает команду для регистрации
func NewRegisterCmd(client *client.Client) *cobra.Command {
	var registerCmd = &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		Run: func(cmd *cobra.Command, args []string) {
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")

			// Вызываем сервисный метод для регистрации
			err := service.Register(client, username, password)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("User registered successfully")
		},
	}

	registerCmd.Flags().StringP("username", "u", "", "Username")
	registerCmd.Flags().StringP("password", "p", "", "Password")
	registerCmd.MarkFlagRequired("username")
	registerCmd.MarkFlagRequired("password")

	return registerCmd
}

// NewLoginCmd создает команду для авторизации
func NewLoginCmd(client *client.Client) *cobra.Command {
	var loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Login a user",
		Run: func(cmd *cobra.Command, args []string) {
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")

			// Вызов функции авторизации
			if err := service.Login(client, username, password); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("User logged in successfully")
		},
	}

	loginCmd.Flags().StringP("username", "u", "", "Username")
	loginCmd.Flags().StringP("password", "p", "", "Password")
	loginCmd.MarkFlagRequired("username")
	loginCmd.MarkFlagRequired("password")

	return loginCmd
}
