package ui

import (
    "strings"

    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type InputField struct {
    textInput textinput.Model
    width     int
}

func NewInputField() *InputField {
    ti := textinput.New()
    ti.Placeholder = "Type a message or /command..."
    ti.PromptStyle = lipgloss.NewStyle().Foreground(ColorPurple)
    ti.TextStyle = lipgloss.NewStyle().Foreground(ColorWhite)
    ti.Cursor.Style = lipgloss.NewStyle().Foreground(ColorWhite)
    ti.Cursor.TextStyle = lipgloss.NewStyle().Foreground(ColorWhite)
    ti.CharLimit = 500
    ti.Width = 60
    ti.Focus()
    return &InputField{
        textInput: ti,
        width:     80,
    }
}

func (i *InputField) SetValue(value string) {
    i.textInput.SetValue(value)
}

func (i *InputField) Value() string {
    return i.textInput.Value()
}

func (i *InputField) Clear() {
    i.textInput.SetValue("")
}

func (i *InputField) SetWidth(width int) {
    i.width = width
    i.textInput.Width = width - 4
}

func (i *InputField) Focus() tea.Cmd {
    return i.textInput.Focus()
}

func (i *InputField) Blur() {
    i.textInput.Blur()
}

func (i *InputField) Focused() bool {
    return i.textInput.Focused()
}

func (i *InputField) Update(msg tea.Msg) {
    i.textInput, _ = i.textInput.Update(msg)
}

func (i *InputField) Render() string {
    prefix := InputStyle.Render("> ")
    input := i.textInput.View()
    result := prefix + input
    if i.width > 0 {
        padding := i.width - lipgloss.Width(result)
        if padding > 0 {
            result = result + strings.Repeat(" ", padding)
        }
    }
    return InputStyle.Render(result)
}

type UpdateMsg struct{}
