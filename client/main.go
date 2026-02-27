package main

import (
	"fmt"
	"os"

	"github.com/stalknet/client/ui"
)

func main() {
	// Получаем имя пользователя из окружения или используем значение по умолчанию
	username := os.Getenv("STALKNET_USER")
	if username == "" {
		username = "guest"
	}

	// Создаём и запускаем TUI приложение
	p := ui.NewApp(username)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running app: %v\n", err)
		os.Exit(1)
	}
}
