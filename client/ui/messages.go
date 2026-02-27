package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// MessageType тип сообщения
type MessageType int

const (
	MessageNormal MessageType = iota
	MessageSystem
	MessageTask
	MessageError
	MessageSuccess
)

// Message сообщение в чате
type Message struct {
	Timestamp time.Time
	Username  string
	Content   string
	Type      MessageType
}

// MessagesArea область сообщений
type MessagesArea struct {
	messages []Message
	scroll   int
	height   int
	width    int
}

// NewMessagesArea создаёт новую область сообщений
func NewMessagesArea() *MessagesArea {
	return &MessagesArea{
		messages: make([]Message, 0),
		scroll:   0,
		height:   20,
		width:    80,
	}
}

// AddMessage добавляет сообщение
func (m *MessagesArea) AddMessage(msg Message) {
	m.messages = append(m.messages, msg)
	// Прокрутка вниз при новом сообщении
	if m.scroll > 0 {
		m.scroll++
	}
}

// AddSystemMessage добавляет системное сообщение
func (m *MessagesArea) AddSystemMessage(content string) {
	m.AddMessage(Message{
		Timestamp: time.Now(),
		Username:  "system",
		Content:   content,
		Type:      MessageSystem,
	})
}

// AddTaskMessage добавляет сообщение о задаче
func (m *MessagesArea) AddTaskMessage(content string) {
	m.AddMessage(Message{
		Timestamp: time.Now(),
		Username:  "task",
		Content:   content,
		Type:      MessageTask,
	})
}

// AddChatMessage добавляет обычное сообщение чата
func (m *MessagesArea) AddChatMessage(username, content string) {
	m.AddMessage(Message{
		Timestamp: time.Now(),
		Username:  username,
		Content:   content,
		Type:      MessageNormal,
	})
}

// SetHeight устанавливает высоту области
func (m *MessagesArea) SetHeight(height int) {
	m.height = height
}

// SetWidth устанавливает ширину области
func (m *MessagesArea) SetWidth(width int) {
	m.width = width
}

// ScrollUp прокрутка вверх
func (m *MessagesArea) ScrollUp() {
	if m.scroll < len(m.messages)-m.height+1 && m.scroll < len(m.messages) {
		m.scroll++
	}
}

// ScrollDown прокрутка вниз
func (m *MessagesArea) ScrollDown() {
	if m.scroll > 0 {
		m.scroll--
	}
}

// Clear очищает сообщения
func (m *MessagesArea) Clear() {
	m.messages = make([]Message, 0)
	m.scroll = 0
}

// Render отрисовывает область сообщений
func (m *MessagesArea) Render() string {
	if len(m.messages) == 0 {
		return HintStyle.Render("  No messages yet. Type /help for commands.")
	}

	var lines []string

	// Начальный индекс с учётом скролла
	start := len(m.messages) - m.height + m.scroll
	if start < 0 {
		start = 0
	}

	// Конечный индекс
	end := start + m.height
	if end > len(m.messages) {
		end = len(m.messages)
	}

	// Рендерим сообщения
	for i := start; i < end; i++ {
		if i < len(m.messages) {
			lines = append(lines, m.renderMessage(m.messages[i]))
		}
	}

	// Добавляем пустые строки для заполнения высоты
	for len(lines) < m.height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderMessage рендерит одно сообщение
func (m *MessagesArea) renderMessage(msg Message) string {
	timeStr := TimeStyle.Render(msg.Timestamp.Format("15:04:05"))

	var content string
	switch msg.Type {
	case MessageSystem:
		content = SystemMessageStyle.Render(fmt.Sprintf("  [%s] <system> %s", timeStr, msg.Content))
	case MessageTask:
		content = TaskMessageStyle.Render(fmt.Sprintf("  [%s] <task> %s", timeStr, msg.Content))
	case MessageError:
		content = ErrorStyle.Render(fmt.Sprintf("  [%s] <error> %s", timeStr, msg.Content))
	case MessageSuccess:
		content = SuccessStyle.Render(fmt.Sprintf("  [%s] <success> %s", timeStr, msg.Content))
	default:
		username := UsernameInMessageStyle.Render("<" + msg.Username + ">")
		content = MessageStyle.Render(fmt.Sprintf("  [%s] %s %s", timeStr, username, msg.Content))
	}

	return content
}
