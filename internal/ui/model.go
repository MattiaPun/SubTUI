package ui

import (
	"github.com/MattiaPun/SubTUI/internal/api"
	"github.com/MattiaPun/SubTUI/internal/integration"
	"github.com/MattiaPun/SubTUI/internal/player"
	"github.com/charmbracelet/bubbles/textinput"
)

var albumTypes = []string{"All", "Random", "Favorites", "Recently Added", "Recently Played", "Most Played"}

// --- MODEL ---
type model struct {
	textInput    textinput.Model
	songs        []api.Song
	albums       []api.Album
	artists      []api.Artist
	playlists    []api.Playlist
	playerStatus player.PlayerStatus

	// Navigation State
	focus       int
	cursorMain  int
	cursorSide  int
	sideOffset  int
	cursorPopup int
	mainOffset  int

	// Window Dimensions
	width  int
	height int

	// View Mode
	viewMode        int
	filterMode      int
	displayMode     int
	displayModePrev int

	// App State
	err              error
	loading          bool
	lastPlayedSongID string
	scrobbled        bool
	loginErr         string
	discordRPC       bool
	notify           bool

	// Integrations
	dbusInstance    *integration.Instance
	discordInstance *integration.DiscordInstance

	// Queue System
	queue      []api.Song
	queueIndex int
	loopMode   int

	// Stars
	starredMap map[string]bool

	// Login State
	loginInputs []textinput.Model
	loginFocus  int

	// Input State
	lastKey string

	// Overlay States
	showHelp      bool
	showPlaylists bool
	showRating    bool
	helpModel     HelpModel

	// Pagination State
	lastSearchQuery string
	albumListType   string
	pageOffset      int
	pageHasMore     bool
}

type HelpModel struct {
	Width  int
	Height int
}

type ContentModel struct {
	Content string
}

type BackgroundWrapper struct {
	RenderedView string
}

type loginResultMsg struct {
	err error
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

type playQueueResultMsg struct {
	result *api.PlayQueue
}

type viewStarredSongsMsg *api.SearchResult3

type createShareMsg struct {
	url string
}

type errMsg struct {
	err error
}

type statusMsg player.PlayerStatus

type SetDBusMsg struct {
	Instance *integration.Instance
}

type SetDiscordMsg struct {
	Instance *integration.DiscordInstance
}
