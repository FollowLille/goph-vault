package commands

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/FollowLille/goph-vault/internal/client"
	"github.com/FollowLille/goph-vault/internal/models/tui"
)

var (
	version   string
	buildDate string
	useTUI    bool
)

// NewRootCmd возвращает корневую команду
func NewRootCmd(client *client.Client) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "goph-vault",
		Short: "GophVault - secure storage for your secrets",
		Long:  `A CLI tool to securely store and sync your private data.`,
		Run: func(cmd *cobra.Command, args []string) {
			if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
				if version == "" || buildDate == "" {
					fmt.Println("Version information is not available.")
					return
				}
				fmt.Printf("Version: %s\n", version)
				fmt.Printf("Build date: %s\n", buildDate)
				return
			}
			if useTUI {
				fmt.Println("Starting TUI...")
				// Запуск TUI
				appModel := tui.NewAppModel(client)
				p := tea.NewProgram(appModel)
				if _, err := p.Run(); err != nil {
					panic(err) // Обработка ошибок запуска
				}
			} else {
				fmt.Println("Welcome to GophVault!")
			}
		},
	}

	// Добавляем флаги
	rootCmd.PersistentFlags().String("server-address", client.ServerAddress, "Server address")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print version and build date")
	rootCmd.PersistentFlags().BoolVarP(&useTUI, "tui", "z", false, "Run the app in TUI mode")

	// Регистрируем команды
	rootCmd.AddCommand(
		NewRegisterCmd(client),
		NewLoginCmd(client),
		NewConfigCmd(client),
		NewGetSecretsCmd(client),
		NewAddSecretCmd(client),
		NewUpdateSecretCmd(client),
		NewDeleteSecretCmd(client),
		NewGetSecretsByNameCmd(client),
	)

	return rootCmd
}
