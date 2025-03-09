// internal/models/tui/app_model.go

package tui

import (
	"github.com/FollowLille/goph-vault/internal/client"
	tea "github.com/charmbracelet/bubbletea"
)

type AppModel struct {
	client *client.Client       // Клиент для взаимодействия с сервером
	screen string               // Текущий экран
	models map[string]tea.Model // Модели для каждого экрана
}

func NewAppModel(client *client.Client) AppModel {
	return AppModel{
		client: client,
		screen: "dashboard",
		models: map[string]tea.Model{
			"dashboard": newDashboardModel(client),
			"login":     newLoginModel(client),
			"register":  newRegisterModel(client),
			"secrets":   newSecretsModel(client),
			"settings":  newSettingsModel(client),
		},
	}
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit // Завершаем приложение
		}
	case changeScreenMsg: // Обработка команды для переключения экранов
		m.screen = string(msg)
		return m, nil
	}

	currentModel := m.models[m.screen]
	updatedModel, cmd := currentModel.Update(msg)
	m.models[m.screen] = updatedModel
	return m, cmd
}

// Команда для переключения экранов
type changeScreenMsg string

func ChangeScreen(screen string) tea.Cmd {
	return func() tea.Msg {
		return changeScreenMsg(screen)
	}
}

func (m AppModel) View() string {
	return m.models[m.screen].View()
}

func formatField(label, value string, focused bool) string {
	if focused {
		return "> " + label + ": " + value
	}
	return "  " + label + ": " + value
}
