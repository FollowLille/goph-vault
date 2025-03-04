// internal/models/tui/settings.go

package tui

import (
	"fmt"
	"sort"
	"time"

	"github.com/FollowLille/goph-vault/internal/client"
	"github.com/FollowLille/goph-vault/internal/service"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SettingsModel struct {
	configKey   textinput.Model
	configValue textinput.Model
	message     string
	mode        string // Режим: "list", "edit"
	selected    int    // Индекс выбранной настройки
	activeField int    // Активное текстовое поле (для редактирования)
	client      *client.Client
	settings    map[string]string // Список доступных настроек
}

func newSettingsModel(client *client.Client) SettingsModel {
	m := SettingsModel{
		client:   client,
		mode:     "list", // Начинаем с режима просмотра списка
		selected: 0,
	}
	m.configKey = textinput.New()
	m.configKey.Placeholder = "Enter config key"
	m.configValue = textinput.New()
	m.configValue.Placeholder = "Enter config value"

	// Инициализация списка настроек
	m.settings = map[string]string{
		"server-address":     client.ServerAddress,
		"allow-http":         fmt.Sprintf("%v", client.AllowHTTP),
		"allow-insecure-tls": fmt.Sprintf("%v", client.AllowInsecureTLS),
		"use-gzip":           fmt.Sprintf("%v", client.UseGzip),
		"encryption-key":     client.EncryptionKey,
	}

	return m
}

func (m SettingsModel) Init() tea.Cmd {
	return nil
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+z": // Возврат на главный экран
			if m.mode == "edit" {
				m.mode = "list" // Возвращаемся к списку настроек
				m.configKey.Reset()
				m.configValue.Reset()
				return m, nil
			}
			return m, ChangeScreen("dashboard")

		case "up", "down": // Перемещение по списку настроек
			if m.mode == "list" {
				keys := getSortedSettingsKeys(m.settings)
				if len(keys) == 0 {
					break
				}
				if msg.String() == "down" {
					m.selected++
				} else if msg.String() == "up" {
					m.selected--
				}
				if m.selected < 0 {
					m.selected = len(keys) - 1
				} else if m.selected >= len(keys) {
					m.selected = 0
				}
			}

		case "enter": // Переход в режим редактирования
			if m.mode == "list" {
				keys := getSortedSettingsKeys(m.settings)
				if len(keys) == 0 {
					m.message = "No settings to edit!"
					break
				}

				// Получаем выбранный ключ и значение
				selectedKey := keys[m.selected]
				selectedValue := m.settings[selectedKey]

				// Заполняем текстовые поля
				m.configKey.SetValue(selectedKey)
				m.configValue.SetValue(selectedValue)

				// Переходим в режим редактирования
				m.mode = "edit"
				m.activeField = 1 // Фокус сразу на значение
				m.configValue.Focus()
			}

		case "left", "right": // Переключение между полями в режиме редактирования
			if m.mode == "edit" {
				if msg.String() == "right" {
					m.activeField++
				} else if msg.String() == "left" {
					m.activeField--
				}
				m.activeField = (m.activeField + 2) % 2 // Ограничиваем диапазон [0, 1]

				// Управление фокусом
				m.configKey.Blur()
				m.configValue.Blur()

				switch m.activeField {
				case 0:
					m.configKey.Focus()
				case 1:
					m.configValue.Focus()
				}
			}

		case "ctrl+s": // Сохранение изменений
			if m.mode == "edit" {
				keys := getSortedSettingsKeys(m.settings)
				if len(keys) == 0 {
					m.message = "No settings to update!"
					break
				}
				key := keys[m.selected]
				value := m.configValue.Value()

				// Проверка значения в зависимости от ключа
				switch key {
				case "allow-http", "allow-insecure-tls", "use-gzip":
					if value != "true" && value != "false" {
						m.message = "Error: Boolean fields must be 'true' or 'false'"
						return m, nil
					}
				case "server-address", "encryption-key":
					if value == "" {
						m.message = "Error: Field cannot be empty"
						return m, nil
					}
				}

				// Обновляем конфигурацию
				err := service.UpdateConfig(m.client, key, value)
				if err != nil {
					m.message = "Error: " + err.Error()
				} else {
					m.message = "Config updated successfully!"
					m.settings[key] = value // Обновляем локальный список настроек
				}

				// Возвращаемся к списку настроек
				m.mode = "list"
				m.configKey.Reset()
				m.configValue.Reset()
				return m, tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
					return nil // Остаемся на том же экране
				})
			}
		default:
			// Обновление текстовых полей для всех типов сообщений
			if m.mode == "edit" {
				switch m.activeField {
				case 0:
					m.configKey, _ = m.configKey.Update(msg)
				case 1:
					m.configValue, _ = m.configValue.Update(msg)
				}
			}
		}

	default:
		// Обновление текстовых полей для всех типов сообщений
		if m.mode == "edit" {
			switch m.activeField {
			case 0:
				m.configKey, _ = m.configKey.Update(msg)
			case 1:
				m.configValue, _ = m.configValue.Update(msg)
			}
		}
	}

	return m, nil
}

func (m SettingsModel) View() string {
	title := "Settings Management"

	switch m.mode {
	case "list":
		// Отображение списка настроек
		settingsView := ""
		keys := getSortedSettingsKeys(m.settings)
		for i, key := range keys {
			prefix := "  "
			if i == m.selected {
				prefix = "> "
			}
			settingsView += prefix + key + ": " + m.settings[key] + "\n"
		}

		help := "Use ↑/↓ to navigate, Enter to edit, Ctrl+Z to return"

		return title + "\n\n" +
			"Settings:\n" + settingsView + "\n" +
			help

	case "edit":
		// Отображение полей для редактирования
		keyField := formatField("Config Key", m.configKey.View(), false)
		valueField := formatField("Config Value", m.configValue.View(), m.activeField == 1)

		message := ""
		if m.message != "" {
			message = "Message: " + m.message
		}

		help := "Use ←/→ to switch fields, Ctrl+S to save, Ctrl+Z to cancel"

		// Подсказка для пользователя
		hint := ""
		selectedKey := getSortedSettingsKeys(m.settings)[m.selected]
		switch selectedKey {
		case "allow-http", "allow-insecure-tls", "use-gzip":
			hint = "(Allowed values: true, false)"
		case "server-address", "encryption-key":
			hint = "(Cannot be empty)"
		}

		return title + "\n\n" +
			"Edit Setting:\n" + keyField + "\n" + valueField + "\n" +
			hint + "\n\n" +
			message + "\n\n" +
			help
	}

	return ""
}

func getSortedSettingsKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Сортировка ключей
	return keys
}
