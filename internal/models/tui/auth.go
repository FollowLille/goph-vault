// internal/models/tui/auth.go

package tui

import (
	"github.com/FollowLille/goph-vault/internal/client"
	"github.com/FollowLille/goph-vault/internal/service"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

type LoginModel struct {
	username textinput.Model
	password textinput.Model
	message  string
	active   int
	client   *client.Client
}

type RegisterModel struct {
	username textinput.Model
	password textinput.Model
	message  string
	active   int
	client   *client.Client
}

func newLoginModel(client *client.Client) LoginModel {
	m := LoginModel{client: client}
	m.username = textinput.New()
	m.username.Placeholder = "Enter username"
	m.username.Focus() // Устанавливаем фокус на первое поле по умолчанию

	m.password = textinput.New()
	m.password.Placeholder = "Enter password"
	m.password.EchoMode = textinput.EchoPassword
	m.password.EchoCharacter = '•'

	return m
}

func newRegisterModel(client *client.Client) RegisterModel {
	m := RegisterModel{client: client}
	m.username = textinput.New()
	m.username.Placeholder = "Enter username"
	m.username.Focus() // Устанавливаем фокус на первое поле по умолчанию

	m.password = textinput.New()
	m.password.Placeholder = "Enter password"
	m.password.EchoMode = textinput.EchoPassword
	m.password.EchoCharacter = '•'

	return m
}

func (m LoginModel) Init() tea.Cmd {
	return nil
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+z": // Возврат на главный экран
			return m, ChangeScreen("dashboard")

		case "left", "right": // Переключение между полями
			if msg.String() == "right" {
				m.active++
			} else if msg.String() == "left" {
				m.active--
			}
			m.active = (m.active + 2) % 2 // Ограничиваем диапазон [0, 1]

			// Управление фокусом
			m.username.Blur()
			m.password.Blur()
			switch m.active {
			case 0:
				m.username.Focus()
			case 1:
				m.password.Focus()
			}

		case "ctrl+s": // Отправка данных на сервер
			// Проверка обязательных полей
			if m.username.Value() == "" || m.password.Value() == "" {
				m.message = "Error: Username and password fields are required!"
				break
			}

			// Вызов сервиса для входа
			err := service.Login(m.client, m.username.Value(), m.password.Value())
			if err != nil {
				m.message = "Error: " + err.Error()
			} else {
				m.message = "Login successful!"
				time.Sleep(3 * time.Second) // Задержка перед переходом
				return m, tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
					return changeScreenMsg("dashboard") // Переход на главный экран
				})
			}
		}
	case changeScreenMsg: // Обработка команды для переключения экранов
		return m, ChangeScreen(string(msg))
	}

	// Обновление текстовых полей для всех типов сообщений
	switch m.active {
	case 0:
		m.username, cmd = m.username.Update(msg)
	case 1:
		m.password, cmd = m.password.Update(msg)
	}

	return m, cmd
}

func (m LoginModel) View() string {
	title := "Login"

	usernameField := formatField("Username", m.username.View(), m.active == 0)
	passwordField := formatField("Password", m.password.View(), m.active == 1)

	message := ""
	if m.message != "" {
		message = "Message: " + m.message
	}

	help := "Use ←/→ to switch fields, Ctrl+S to submit, Ctrl+Z to return"

	return title + "\n\n" +
		usernameField + "\n" +
		passwordField + "\n" +
		message + "\n\n" +
		help
}

func (m RegisterModel) Init() tea.Cmd {
	return nil
}

func (m RegisterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+z": // Возврат на главный экран
			return m, ChangeScreen("dashboard")

		case "left", "right": // Переключение между полями
			if msg.String() == "right" {
				m.active++
			} else if msg.String() == "left" {
				m.active--
			}
			m.active = (m.active + 2) % 2 // Ограничиваем диапазон [0, 1]

			// Управление фокусом
			m.username.Blur()
			m.password.Blur()

			switch m.active {
			case 0:
				m.username.Focus()
			case 1:
				m.password.Focus()
			}

		case "ctrl+s": // Отправка данных на сервер
			// Проверка обязательных полей
			if m.username.Value() == "" || m.password.Value() == "" {
				m.message = "Error: Username and password fields are required!"
				break
			}

			// Вызов сервиса для регистрации
			err := service.Register(m.client, m.username.Value(), m.password.Value())
			if err != nil {
				m.message = "Error: " + err.Error()
			} else {
				m.message = "Registration successful!"
				return m, tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
					return changeScreenMsg("dashboard") // Переход на главный экран
				})
			}
		}
	}

	// Обновление текстовых полей для всех типов сообщений
	switch m.active {
	case 0:
		m.username, cmd = m.username.Update(msg)
	case 1:
		m.password, cmd = m.password.Update(msg)
	}

	return m, cmd
}

func (m RegisterModel) View() string {
	title := "Register"

	usernameField := formatField("Username", m.username.View(), m.active == 0)
	passwordField := formatField("Password", m.password.View(), m.active == 1)

	message := ""
	if m.message != "" {
		message = "Message: " + m.message
	}

	help := "Use ←/→ to switch fields, Ctrl+S to submit, Ctrl+Z to return"

	return title + "\n\n" +
		usernameField + "\n" +
		passwordField + "\n\n" +
		message + "\n\n" +
		help
}
