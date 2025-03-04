// internal/models/tui/secrets.go

package tui

import (
	"github.com/FollowLille/goph-vault/internal/client"
	"sort"
	"time"

	"github.com/FollowLille/goph-vault/internal/models"
	"github.com/FollowLille/goph-vault/internal/service"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SecretsModel struct {
	secretsList       map[string]models.Secret // Список секретов
	newSecretName     textinput.Model
	newSecretType     textinput.Model
	newSecretMetadata textinput.Model
	newSecretData     textinput.Model
	message           string
	mode              string // Режим: "list", "add", "edit", "view"
	client            *client.Client
	selected          int           // Индекс выбранного секрета (для списка)
	activeField       int           // Активное текстовое поле (для добавления/редактирования)
	viewedSecret      models.Secret // Секрет, который просматривается
}

func newSecretsModel(client *client.Client) SecretsModel {
	m := SecretsModel{
		client:      client,
		secretsList: make(map[string]models.Secret),
		mode:        "list", // Начинаем с режима просмотра списка
		selected:    0,
		activeField: 0,
	}
	m.newSecretName = textinput.New()
	m.newSecretName.Placeholder = "Enter secret name"

	m.newSecretType = textinput.New()
	m.newSecretType.Placeholder = "Enter secret type"

	m.newSecretMetadata = textinput.New()
	m.newSecretMetadata.Placeholder = "Enter metadata"

	m.newSecretData = textinput.New()
	m.newSecretData.Placeholder = "Enter secret data"

	// Получаем список секретов при инициализации
	secrets, err := service.GetSecrets(m.client)
	if err != nil {
		m.message = "Error loading secrets: " + err.Error()
	} else {
		m.secretsList = secrets
	}

	return m
}

func (m SecretsModel) Init() tea.Cmd {
	return nil
}

func (m SecretsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+z": // Возврат на предыдущий экран
			if m.mode == "add" || m.mode == "edit" || m.mode == "view" {
				m.mode = "list" // Возвращаемся к списку секретов
				m.newSecretName.Reset()
				m.newSecretType.Reset()
				m.newSecretMetadata.Reset()
				m.newSecretData.Reset()
				return m, nil
			}
			return m, ChangeScreen("dashboard")

		case "up", "down": // Перемещение по списку секретов
			if m.mode == "list" {
				keys := getSortedKeys(m.secretsList)
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

		case "left", "right": // Переключение между полями в режиме добавления/редактирования
			if m.mode == "add" || m.mode == "edit" {
				if msg.String() == "right" {
					m.activeField++
				} else if msg.String() == "left" {
					m.activeField--
				}
				m.activeField = (m.activeField + 4) % 4 // Ограничиваем диапазон [0, 3]

				// Управление фокусом
				m.newSecretName.Blur()
				m.newSecretType.Blur()
				m.newSecretMetadata.Blur()
				m.newSecretData.Blur()

				switch m.activeField {
				case 0:
					m.newSecretName.Focus()
				case 1:
					m.newSecretType.Focus()
				case 2:
					m.newSecretMetadata.Focus()
				case 3:
					m.newSecretData.Focus()
				}
			}

		case "ctrl+a": // Добавить новый секрет
			if m.mode == "list" {
				m.mode = "add"
				m.activeField = 0
				m.newSecretName.Focus()
			}

		case "ctrl+d": // Удалить выбранный секрет
			if m.mode == "list" {
				keys := getSortedKeys(m.secretsList)
				if len(keys) == 0 {
					m.message = "No secrets to delete!"
					break
				}
				name := keys[m.selected]
				err := service.DeleteSecret(m.client, name)
				if err != nil {
					m.message = "Error: " + err.Error()
				} else {
					m.message = "Secret deleted successfully!"
					delete(m.secretsList, name)
					m.selected = 0 // Сбрасываем выбор
				}
			}

		case "ctrl+e": // Редактировать выбранный секрет
			if m.mode == "list" {
				keys := getSortedKeys(m.secretsList)
				if len(keys) == 0 {
					m.message = "No secrets to edit!"
					break
				}
				name := keys[m.selected]
				secret := m.secretsList[name]

				// Заполняем текстовые поля значениями секрета
				m.newSecretName.SetValue(secret.Name)
				m.newSecretType.SetValue(string(secret.Type))
				m.newSecretMetadata.SetValue(secret.Metadata)
				m.newSecretData.SetValue(string(secret.Data))

				// Переходим в режим редактирования
				m.mode = "edit"
				m.activeField = 0
				m.newSecretName.Focus()
			}

		case "ctrl+l": // Просмотреть выбранный секрет
			if m.mode == "list" {
				keys := getSortedKeys(m.secretsList)
				if len(keys) == 0 {
					m.message = "No secrets to view!"
					break
				}
				name := keys[m.selected]
				m.viewedSecret = m.secretsList[name]
				m.mode = "view"
			}

		case "ctrl+s": // Сохранение изменений
			if m.mode == "add" || m.mode == "edit" {
				// Проверка обязательных полей
				if m.newSecretName.Value() == "" || m.newSecretData.Value() == "" {
					m.message = "Error: Name and Data fields are required!"
					break
				}

				if m.mode == "add" {
					// Вызов сервиса для добавления нового секрета
					err := service.AddSecret(
						m.client,
						m.newSecretName.Value(),
						m.newSecretType.Value(),
						m.newSecretMetadata.Value(),
						m.newSecretData.Value(),
					)
					if err != nil {
						m.message = "Error: " + err.Error()
					} else {
						m.message = "Secret added successfully!"
					}
				} else if m.mode == "edit" {
					// Вызов сервиса для обновления секрета
					keys := getSortedKeys(m.secretsList)
					if len(keys) == 0 {
						m.message = "No secrets to update!"
						break
					}
					name := keys[m.selected]
					err := service.UpdateSecret(
						m.client,
						name,
						m.newSecretType.Value(),
						m.newSecretMetadata.Value(),
						m.newSecretData.Value(),
					)
					if err != nil {
						m.message = "Error: " + err.Error()
					} else {
						m.message = "Secret updated successfully!"
					}
				}

				// Обновляем список секретов
				secrets, err := service.GetSecrets(m.client)
				if err != nil {
					m.message = "Error updating secrets list: " + err.Error()
				} else {
					m.secretsList = secrets
				}
				m.newSecretName.Reset()
				m.newSecretType.Reset()
				m.newSecretMetadata.Reset()
				m.newSecretData.Reset()
				m.mode = "list"
				return m, tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
					return nil // Остаемся на том же экране
				})
			}

		default:
			// Обновление текстовых полей при вводе текста
			if m.mode == "add" || m.mode == "edit" {
				switch m.activeField {
				case 0:
					m.newSecretName, cmd = m.newSecretName.Update(msg)
				case 1:
					m.newSecretType, cmd = m.newSecretType.Update(msg)
				case 2:
					m.newSecretMetadata, cmd = m.newSecretMetadata.Update(msg)
				case 3:
					m.newSecretData, cmd = m.newSecretData.Update(msg)
				}
			}
		}

	default:
		// Обновление текстовых полей для всех типов сообщений
		if m.mode == "add" || m.mode == "edit" {
			switch m.activeField {
			case 0:
				m.newSecretName, cmd = m.newSecretName.Update(msg)
			case 1:
				m.newSecretType, cmd = m.newSecretType.Update(msg)
			case 2:
				m.newSecretMetadata, cmd = m.newSecretMetadata.Update(msg)
			case 3:
				m.newSecretData, cmd = m.newSecretData.Update(msg)
			}
		}
	}

	return m, cmd
}

