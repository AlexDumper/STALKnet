package ui

import "github.com/charmbracelet/lipgloss"

// Цветовая схема
var (
	ColorBlack       = lipgloss.Color("#000000")
	ColorDarkGray    = lipgloss.Color("#1a1a1a")
	ColorGray        = lipgloss.Color("#3a3a3a")
	ColorLightGray   = lipgloss.Color("#888888")
	ColorWhite       = lipgloss.Color("#ffffff")
	ColorGreen       = lipgloss.Color("#00ff00")
	ColorRed         = lipgloss.Color("#ff0000")
	ColorPurple      = lipgloss.Color("#7D56F4")
	ColorBlue        = lipgloss.Color("#55aaff")
	ColorYellow      = lipgloss.Color("#ffff55")
	ColorOrange      = lipgloss.Color("#ffaa00")
)

// Стили приложения
var (
	// Основной стиль приложения - чёрный фон
	AppStyle = lipgloss.NewStyle().
			Background(ColorBlack).
			Foreground(ColorWhite)

	// Header - верхняя панель
	HeaderStyle = lipgloss.NewStyle().
			Background(ColorPurple).
			Foreground(ColorWhite).
			Bold(true).
			Padding(0, 1)

	// Название сети
	NetworkNameStyle = lipgloss.NewStyle().
				Background(ColorPurple).
				Foreground(ColorWhite).
				Bold(true)

	// Имя пользователя
	UsernameStyle = lipgloss.NewStyle().
			Background(ColorPurple).
			Foreground(ColorWhite)

	// Статус подключения - подключен
	StatusConnectedStyle = lipgloss.NewStyle().
				Background(ColorPurple).
				Foreground(ColorGreen).
				Bold(true)

	// Статус подключения - отключен
	StatusDisconnectedStyle = lipgloss.NewStyle().
					Background(ColorPurple).
					Foreground(ColorRed).
					Bold(true)

	// Разделитель
	DividerStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	// Сообщение обычное
	MessageStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	// Сообщение системное
	SystemMessageStyle = lipgloss.NewStyle().
				Foreground(ColorBlue)

	// Сообщение задачи
	TaskMessageStyle = lipgloss.NewStyle().
				Foreground(ColorYellow)

	// Сообщение ошибки
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed)

	// Сообщение успеха
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	// Поле ввода
	InputStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	// Подсказка
	HintStyle = lipgloss.NewStyle().
			Foreground(ColorLightGray)

	// Время сообщения
	TimeStyle = lipgloss.NewStyle().
			Foreground(ColorLightGray)

	// Имя пользователя в сообщении
	UsernameInMessageStyle = lipgloss.NewStyle().
				Foreground(ColorPurple).
				Bold(true)

	// Скроллбар
	ScrollbarStyle = lipgloss.NewStyle().
			Foreground(ColorPurple)
)
