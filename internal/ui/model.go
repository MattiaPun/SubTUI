package ui

import (
	"github.com/MattiaPun/SubTUI/internal/api"
	"github.com/MattiaPun/SubTUI/internal/player"
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
	viewLogin = 99
)

const (
	filterSongs = iota
	filterAlbums
	filterArtist
)

const (
	displaySongs = iota
	displayAlbums
	displayArtist
)

const (
	LoopNone = 0
	LoopAll  = 1
	LoopOne  = 2
)

var albumTypes = []string{"Random", "Favorites", "Recently Added", "Recently Played", "Most Played"}

var (
	// Colors
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#6b6b6bff"}
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

	loginBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Padding(1, 4).
			Align(lipgloss.Center)

	// The "Welcome" header
	loginHeaderStyle = lipgloss.NewStyle().
				Foreground(special).
				Bold(true).
				MarginBottom(1)

	// The footer instruction
	loginHelpStyle = lipgloss.NewStyle().
			Foreground(subtle).
			MarginTop(2)
)

// --- MODEL ---
type model struct {
	textInput    textinput.Model
	songs        []api.Song
	albums       []api.Album
	artists      []api.Artist
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
	viewMode    int
	filterMode  int
	displayMode int

	// App State
	err              error
	loading          bool
	lastPlayedSongID string
	scrobbled        bool

	// Queue System
	queue      []api.Song
	queueIndex int
	loopMode   int

	// Stars
	starredMap map[string]bool

	// Login State
	loginInputs []textinput.Model
	loginFocus  int
}

type songsResultMsg struct {
	songs []api.Song
}

type albumsResultMsg struct {
	albums []api.Album
}

type artistsResultMsg struct {
	artists []api.Artist
}

type playlistResultMsg struct {
	playlists []api.Playlist
}

type starredResultMsg struct {
	result *api.SearchResult3
}

type viewLikedSongsMsg *api.SearchResult3

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

	startMode := viewList
	if api.AppConfig.Username == "" || api.AppConfig.Password == "" || api.AppConfig.URL == "" {
		startMode = viewLogin
	}

	return model{
		textInput:        ti,
		songs:            []api.Song{},
		focus:            focusSearch,
		cursorMain:       0,
		cursorSide:       0,
		viewMode:         startMode,
		filterMode:       filterSongs,
		displayMode:      displaySongs,
		starredMap:       make(map[string]bool),
		lastPlayedSongID: "",
		loginInputs:      initialLoginInputs(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		getPlaylists(),
		syncPlayerCmd(),
		getStarredCmd(),
	)
}

func initialLoginInputs() []textinput.Model {
	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "http(s)://music.example.com"
	inputs[0].Width = 30
	inputs[0].Focus()
	inputs[0].Prompt = "URL:      "

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "username"
	inputs[1].Width = 30
	inputs[1].Prompt = "Username: "

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "password"
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].Width = 30
	inputs[2].Prompt = "Password: "

	return inputs
}
