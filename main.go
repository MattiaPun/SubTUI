package main

import (
	"fmt"
	"os"

	"git.punjwani.pm/Mattia/DepthTUI/internal/api"
	"git.punjwani.pm/Mattia/DepthTUI/internal/player"
	"git.punjwani.pm/Mattia/DepthTUI/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	_ = api.LoadConfig()

	if api.AppConfig.Password != "" {
		if err := player.InitPlayer(); err != nil {
			fmt.Printf("Failed to start player: %v\n", err)
		}

	}

	defer player.ShutdownPlayer()

	p := tea.NewProgram(ui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
