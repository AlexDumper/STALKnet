package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputField поле ввода
type InputField struct {
	textInput textinput.Model
	width     int
}

// NewInputField создаёт новое поле ввода
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

// SetValue устанавливает значение
func (i *InputField) SetValue(value string) {
	i.textInput.SetValue(value)
}

// Value возвращает значение
func (i *InputField) Value() string {
	return i.textInput.Value()
}

// Clear очищает поле
func (i *InputField) Clear() {
	i.textInput.SetValue("")
}

// SetWidth устанавливает ширину
func (i *InputField) SetWidth(width int) {
	i.width = width
	i.textInput.Width = width - 4
}

// Focus фокусирует поле
func (i *InputField) Focus() tea.Cmd {
	return i.textInput.Focus()
}

// Blur убирает фокус
func (i *InputField) Blur() {
	i.textInput.Blur()
}

// Focused возвращает статус фокуса
func (i *InputField) Focused() bool {
	return i.textInput.Focused()
}

// Update обновляет состояние
func (i *InputField) Update(msg tea.Msg) tea.Cmd {
	return i.textInput.Update(msg)
}

// Render отрисовывает поле ввода
func (i *InputField) Render() string {
	// Префикс ">"
	prefix := InputStyle.Render("> ")

	// Поле ввода
	input := i.textInput.View()

	// Собираем вместе
	result := prefix + input

	// Добавляем заполнение до нужной ширины
	if i.width > 0 {
		padding := i.width - lipgloss.Width(result)
		if padding > 0 {
			result = result + strings.Repeat(" ", padding)
		}
	}

	return InputStyle.Render(result)
}

// UpdateMsg сообщение обновления
type UpdateMsg struct{}
