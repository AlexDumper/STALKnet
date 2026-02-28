package ui

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/stalknet/client/auth"
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	username        string
	sessionID       string
	authManager     *auth.AuthManager
	authState       auth.AuthState
	pendingUsername string
	pendingPassword string
	header          *Header
	messages        *MessagesArea
	input           *InputField
	commandHandler  *CommandHandler
	quitting        bool
	width           int
	height          int
	loading         bool
	resultChan      chan AuthResultMsg
}

func NewApp(username string) *tea.Program {
	authServiceURL := os.Getenv("STALKNET_AUTH_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8081"
	}

	app := &App{
		username:    username,
		authManager: auth.NewAuthManager(authServiceURL),
		authState:   auth.StateGuest,
		header:      NewHeader(username),
		messages:    NewMessagesArea(),
		input:       NewInputField(),
		quitting:    false,
		loading:     false,
		resultChan:  make(chan AuthResultMsg, 1),
	}

	app.commandHandler = NewCommandHandler(app)
	app.messages.AddSystemMessage("Welcome to STALKnet!")
	app.messages.AddSystemMessage("Type /help for available commands")
	app.messages.AddSystemMessage("Type /auth to start authorization process")
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

type AuthResultMsg struct {
	Success bool
	Error   string
	Token   *auth.TokenResponse
}

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
		a.messages.AddSystemMessage("Connected to STALKnet gateway")
		return a, nil
	case AuthResultMsg:
		a.loading = false
		if msg.Success {
			a.username = msg.Token.Username
			a.sessionID = msg.Token.SessionID
			a.authState = auth.StateAuthorized
			a.header.SetUsername(a.username)
			a.header.SetSessionID(a.sessionID)
			a.messages.AddSystemMessage("Login successful!")
			a.messages.AddSystemMessage("Welcome, " + a.username + "!")
			a.messages.AddSystemMessage("Your session ID: " + a.sessionID)
			a.pendingPassword = ""
		} else {
			a.messages.AddSystemMessage("Auth error: " + msg.Error)
			a.authState = auth.StateGuest
			a.pendingUsername = ""
			a.pendingPassword = ""
		}
		return a, nil
	case MockMessageMsg:
		a.messages.AddChatMessage(msg.Username, msg.Content)
		return a, nil
	}

	// Проверка канала результатов
	select {
	case result := <-a.resultChan:
		return a.Update(result)
	default:
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
			if a.loading {
				a.messages.AddSystemMessage("Please wait, processing...")
				return a, nil
			}
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
	switch a.authState {
	case auth.StateEnteringName:
		a.handleEnteringName(input)
		return
	case auth.StateEnteringPassword:
		a.handleEnteringPassword(input)
		return
	}

	if strings.HasPrefix(input, "/") {
		a.commandHandler.Handle(input)
	} else {
		if a.authState != auth.StateAuthorized {
			a.messages.AddSystemMessage("Authorization required. Type /auth to authorize.")
			return
		}
		a.messages.AddChatMessage(a.username, input)
	}
}

func (a *App) handleEnteringName(input string) {
	username := strings.TrimSpace(input)
	if username == "" {
		a.messages.AddSystemMessage("Username cannot be empty. Try again or /cancel to abort.")
		return
	}

	a.pendingUsername = username
	a.authState = auth.StateConfirmCreate
	a.messages.AddSystemMessage("Username '" + username + "' will be checked.")
	a.messages.AddSystemMessage("Create profile? Type /y to confirm or /n to cancel.")
}

func (a *App) handleEnteringPassword(input string) {
	password := strings.TrimSpace(input)
	if password == "" {
		a.messages.AddSystemMessage("Password cannot be empty. Try again or /cancel to abort.")
		return
	}

	a.loading = true
	a.messages.AddSystemMessage("Creating profile and logging in...")

	// Асинхронный вызов
	go func() {
		ctx := context.Background()
		username := a.pendingUsername

		// Регистрация (если пользователь не существует)
		err := a.authManager.Register(ctx, username, password)
		if err != nil && !strings.Contains(err.Error(), "already taken") {
			a.sendAuthResult(AuthResultMsg{Success: false, Error: err.Error()})
			return
		}

		// Вход
		token, err := a.authManager.Login(ctx, username, password)
		if err != nil {
			a.sendAuthResult(AuthResultMsg{Success: false, Error: err.Error()})
			return
		}

		a.sendAuthResult(AuthResultMsg{Success: true, Token: token})
	}()
}

func (a *App) sendAuthResult(result AuthResultMsg) {
	a.resultChan <- result
}

func (a *App) cancelAuth() {
	a.messages.AddSystemMessage("Authorization cancelled.")
	a.authState = auth.StateGuest
	a.pendingUsername = ""
	a.pendingPassword = ""
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
	if a.loading {
		b.WriteString("  [Processing...]")
	} else {
		b.WriteString("  Controls: Ctrl+C=quit | Ctrl+L=clear | PgUp/PgDn=scroll")
	}
	return AppStyle.Render(b.String())
}

type MockMessageMsg struct {
	Username string
	Content  string
}

// Методы для commands.go
func (a *App) GetAuthState() auth.AuthState       { return a.authState }
func (a *App) GetUsername() string                 { return a.username }
func (a *App) GetSessionID() string                { return a.sessionID }
func (a *App) SetAuthState(state auth.AuthState)   { a.authState = state }
func (a *App) GetPendingUsername() string          { return a.pendingUsername }
func (a *App) SetPendingUsername(name string)      { a.pendingUsername = name }
func (a *App) GetPendingPassword() string          { return a.pendingPassword }
func (a *App) SetPendingPassword(password string)  { a.pendingPassword = password }
func (a *App) GetAuthManager() *auth.AuthManager   { return a.authManager }
func (a *App) IsLoading() bool                     { return a.loading }
func (a *App) SetLoading(loading bool)             { a.loading = loading }
