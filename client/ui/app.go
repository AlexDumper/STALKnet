package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// App основное приложение
type App struct {
	username   string
	header     *Header
	messages   *MessagesArea
	input      *InputField
	commandHandler *CommandHandler
	quitting   bool
	width      int
	height     int
}

// NewApp создаёт новое приложение
func NewApp(username string) *tea.Program {
	app := &App{
		username: username,
		header:   NewHeader(username),
		messages: NewMessagesArea(),
		input:    NewInputField(),
		quitting: false,
	}

	app.commandHandler = NewCommandHandler(app)

	// Приветственное сообщение
	app.messages.AddSystemMessage("Welcome to STALKnet!")
	app.messages.AddSystemMessage("Type /help for available commands")
	app.messages.AddSystemMessage("Type /mockmsg to see a mock message")
	app.messages.AddSystemMessage("Type /mocktask to see a mock task notification")

	return tea.NewProgram(app, tea.WithAltScreen())
}

// Init инициализация
func (a *App) Init() tea.Cmd {
	// Симуляция подключения через 1 секунду
	return tea.Tick(1000*time.Millisecond, func(t time.Time) tea.Msg {
		return ConnectedMsg{}
	})
}

// ConnectedMsg сообщение о подключении
type ConnectedMsg struct{}

// Update обработка сообщений
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return a.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.header.SetWidth(msg.Width)
		a.messages.SetWidth(msg.Width)
		a.messages.SetHeight(msg.Height - 5) // Вычитаем header, input и отступы
		a.input.SetWidth(msg.Width)
		return a, nil

	case ConnectedMsg:
		a.header.SetConnected(true)
		a.messages.AddSystemMessage("Connected to STALKnet server")
		return a, nil

	case MockMessageMsg:
		a.messages.AddChatMessage(msg.Username, msg.Content)
		return a, nil
	}

	// Обновление поля ввода
	var cmd tea.Cmd
	a.input.Update(msg)

	return a, cmd
}

// handleKeyPress обработка нажатий клавиш
func (a *App) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Ctrl+C - выход
	if msg.String() == "ctrl+c" {
		a.messages.AddSystemMessage("Goodbye!")
		a.quitting = true
		return a, tea.Quit
	}

	// Ctrl+L - очистка экрана
	if msg.String() == "ctrl+l" {
		a.messages.Clear()
		return a, nil
	}

	// Enter - отправка
	if msg.String() == "enter" {
		input := a.input.Value()
		if input != "" {
			a.processInput(input)
			a.input.Clear()
		}
		return a, nil
	}

	// Page Up - прокрутка вверх
	if msg.String() == "pgup" {
		a.messages.ScrollUp()
		return a, nil
	}

	// Page Down - прокрутка вниз
	if msg.String() == "pgdown" {
		a.messages.ScrollDown()
		return a, nil
	}

	// Стрелки вверх/вниз для истории (будущая функциональность)
	if msg.String() == "up" {
		// TODO: история команд
		return a, nil
	}

	if msg.String() == "down" {
		// TODO: история команд
		return a, nil
	}

	// Остальные клавиши - ввод текста
	a.input.Update(msg)

	return a, nil
}

// processInput обрабатывает ввод
func (a *App) processInput(input string) {
	// Если команда
	if strings.HasPrefix(input, "/") {
		a.commandHandler.Handle(input)
	} else {
		// Обычное сообщение
		a.messages.AddChatMessage(a.username, input)
	}
}

// View отрисовка
func (a *App) View() string {
	if a.quitting {
		return "\n  Goodbye!\n\n"
	}

	var b strings.Builder

	// Header
	b.WriteString(a.header.Render())
	b.WriteString("\n")

	// Разделитель
	b.WriteString(DividerStyle.Render(strings.Repeat("─", a.width)))
	b.WriteString("\n")

	// Область сообщений
	b.WriteString(a.messages.Render())
	b.WriteString("\n")

	// Разделитель перед input
	b.WriteString(DividerStyle.Render(strings.Repeat("─", a.width)))
	b.WriteString("\n")

	// Поле ввода
	b.WriteString(a.input.Render())
	b.WriteString("\n")

	// Подсказка
	hint := HintStyle.Render("  Ctrl+C: quit | Ctrl+L: clear | PgUp/PgDn: scroll | /help: commands")
	b.WriteString(hint)

	return AppStyle.Render(b.String())
}

// MockMessageMsg сообщение для мок-данных
type MockMessageMsg struct {
	Username string
	Content  string
}
