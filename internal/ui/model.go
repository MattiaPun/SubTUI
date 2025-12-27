package ui

import (
	"git.punjwani.pm/Mattia/DepthTUI/internal/api"
	"git.punjwani.pm/Mattia/DepthTUI/internal/player"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	focusSearch = iota
	focusSidebar
	focusMain
	focusSong
)

const (
	viewList = iota
	viewQueue
)

var (
	// Colors
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	// Global Borders
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle)

	// Focused Border (Brighter)
	activeBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(highlight)
)

// --- MODEL ---

type model struct {
	textInput    textinput.Model
	songs        []api.Song
	playlists    []api.Playlist
	playerStatus player.PlayerStatus

	// Navigation State
	focus      int
	cursorMain int
	cursorSide int
	mainOffset int

	// Window Dimensions
	width  int
	height int

	// View Mode
	viewMode int

	// App State
	err            error
	loading        bool
	playlistAmount int

	// Queue System
	queue      []api.Song
	queueIndex int
}

type songsResultMsg struct {
	songs []api.Song
}

type playlistResultMsg struct {
	playlists []api.Playlist
}

type errMsg struct {
	err error
}

type statusMsg player.PlayerStatus

func InitialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Search songs..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return model{
		textInput:  ti,
		songs:      []api.Song{},
		focus:      focusSearch,
		cursorMain: 0,
		cursorSide: 0,
		viewMode:   viewList,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		getPlaylists(),
		syncPlayerCmd(),
	)
}
