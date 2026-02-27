package ui

import (
	"fmt"
	"strings"
	"time"
)

// CommandHandler обработчик команд
type CommandHandler struct {
	app *App
}

// NewCommandHandler создаёт обработчик команд
func NewCommandHandler(app *App) *CommandHandler {
	return &CommandHandler{app: app}
}

// Handle обрабатывает команду
func (h *CommandHandler) Handle(input string) {
	if !strings.HasPrefix(input, "/") {
		return
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "/help":
		h.handleHelp()
	case "/clear":
		h.handleClear()
	case "/nick":
		h.handleNick(args)
	case "/connect":
		h.handleConnect()
	case "/quit":
		h.handleQuit()
	case "/mock":
		h.handleMock(args)
	case "/mockmsg":
		h.handleMockMsg(args)
	case "/mocktask":
		h.handleMockTask()
	case "/scroll":
		h.handleScroll(args)
	default:
		h.app.messages.AddSystemMessage(fmt.Sprintf("Unknown command: %s. Type /help for list.", cmd))
	}
}

// handleHelp команда помощи
func (h *CommandHandler) handleHelp() {
	helpText := []string{
		"Available commands:",
		"  /help              - Show this help message",
		"  /clear             - Clear messages screen",
		"  /nick <name>       - Change your username",
		"  /connect           - Show connection status",
		"  /quit              - Exit the application",
		"  /mock <text>       - Send a mock message",
		"  /mockmsg           - Generate random mock message",
		"  /mocktask          - Generate mock task notification",
		"  /scroll up/down    - Scroll messages",
	}

	for _, line := range helpText {
		h.app.messages.AddSystemMessage(line)
	}
}

// handleClear очистка экрана
func (h *CommandHandler) handleClear() {
	h.app.messages.Clear()
	h.app.messages.AddSystemMessage("Screen cleared")
}

// handleNick смена имени
func (h *CommandHandler) handleNick(args []string) {
	if len(args) == 0 {
		h.app.messages.AddSystemMessage("Usage: /nick <username>")
		return
	}

	newNick := args[0]
	oldNick := h.app.username
	h.app.username = newNick
	h.app.header.username = newNick

	h.app.messages.AddSystemMessage(fmt.Sprintf("Username changed from '%s' to '%s'", oldNick, newNick))
}

// handleConnect статус подключения
func (h *CommandHandler) handleConnect() {
	status := "Disconnected"
	if h.app.header.connected {
		status = "Connected"
	}

	h.app.messages.AddSystemMessage(fmt.Sprintf("Connection status: %s", status))
	h.app.messages.AddSystemMessage(fmt.Sprintf("Server: %s", h.app.header.serverStatus))
}

// handleQuit выход
func (h *CommandHandler) handleQuit() {
	h.app.messages.AddSystemMessage("Goodbye!")
	h.app.quitting = true
}

// handleMock мок команда
func (h *CommandHandler) handleMock(args []string) {
	if len(args) == 0 {
		h.app.messages.AddSystemMessage("Usage: /mock <message text>")
		return
	}

	text := strings.Join(args, " ")
	h.app.messages.AddChatMessage(h.app.username, text)
}

// handleMockMsg случайное сообщение
func (h *CommandHandler) handleMockMsg(args []string) {
	mockMessages := []struct {
		user    string
		message string
	}{
		{"alice", "Привет всем!"},
		{"bob", "Как дела?"},
		{"charlie", "Работаю над задачей #42"},
		{"diana", "Кто в комнате general?"},
		{"eve", "Нужна помощь с кодом"},
		{"frank", "Задача выполнена, жду подтверждения"},
		{"grace", "Может созвонимся?"},
		{"henry", "Обед через 30 минут"},
	}

	// Выбираем случайное сообщение
	idx := int(time.Now().Unix()) % len(mockMessages)
	mock := mockMessages[idx]

	h.app.messages.AddChatMessage(mock.user, mock.message)
}

// handleMockTask мок задача
func (h *CommandHandler) handleMockTask() {
	tasks := []string{
		"Сделать ревью кода",
		"Обновить документацию",
		"Исправить баг в модуле auth",
		"Добавить тесты для handlers",
		"Оптимизировать запросы к БД",
	}

	idx := int(time.Now().Unix()) % len(tasks)
	task := tasks[idx]

	taskID := int(time.Now().Unix()) % 100
	h.app.messages.AddTaskMessage(fmt.Sprintf("Task #%d created: \"%s\"", taskID, task))
	h.app.messages.AddTaskMessage(fmt.Sprintf("Assigned to: %s", h.app.username))
}

// handleScroll прокрутка
func (h *CommandHandler) handleScroll(args []string) {
	if len(args) == 0 {
		h.app.messages.AddSystemMessage("Usage: /scroll <up|down>")
		return
	}

	direction := strings.ToLower(args[0])
	switch direction {
	case "up":
		h.app.messages.ScrollUp()
	case "down":
		h.app.messages.ScrollDown()
	default:
		h.app.messages.AddSystemMessage("Unknown direction. Use 'up' or 'down'")
	}
}
