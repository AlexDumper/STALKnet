package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/stalknet/client/auth"
)

type CommandHandler struct {
	app *App
}

func NewCommandHandler(app *App) *CommandHandler {
	return &CommandHandler{app: app}
}

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
		h.handleMockMsg()
	case "/mocktask":
		h.handleMockTask()
	case "/scroll":
		h.handleScroll(args)
	case "/auth":
		h.handleAuth()
	case "/y":
		h.handleConfirm()
	case "/n":
		h.handleDecline()
	case "/cancel":
		h.handleCancel()
	case "/login":
		h.handleLogin(args)
	case "/logout":
		h.handleLogout()
	default:
		h.app.messages.AddSystemMessage(fmt.Sprintf("Unknown command: %s", cmd))
	}
}

func (h *CommandHandler) handleHelp() {
	h.app.messages.AddSystemMessage("╭────────────────────────────────────────────╮")
	h.app.messages.AddSystemMessage("│ Available commands:")
	h.app.messages.AddSystemMessage("│ /help - Show this help")
	h.app.messages.AddSystemMessage("│ /clear - Clear screen")
	h.app.messages.AddSystemMessage("│ /connect - Connection status")
	h.app.messages.AddSystemMessage("│ /quit - Exit")
	h.app.messages.AddSystemMessage("│ /auth - Start registration/login")
	h.app.messages.AddSystemMessage("│ /login <user> <pass> - Direct login")
	h.app.messages.AddSystemMessage("│ /logout - Logout current user")

	if h.app.authState == auth.StateAuthorized {
		h.app.messages.AddSystemMessage("│ /nick <name> - Change username")
		h.app.messages.AddSystemMessage("│ /mock <text> - Send message")
		h.app.messages.AddSystemMessage("│ /mockmsg - Random message")
		h.app.messages.AddSystemMessage("│ /mocktask - Show task")
	}
	h.app.messages.AddSystemMessage("╰────────────────────────────────────────────╯")
}

func (h *CommandHandler) handleClear() {
	h.app.messages.Clear()
	h.app.messages.AddSystemMessage("Screen cleared")
}

func (h *CommandHandler) handleConnect() {
	s := "Disconnected"
	if h.app.header.connected {
		s = "Connected"
	}
	h.app.messages.AddSystemMessage("Status: " + s)
}

func (h *CommandHandler) handleQuit() {
	h.app.messages.AddSystemMessage("Goodbye!")
	h.app.quitting = true
}

func (h *CommandHandler) handleMock(args []string) {
	if h.app.authState != auth.StateAuthorized {
		h.app.messages.AddSystemMessage("Authorization required. Type /auth to authorize.")
		return
	}
	if len(args) == 0 {
		h.app.messages.AddSystemMessage("Usage: /mock <text>")
		return
	}
	h.app.messages.AddChatMessage(h.app.username, strings.Join(args, " "))
}

func (h *CommandHandler) handleMockMsg() {
	if h.app.authState != auth.StateAuthorized {
		h.app.messages.AddSystemMessage("Authorization required. Type /auth to authorize.")
		return
	}
	h.app.messages.AddChatMessage("alice", "Hello!")
}

func (h *CommandHandler) handleMockTask() {
	if h.app.authState != auth.StateAuthorized {
		h.app.messages.AddSystemMessage("Authorization required. Type /auth to authorize.")
		return
	}
	h.app.messages.AddTaskMessage("Task #1: Review code")
}

func (h *CommandHandler) handleScroll(args []string) {
	if len(args) == 0 {
		h.app.messages.AddSystemMessage("Usage: /scroll up/down")
		return
	}
	if args[0] == "up" {
		h.app.messages.ScrollUp()
	} else {
		h.app.messages.ScrollDown()
	}
}

func (h *CommandHandler) handleNick(args []string) {
	if h.app.authState != auth.StateAuthorized {
		h.app.messages.AddSystemMessage("Authorization required. Type /auth to authorize.")
		return
	}
	if len(args) == 0 {
		h.app.messages.AddSystemMessage("Usage: /nick <name>")
		return
	}
	h.app.messages.AddSystemMessage("Username cannot be changed after authorization.")
	h.app.messages.AddSystemMessage("Use /logout and login with different user.")
}

// Команды авторизации
func (h *CommandHandler) handleAuth() {
	if h.app.authState == auth.StateAuthorized {
		h.app.messages.AddSystemMessage("You are already authorized as: " + h.app.username)
		h.app.messages.AddSystemMessage("Session ID: " + h.app.sessionID)
		return
	}

	h.app.authState = auth.StateEnteringName
	h.app.messages.AddSystemMessage("╭────────────────────────────────────────────╮")
	h.app.messages.AddSystemMessage("│ STALKnet Authorization")
	h.app.messages.AddSystemMessage("│ Enter your stalker name:")
	h.app.messages.AddSystemMessage("│ Type /cancel to abort")
	h.app.messages.AddSystemMessage("╰────────────────────────────────────────────╯")
}

func (h *CommandHandler) handleConfirm() {
	if h.app.authState != auth.StateConfirmCreate {
		h.app.messages.AddSystemMessage("Nothing to confirm. Type /auth to start authorization.")
		return
	}

	h.app.authState = auth.StateEnteringPassword
	h.app.messages.AddSystemMessage("╭────────────────────────────────────────────╮")
	h.app.messages.AddSystemMessage("│ Enter password for: " + h.app.pendingUsername)
	h.app.messages.AddSystemMessage("│ Type /cancel to abort")
	h.app.messages.AddSystemMessage("╰────────────────────────────────────────────╯")
}

func (h *CommandHandler) handleDecline() {
	if h.app.authState != auth.StateConfirmCreate {
		h.app.messages.AddSystemMessage("Nothing to decline.")
		return
	}

	h.app.cancelAuth()
}

func (h *CommandHandler) handleCancel() {
	if h.app.authState == auth.StateEnteringName ||
		h.app.authState == auth.StateConfirmCreate ||
		h.app.authState == auth.StateEnteringPassword {
		h.app.cancelAuth()
	} else {
		h.app.messages.AddSystemMessage("Nothing to cancel.")
	}
}

func (h *CommandHandler) handleLogin(args []string) {
	if h.app.authState == auth.StateAuthorized {
		h.app.messages.AddSystemMessage("Already logged in as: " + h.app.username)
		return
	}

	if len(args) < 2 {
		h.app.messages.AddSystemMessage("Usage: /login <username> <password>")
		return
	}

	username := args[0]
	password := strings.Join(args[1:], " ")

	h.app.loading = true
	h.app.messages.AddSystemMessage("Logging in...")

	go func() {
		ctx := context.Background()
		token, err := h.app.authManager.Login(ctx, username, password)
		if err != nil {
			h.app.resultChan <- AuthResultMsg{Success: false, Error: err.Error()}
			return
		}
		h.app.resultChan <- AuthResultMsg{Success: true, Token: token}
	}()
}

func (h *CommandHandler) handleLogout() {
	if h.app.authState != auth.StateAuthorized {
		h.app.messages.AddSystemMessage("Not logged in.")
		return
	}

	ctx := context.Background()
	err := h.app.authManager.Logout(ctx)
	if err != nil {
		h.app.messages.AddSystemMessage("Logout error: " + err.Error())
		return
	}

	h.app.authState = auth.StateGuest
	h.app.username = "guest"
	h.app.sessionID = ""
	h.app.header.SetUsername("guest")
	h.app.header.SetSessionID("")

	h.app.messages.AddSystemMessage("Logged out successfully.")
}
