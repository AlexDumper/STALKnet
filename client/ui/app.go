package ui

import (
    "strings"
    "time"

    tea "github.com/charmbracelet/bubbletea"
)

type App struct {
    username       string
    header         *Header
    messages       *MessagesArea
    input          *InputField
    commandHandler *CommandHandler
    quitting       bool
    width          int
    height         int
}

func NewApp(username string) *tea.Program {
    app := &App{
        username: username,
        header:   NewHeader(username),
        messages: NewMessagesArea(),
        input:    NewInputField(),
        quitting: false,
    }

    app.commandHandler = NewCommandHandler(app)
    app.messages.AddSystemMessage("Welcome to STALKnet!")
    app.messages.AddSystemMessage("Type /help for available commands")
    app.messages.AddSystemMessage("Type /mockmsg to see a mock message")
    app.messages.AddSystemMessage("Type /mocktask to see a mock task notification")

    return tea.NewProgram(app, tea.WithAltScreen())
}

func (a *App) Init() tea.Cmd {
    return tea.Tick(1000*time.Millisecond, func(t time.Time) tea.Msg {
        return ConnectedMsg{}
    })
}

type ConnectedMsg struct{}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return a.handleKeyPress(msg)
    case tea.WindowSizeMsg:
        a.width = msg.Width
        a.height = msg.Height
        a.header.SetWidth(msg.Width)
        a.messages.SetWidth(msg.Width)
        a.messages.SetHeight(msg.Height - 5)
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
    a.input.Update(msg)
    return a, nil
}

func (a *App) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    if msg.String() == "ctrl+c" {
        a.messages.AddSystemMessage("Goodbye!")
        a.quitting = true
        return a, tea.Quit
    }
    if msg.String() == "ctrl+l" {
        a.messages.Clear()
        return a, nil
    }
    if msg.String() == "enter" {
        input := a.input.Value()
        if input != "" {
            a.processInput(input)
            a.input.Clear()
        }
        return a, nil
    }
    if msg.String() == "pgup" {
        a.messages.ScrollUp()
        return a, nil
    }
    if msg.String() == "pgdown" {
        a.messages.ScrollDown()
        return a, nil
    }
    a.input.Update(msg)
    return a, nil
}

func (a *App) processInput(input string) {
    if strings.HasPrefix(input, "/") {
        a.commandHandler.Handle(input)
    } else {
        a.messages.AddChatMessage(a.username, input)
    }
}

func (a *App) View() string {
    if a.quitting {
        return "\n  Goodbye!\n\n"
    }
    var b strings.Builder
    b.WriteString(a.header.Render())
    b.WriteString("\n")
    b.WriteString(DividerStyle.Render(strings.Repeat("-", a.width)))
    b.WriteString("\n")
    b.WriteString(a.messages.Render())
    b.WriteString("\n")
    b.WriteString(DividerStyle.Render(strings.Repeat("-", a.width)))
    b.WriteString("\n")
    b.WriteString(a.input.Render())
    b.WriteString("\n")
    b.WriteString("  Controls")
    return AppStyle.Render(b.String())
}

type MockMessageMsg struct {
    Username string
    Content  string
}
