package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/MattiaPun/SubTUI/v2/internal/integration"
	"github.com/MattiaPun/SubTUI/v2/internal/player"
	"github.com/MattiaPun/SubTUI/v2/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/beeep"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	// Debug flag
	debug := flag.Bool("debug", false, "Enable debug logging to subtui.log")
	showVersion := flag.Bool("v", false, "Print version and exit")
	flag.Parse()

	// Check for version
	if *showVersion {
		fmt.Printf("Version: %s | Commit: %s\n", version, commit)
		os.Exit(0)
	}

	// Check for debug mode
	if *debug {
		f, err := tea.LogToFile("subtui.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()

		log.Printf("=== SubTUI Started ===")
		log.Printf("Version: %s | Commit: %s", version, commit)
	} else {
		log.SetOutput(io.Discard)
	}

	// Load Config
	if err := api.LoadConfig(); err != nil {
		fmt.Printf("Fatal error loading config: %v\n", err)
		os.Exit(1)
	}

	// Log Startup
	if *debug {
		log.Printf("Config Loaded: URL=%s User=%s", api.AppServerConfig.Server.URL, api.AppServerConfig.Server.Username)
	}

	// Init variables
	ui.InitStyles()
	beeep.AppName = "SubTUI"

	// Quiet MPV when TUI is killed
	defer player.ShutdownPlayer()

	// Init TUI
	p := tea.NewProgram(ui.InitialModel(), tea.WithAltScreen())

	// Start background services
	instance := integration.Init(p)
	if instance != nil {
		defer instance.Close()
		go p.Send(ui.SetDBusMsg{Instance: instance})
	}

	discordIns := integration.InitDiscord()
	if discordIns != nil {
		go p.Send(ui.SetDiscordMsg{Instance: discordIns})
	}

	// Start TUI
	if _, err := p.Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
