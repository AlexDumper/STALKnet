package ui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
)

type Header struct {
    networkName  string
    username     string
    connected    bool
    serverStatus string
    width        int
}

func NewHeader(username string) *Header {
    return &Header{
        networkName:  "STALKnet",
        username:     username,
        connected:    false,
        serverStatus: "Connecting...",
        width:        80,
    }
}

func (h *Header) SetConnected(connected bool) {
    h.connected = connected
    if connected {
        h.serverStatus = "Connected"
    } else {
        h.serverStatus = "Disconnected"
    }
}

func (h *Header) SetServerStatus(status string) {
    h.serverStatus = status
}

func (h *Header) SetWidth(width int) {
    h.width = width
}

func (h *Header) Render() string {
    indicator := "[*]"
    statusStyle := StatusDisconnectedStyle
    if h.connected {
        statusStyle = StatusConnectedStyle
    }
    left := NetworkNameStyle.Render(fmt.Sprintf(" %s ", h.networkName))
    center := UsernameStyle.Render(fmt.Sprintf(" user: %s ", h.username))
    right := statusStyle.Render(fmt.Sprintf(" %s %s ", indicator, h.serverStatus))
    parts := []string{left, center, right}
    result := strings.Join(parts, " ")
    if h.width > 0 {
        padding := h.width - lipgloss.Width(result)
        if padding > 0 {
            result = lipgloss.NewStyle().Background(ColorGreen).Width(h.width).Render(result)
        }
    }
    return HeaderStyle.Render(result)
}
