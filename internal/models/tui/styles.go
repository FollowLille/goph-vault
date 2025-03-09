package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Основные стили
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")). // Белый цвет текста
			Background(lipgloss.Color("#4A90E2")). // Синий фон
			Padding(0, 1)                          // Отступы

	// Стили для текста
	textStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333")) // Темно-серый цвет текста

	// Стили для полей ввода
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A90E2")). // Синяя рамка
			Padding(0, 1)                                // Отступы

	// Стили для активного поля ввода
	focusedInputStyle = inputStyle.Copy().
				BorderForeground(lipgloss.Color("#FFD700")) // Золотая рамка для активного поля

	// Стили для сообщений
	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")). // Красный цвет для ошибок
			Bold(true)                             // Жирный шрифт

	// Стили для списка
	listStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4A90E2")). // Синяя рамка
			Padding(0, 1)                                // Отступы

	// Стили для выбранного элемента списка
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")). // Золотой цвет для выбранного элемента
				Bold(true)                             // Жирный шрифт
)
