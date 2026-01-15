package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/MattiaPun/SubTUI/internal/api"
	"github.com/MattiaPun/SubTUI/internal/integration"
	"github.com/MattiaPun/SubTUI/internal/player"
	"github.com/MattiaPun/SubTUI/internal/ui"
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

	beeep.AppName = "SubTUI"

	if *debug {
		f, err := tea.LogToFile("subtui.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}

		log.Printf("=== SubTUI Started ===")
		log.Printf("Version: %s | Commit: %s", version, commit)
		log.Printf("Config Loaded: %v", api.AppConfig.URL)

		defer func() { _ = f.Close() }()
	} else {
		log.SetOutput(io.Discard)
	}

	if *showVersion {
		fmt.Printf("Version: %s | Commit: %s\n", version, commit)
		os.Exit(0)
	}

	_ = api.LoadConfig()

	// Quiet MPV when TUI is killed
	defer player.ShutdownPlayer()

	p := tea.NewProgram(ui.InitialModel(), tea.WithAltScreen())

	instance := integration.Init(p)
	if instance != nil {
		defer instance.Close()
		go p.Send(ui.SetDBusMsg{Instance: instance})
	}

	discordIns := integration.InitDiscord()
	if discordIns != nil {
		go p.Send(ui.SetDiscordMsg{Instance: discordIns})
	}

	if _, err := p.Run(); err != nil {
		fmt.Println("Error while running program:", err)
		os.Exit(1)
	}
}
