package ui

import "fmt"
import "strings"
// time imported but not used

type CommandHandler struct { app *App }

func NewCommandHandler(app *App) *CommandHandler { return &CommandHandler{app: app} }

func (h *CommandHandler) Handle(input string) {
if !strings.HasPrefix(input, "/") { return }
parts := strings.Fields(input)
if len(parts) == 0 { return }
cmd := strings.ToLower(parts[0])
args := parts[1:]
switch cmd {
case "/help": h.handleHelp()
case "/clear": h.handleClear()
case "/nick": h.handleNick(args)
case "/connect": h.handleConnect()
case "/quit": h.handleQuit()
case "/mock": h.handleMock(args)
case "/mockmsg": h.handleMockMsg()
case "/mocktask": h.handleMockTask()
case "/scroll": h.handleScroll(args)
default: h.app.messages.AddSystemMessage(fmt.Sprintf("Unknown: %s", cmd))
}}

func (h *CommandHandler) handleHelp() {
h.app.messages.AddSystemMessage("Commands: /help /clear /nick /connect /quit /mock /mockmsg /mocktask /scroll")
}
func (h *CommandHandler) handleClear() { h.app.messages.Clear(); h.app.messages.AddSystemMessage("Screen cleared") }
  
func (h *CommandHandler) handleConnect() {  
s := "Disconnected"  
if h.app.header.connected { s = "Connected" }  
h.app.messages.AddSystemMessage("Status: " + s)  
}  
  
func (h *CommandHandler) handleQuit() {  
h.app.messages.AddSystemMessage("Goodbye!")  
h.app.quitting = true  
} 
  
func (h *CommandHandler) handleMock(args []string) {  
if len(args) == 0 { h.app.messages.AddSystemMessage("Usage: /mock text"); return }  
h.app.messages.AddChatMessage(h.app.username, strings.Join(args, " "))  
}  
  
func (h *CommandHandler) handleMockMsg() {  
h.app.messages.AddChatMessage("alice", "Hello!")  
}  
  
func (h *CommandHandler) handleMockTask() {  
h.app.messages.AddTaskMessage("Task #1: Review code")  
}  
  
func (h *CommandHandler) handleScroll(args []string) {  
if len(args) == 0 { h.app.messages.AddSystemMessage("Usage: /scroll up/down"); return }  
if args[0] == "up" { h.app.messages.ScrollUp() } else { h.app.messages.ScrollDown() }  
} 
  
func (h *CommandHandler) handleNick(args []string) {  
if len(args) == 0 { h.app.messages.AddSystemMessage("Usage: /nick name"); return }  
h.app.username = args[0]  
h.app.header.username = args[0]  
h.app.messages.AddSystemMessage("Nick changed")  
} 
