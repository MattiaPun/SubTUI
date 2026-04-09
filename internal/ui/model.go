package ui

import (
	"image"
	"time"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/MattiaPun/SubTUI/v2/internal/integration"
	"github.com/MattiaPun/SubTUI/v2/internal/player"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/x/mosaic"
)

var albumTypes = []string{"All", "Random", "Favorites", "Recently Added", "Recently Played", "Most Played"}

// --- MODEL ---
type model struct {
	textInput    textinput.Model
	songs        []api.Song
	songsPrev    []api.Song
	albums       []api.Album
	albumsPrev   []api.Album
	artists      []api.Artist
	artistsPrev  []api.Artist
	playlists    []api.Playlist
	playerStatus player.PlayerStatus

	// Navigation State
	focus          int
	cursorMain     int
	cursorMainPrev int
	cursorSide     int
	sideOffset     int
	cursorPopup    int
	mainOffset     int
	mainOffsetPrev int

	// Window Dimensions
	width  int
	height int

	// View Mode
	viewMode        int
	filterMode      int
	displayMode     int
	displayModePrev int

	// Cover Art
	coverArt    image.Image
	coverMosaic mosaic.Mosaic

	// App State
	err                error
	loading            bool
	lastPlayedSongPath string
	scrobbled          bool
	loginErr           string
	discordRPC         bool
	notify             bool

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
	loginType   int

	// Input State
	lastKey string

	// View States
	showMediaPlayer bool
	showHelp        bool
	showLoginExtras bool
	showPlaylists   bool
	showRating      bool
	helpModel       HelpModel

	// Pagination State
	lastSearchQuery string
	albumListType   string
	pageOffset      int
	pageHasMore     bool

	// Mouse state
	lastClickTime time.Time
	lastClickId   string

	// Lyrics sidebar state
	lyricsVisible        bool             // whether the lyrics panel is shown
	lyricsLoading        bool             // true while a fetch is in flight
	lyricsResult         api.LyricsResult // the fetched lyrics
	lyricsSongID         string           // the song ID for which lyrics are loaded
	lyricsPrefetchSongID string           // the song ID currently being prefetched
	lyricsScrollOff      int              // current scroll offset (line index) for the viewport
	lyricsFocused        bool             // true when keyboard focus is inside the lyrics panel
	lyricsError          string           // non-empty if the last fetch failed
	lyricsCurrentLine    int              // index of the currently active line (synced mode)
	lyricsPlainLines     []string         // cached plain lyrics lines to avoid repeated splitting
	lyricsManualScroll   bool             // true if the user has manually scrolled; suppresses auto-scroll
	lyricsScrollStopTime time.Time        // when manual scrolling was initiated; used for the 3-second delay
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

type shuffledSongsMsg struct {
	songs      []api.Song
	updateView bool
}

type starredResultMsg struct {
	result *api.SearchResult3
}

type playQueueResultMsg struct {
	result *api.PlayQueue
}

type viewStarredSongsMsg *api.SearchResult3

type coverArtMsg struct {
	img image.Image
}

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
