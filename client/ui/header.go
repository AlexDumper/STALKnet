package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Header верхняя панель приложения
type Header struct {
	networkName    string
	username       string
	connected      bool
	serverStatus   string
	width          int
}

// NewHeader создаёт новую верхнюю панель
func NewHeader(username string) *Header {
	return &Header{
		networkName:  "STALKnet",
		username:     username,
		connected:    false,
		serverStatus: "Connecting...",
		width:        80,
	}
}

// SetConnected устанавливает статус подключения
func (h *Header) SetConnected(connected bool) {
	h.connected = connected
	if connected {
		h.serverStatus = "Connected"
	} else {
		h.serverStatus = "Disconnected"
	}
}

// SetServerStatus устанавливает статус сервера
func (h *Header) SetServerStatus(status string) {
	h.serverStatus = status
}

// SetWidth устанавливает ширину header
func (h *Header) SetWidth(width int) {
	h.width = width
}

// Render отрисовывает header
func (h *Header) Render() string {
	// Индикатор подключения
	indicator := "●"
	statusStyle := StatusDisconnectedStyle
	if h.connected {
		statusStyle = StatusConnectedStyle
	}

	// Левая часть - название сети
	left := NetworkNameStyle.Render(fmt.Sprintf(" %s ", h.networkName))

	// Центральная часть - имя пользователя
	center := UsernameStyle.Render(fmt.Sprintf(" user: %s ", h.username))

	// Правая часть - статус
	right := statusStyle.Render(fmt.Sprintf(" %s %s ", indicator, h.serverStatus))

	// Собираем всё вместе
	parts := []string{left, center, right}
	
	// Вычисляем общую длину контента
	contentWidth := 0
	for _, p := range parts {
		contentWidth += lipgloss.Width(p)
	}

	// Добавляем отступы между элементами
	result := strings.Join(parts, " ")

	// Если есть место, добавляем заполнение
	if h.width > 0 {
		padding := h.width - lipgloss.Width(result)
		if padding > 0 {
			result = lipgloss.NewStyle().
				Background(ColorPurple).
				Width(h.width).
				Render(result)
		}
	}

	return HeaderStyle.Render(result)
}
