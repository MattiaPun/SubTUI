package ui

import (
	"time"

	"github.com/MattiaPun/SubTUI/v2/internal/player"
	tea "github.com/charmbracelet/bubbletea"
)

func syncPlayerCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return statusMsg(player.GetPlayerStatus())
	})
}