func (m SecretsModel) View() string {
	title := "Secrets Management"

	switch m.mode {
	case "list":
		// Отображение списка секретов
		secretsView := ""
		keys := getSortedKeys(m.secretsList)
		for i, name := range keys {
			prefix := "  "
			if i == m.selected {
				prefix = "> "
			}
			secretsView += prefix + name + "\n"
		}

		help := "Use ↑/↓ to navigate, Ctrl+A to add, Ctrl+E to edit\nCtrl+D to delete, Ctrl+L to view, Ctrl+Z to return"

		return title + "\n\n" +
			"Secrets:\n" + secretsView + "\n" +
			help

	case "add", "edit":
		// Отображение полей для добавления/редактирования секрета
		formTitle := "Add New Secret"
		if m.mode == "edit" {
			formTitle = "Edit Secret"
		}

		fields := formatField("Name", m.newSecretName.View(), m.activeField == 0) + "\n" +
			formatField("Type", m.newSecretType.View(), m.activeField == 1) + "\n" +
			formatField("Metadata", m.newSecretMetadata.View(), m.activeField == 2) + "\n" +
			formatField("Data", m.newSecretData.View(), m.activeField == 3)

		message := ""
		if m.message != "" {
			message = "Message: " + m.message
		}

		help := "Use ←/→ to switch fields, Ctrl+S to save, Ctrl+Z to cancel"

		return title + "\n\n" +
			formTitle + ":\n" + fields + "\n" +
			message + "\n\n" +
			help

	case "view":
		// Отображение просматриваемого секрета
		secret := m.viewedSecret
		fields := formatField("Name", secret.Name, false) + "\n" +
			formatField("Type", string(secret.Type), false) + "\n" +
			formatField("Metadata", secret.Metadata, false) + "\n" +
			formatField("Data", string(secret.Data), false)

		help := "Press Ctrl+Z to return"

		return title + "\n\n" +
			"View Secret:\n" + fields + "\n" +
			help
	}

	return ""
}

func getSortedKeys(m map[string]models.Secret) []string {
	keys := make([]string, 0, len(m))
	for k, secret := range m {
		if secret.Name != "" { // Исключаем секреты с пустым именем
			keys = append(keys, k)
		}
	}
	sort.Strings(keys) // Сортировка ключей
	return keys
}
