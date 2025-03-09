// internal/models/tui/dashboard.go

package tui

import (
	"github.com/FollowLille/goph-vault/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

type DashboardModel struct {
	cursor int // Позиция курсора
	client *client.Client
}

func newDashboardModel(client *client.Client) DashboardModel {
	return DashboardModel{
		cursor: 0,
		client: client,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < 3 {
				m.cursor++
			}
		case "enter":
			// Переход на другой экран через команду
			switch m.cursor {
			case 0:
				return m, ChangeScreen("login")
			case 1:
				return m, ChangeScreen("register")
			case 2:
				return m, ChangeScreen("secrets")
			case 3:
				return m, ChangeScreen("settings")
			}
		}
	}
	return m, nil
}

func (m DashboardModel) View() string {
	menu := "Main Menu:\n\n"
	options := []string{"Login", "Register", "Manage Secrets", "Settings"}

	for i, option := range options {
		if i == m.cursor {
			menu += "> " + option + "\n"
		} else {
			menu += "  " + option + "\n"
		}
	}

	help := "\nUse ↑ ↓ to navigate, Enter to select"
	return menu + help
}
