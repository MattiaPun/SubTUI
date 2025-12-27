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
	if err := api.InitConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Config Error: %v\n", err)
		os.Exit(1)
	}

	if err := api.SubsonicPing(); err != nil {
		fmt.Fprintf(os.Stderr, "Auth Error: %v\n", err)
		os.Exit(1)
	}

	if err := player.InitPlayer(); err != nil {
		panic(err)
	}
	defer player.ShutdownPlayer()

	p := tea.NewProgram(ui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
