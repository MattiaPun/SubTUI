package ui

import (
	"math/rand"
	"strings"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/MattiaPun/SubTUI/v2/internal/integration"
	"github.com/MattiaPun/SubTUI/v2/internal/player"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) handlesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if keyMatches(key, api.AppConfig.Keybinds.Global.HardQuit) {
		return hardQuit(m)
	}

	if m.viewMode == viewLogin {
		return login(m, msg)
	}

	if m.showMediaPlayer {
		return playerMenu(m, msg)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Global.CycleFocusNext) {
		return cycleFocus(m, true), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Global.CycleFocusPrev) {
		return cycleFocus(m, false), nil
	}

	if m.focus == focusSearch {
		if keyMatches(key, api.AppConfig.Keybinds.Navigation.Select) {
			return enter(m)
		}

		if keyMatches(key, api.AppConfig.Keybinds.Search.FilterNext) {
			return cycleFilter(m, true), nil
		}

		if keyMatches(key, api.AppConfig.Keybinds.Search.FilterPrev) {
			return cycleFilter(m, false), nil
		}

		return typeInput(m, msg)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Global.Back) {
		return goBack(m)
	}

	if m.showPlaylists {
		return playlistsMenu(key, m)
	}

	if m.showRating {
		return ratingMenu(key, m)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Global.Help) {
		m.showHelp = !m.showHelp
		return m, nil
	} else if m.showHelp {
		return m, nil
	}

	if key == "g" || m.lastKey == "g" {
		switch key {
		case "g":
			if m.lastKey == "g" {
				return navigateTop(m), nil
			} else {
				m.lastKey = "g"
				return m, nil
			}
		case "a":
			return displayAlbumFromSelected(m)
		case "r":
			return displayArtistFromSelected(m)
		default:
			m.lastKey = ""
		}
	}

	// GLOBAL KEYBINDS

	if keyMatches(key, api.AppConfig.Keybinds.Global.Quit) {
		return quit(m, msg)
	}

	// NAVIGATION KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Up) {
		return navigateUp(m, 1), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Down) {
		return navigateDown(m, 1)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Bottom) {
		return navigateBottom(m)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Select) {
		return enter(m)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.ToggleSelection) {
		return toggleSelection(m)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.PlayShuffled) {
		return playShuffled(m)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.GoHalfPageUp) {
		return navigateUp(m, (m.height-17)/2), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.GoHalfPageDown) {
		return navigateDown(m, (m.height-17)/2)
	}

	// SEARCH KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Search.FocusSearch) {
		return focusSearchBar(m), nil
	}

	// LIBRARY KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Library.AddToPlaylist) {
		return toggleAddToPlaylistPopup(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Library.AddRating) {
		return toggleAddRatingPopup(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Library.Rate0) {
		return setRating(m, 0)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Library.Rate1) {
		return setRating(m, 1)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Library.Rate2) {
		return setRating(m, 2)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Library.Rate3) {
		return setRating(m, 3)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Library.Rate4) {
		return setRating(m, 4)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Library.Rate5) {
		return setRating(m, 5)
	}

	// MEDIA KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Media.PlayPause) {
		return mediaTogglePlay(m, msg), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Next) {
		return mediaSongSkip(m, msg)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Prev) {
		return mediaSongPrev(m, msg)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.VolumeUp) {
		return mediaVolumeUp(m, msg)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.VolumeDown) {
		return mediaVolumeDown(m, msg)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Shuffle) {
		return mediaShuffle(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Loop) {
		return mediaToggleLoop(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Restart) {
		return mediaRestartSong(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Rewind) {
		return mediaSeekRewind(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Forward) {
		return mediaSeekForward(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.ToggleMediaPlayer) {
		return mediaToggleMediaPlayer(m), nil
	}

	// QUEUE KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Queue.ToggleQueueView) {
		return toggleQueue(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Queue.QueueNext) {
		return mediaQueueNext(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Queue.QueueLast) {
		return mediaQueueLast(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Queue.RemoveFromQueue) {
		return mediaDeleteSongFromQueue(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Queue.ClearQueue) {
		return mediaClearQueue(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Queue.MoveUp) {
		return mediaSongUpQueue(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Queue.MoveDown) {
		return mediaSongDownQueue(m), nil
	}

	// FAVORITES KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Favorites.ToggleFavorite) {
		return mediaToggleFavorite(m)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Favorites.ViewFavorites) {
		return mediaShowFavorites(m, msg)
	}

	// OTHER KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Other.CreateShareLink) {
		return m, mediaCreateShare(m)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Other.ToggleNotifications) {
		return toggleNotifications(m), nil
	}

	return m, nil
}

func keyMatches(key string, bindings []string) bool {
	for _, k := range bindings {
		if k == key {
			return true
		}
	}
	return false
}

func typeInput(m model, msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func hardQuit(m model) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func quit(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.focus != focusSearch {
		return m, tea.Quit
	} else {
		return typeInput(m, msg)
	}
}

func focusSearchBar(m model) model {
	m.focus = focusSearch
	m.textInput.SetValue("")
	m.textInput.Focus()
	return m
}

func cycleFocus(m model, forward bool) model {
	// Cycles Focus: Search -> Sidebar -> Main -> Song -> Search
	if forward {
		m.focus = (m.focus + 1) % 4
	} else {
		m.focus = (((m.focus-1)%4 + 4) % 4)
	}

	if m.focus == focusSearch {
		m.textInput.Focus()
	} else {
		m.textInput.Blur()
	}

	return m
}

func enter(m model) (tea.Model, tea.Cmd) {
	switch m.focus {
	case focusSearch:
		query := m.textInput.Value()
		if query != "" {
			m.loading = true
			m.focus = focusMain
			m.viewMode = viewList
			m.textInput.Blur()

			// Reset paging
			m.pageOffset = 0
			m.pageHasMore = true
			m.lastSearchQuery = query

			// Reset selection
			m = resetSelection(m)

			switch m.filterMode {
			case filterSongs:
				m.displayMode = displaySongs
			case filterAlbums:
				m.displayMode = displayAlbums
			case filterArtist:
				m.displayMode = displayArtist
			}

			return m, searchCmd(query, m.filterMode, 0)
		}

	case focusMain:
		if m.viewMode == viewList {
			switch m.displayMode {
			// Play song
			case filterSongs:
				if len(m.songs) > 0 {
					return m, m.setQueue(m.cursorMain)
				}

			// Open songs in album
			case filterAlbums:
				if len(m.albums) > 0 {
					selectedAlbum := m.albums[m.cursorMain]
					m.loading = true

					// Save state
					m.displayModePrev = m.displayMode
					m.albumsPrev = m.albums
					m.cursorMainPrev = m.cursorMain
					m.mainOffsetPrev = m.mainOffset

					// New state
					m.displayMode = displaySongs
					m.songs = nil
					m.cursorMain = 0
					m.mainOffset = 0
					m.pageOffset = 0
					m.pageHasMore = true
					m.lastSearchQuery = ""

					// Reset selection
					m = resetSelection(m)

					return m, getAlbumSongs(selectedAlbum.ID, false)
				}

				// Open albums of artist
			case filterArtist:
				if len(m.artists) > 0 {
					selectedArtist := m.artists[m.cursorMain]
					m.loading = true

					// Save state
					m.displayModePrev = m.displayMode
					m.artistsPrev = m.artists
					m.cursorMainPrev = m.cursorMain
					m.mainOffsetPrev = m.mainOffset

					// New state
					m.displayMode = displayAlbums
					m.albums = nil
					m.cursorMain = 0
					m.mainOffset = 0
					m.pageOffset = 0
					m.pageHasMore = true
					m.lastSearchQuery = ""

					// Reset selection
					m = resetSelection(m)

					return m, getArtistAlbums(selectedArtist.ID)
				}
			}
		} else {
			// Queue View: Jump to selected song
			if len(m.queue) > 0 {
				return m, m.playQueueIndex(m.cursorMain, false)
			}
		}

	case focusSidebar:
		albumOffset := len(albumTypes)

		m.loading = true
		m.focus = focusMain
		m.viewMode = viewList

		// Reset selection
		m = resetSelection(m)

		if m.cursorSide < albumOffset {
			m.displayMode = displayAlbums
			// Initialize pagination state
			m.pageOffset = 0
			m.pageHasMore = true
			m.lastSearchQuery = ""
			switch m.cursorSide {
			case 0:
				m.albumListType = "alphabeticalByArtist"
				return m, getAlbumList("alphabeticalByArtist", 0)
			case 1:
				m.albumListType = "random"
				return m, getAlbumList("random", 0)
			case 2:
				m.albumListType = "starred"
				return m, getAlbumList("starred", 0)
			case 3:
				m.albumListType = "newest"
				return m, getAlbumList("newest", 0)
			case 4:
				m.albumListType = "recent"
				return m, getAlbumList("recent", 0)
			case 5:
				m.albumListType = "frequent"
				return m, getAlbumList("frequent", 0)
			}

		} else {
			m.displayMode = displaySongs
			return m, getPlaylistSongs((m.playlists[m.cursorSide-albumOffset]).ID, false) // - because of the Album offset

		}

	}

	return m, nil
}

func toggleSelection(m model) (tea.Model, tea.Cmd) {
	if m.showSelection {
		m.showSelection = false
		m.selectionAnchor = -1
	} else {
		m.showSelection = true
		m.selectionAnchor = m.cursorMain
	}

	// Clear map
	m.selectionMap = make(map[int]bool)

	return m, nil
}

func playShuffled(m model) (tea.Model, tea.Cmd) {
	switch m.focus {
	case focusMain:
		if m.displayMode == displayAlbums && m.cursorMain < len(m.albums) && (m.albums[m.cursorMain]).ID != "" {
			m.loading = true

			return m, getAlbumSongs(m.albums[m.cursorMain].ID, true)
		}

	case focusSidebar:
		if m.cursorSide > (len(albumTypes)-1) && (m.playlists[m.cursorSide-len(albumTypes)]).ID != "" {
			m.loading = true
			m.displayMode = displaySongs
			m.focus = focusMain

			return m, getPlaylistSongs((m.playlists[m.cursorSide-len(albumTypes)]).ID, true)
		}
	}

	return m, nil
}

func goBack(m model) (tea.Model, tea.Cmd) {
	if m.showHelp || m.showPlaylists || m.showRating {
		m.showHelp = false
		m.showPlaylists = false
		m.showRating = false

		return m, nil
	}

	if m.viewMode == viewQueue {
		return toggleQueue(m), nil
	}

	// Swap display modes
	tempDisplay := m.displayMode
	m.displayMode = m.displayModePrev
	m.displayModePrev = tempDisplay

	// Swap attributes
	tempSongs := m.songs
	m.songs = m.songsPrev
	m.songsPrev = tempSongs

	tempAlbums := m.albums
	m.albums = m.albumsPrev
	m.albumsPrev = tempAlbums

	tempArtists := m.artists
	m.artists = m.artistsPrev
	m.artistsPrev = tempArtists

	// Swap offsets
	tempMainOffset := m.mainOffset
	m.mainOffset = m.mainOffsetPrev
	m.mainOffsetPrev = tempMainOffset

	// Swap cursors
	tempCursorMain := m.cursorMain
	m.cursorMain = m.cursorMainPrev
	m.cursorMainPrev = tempCursorMain

	m.viewMode = viewList

	return m, nil
}

func navigateTop(m model) model {
	switch m.focus {
	case focusMain:
		m.cursorMain = 0
		m.mainOffset = 0

		if m.showSelection {
			m = selectionScroller(m)
		}

	case focusSidebar:
		m.cursorSide = 0
		m.sideOffset = 0
	}

	return m
}

func navigateBottom(m model) (model, tea.Cmd) {
	switch m.focus {
	case focusMain:

		listLen := 0

		switch m.displayMode {
		case displaySongs:
			if m.viewMode == viewQueue {
				listLen = len(m.queue)
			} else {
				listLen = len(m.songs)
			}
		case displayAlbums:
			listLen = len(m.albums)
		case displayArtist:
			listLen = len(m.artists)
		}

		m.cursorMain = listLen - 1
		if m.height-17 >= 17 && listLen >= 17 {
			m.mainOffset = listLen - 17
		} else {
			m.mainOffset = 0
		}

		if m.showSelection {
			m = selectionScroller(m)
		}

	case focusSidebar:
		total := len(albumTypes) + len(m.playlists)
		m.cursorSide = total - 1

		headerHeight := 1

		footerHeight := int(float64(m.height) * 0.10)
		if footerHeight < 5 {
			footerHeight = 5
		}

		mainHeight := m.height - headerHeight - footerHeight - (3 * 2) // 3 sections with each 2 borders (top and bottom)
		if mainHeight < 0 {
			mainHeight = 0
		}

		visibleRows := mainHeight - 6 // Conservative estimate for headers
		if visibleRows < 1 {
			visibleRows = 1
		}

		if total > visibleRows {
			m.sideOffset = total - visibleRows
		} else {
			m.sideOffset = 0
		}
	}

	return loadMore(m)
}

func navigateUp(m model, steps int) model {
	switch m.focus {
	case focusMain:
		m.cursorMain -= steps
		if m.cursorMain < 0 {
			m.cursorMain = 0
		}
		if m.cursorMain < m.mainOffset {
			m.mainOffset = m.cursorMain
		}

		if m.showSelection {
			m = selectionScroller(m)
		}

	case focusSidebar:
		m.cursorSide -= steps
		if m.cursorSide < 0 {
			m.cursorSide = 0
		}
		if m.cursorSide < m.sideOffset {
			m.sideOffset = m.cursorSide
		}
		if m.cursorSide < m.sideOffset {
			m.sideOffset = m.cursorSide
		}
	}

	return m
}

func navigateDown(m model, steps int) (model, tea.Cmd) {
	listLen := 0
	if m.viewMode == viewQueue {
		listLen = len(m.queue)
	} else if m.displayMode == displaySongs {
		listLen = len(m.songs)
	} else if m.displayMode == displayAlbums {
		listLen = len(m.albums)
	} else if m.displayMode == displayArtist {
		listLen = len(m.artists)
	}

	albumOffset := len(albumTypes)

	switch m.focus {
	case focusMain:
		if m.cursorMain < listLen-1 {
			m.cursorMain += steps

			if m.cursorMain > listLen-1 {
				m.cursorMain = listLen - 1
			}

			// Height - Search(3) - Footer(6) - Margins(4) - TableHeader(2) = 17
			visibleRows := m.height - 17
			if m.cursorMain >= m.mainOffset+visibleRows {
				m.mainOffset = m.cursorMain - visibleRows + 1
			}
		}

		if m.showSelection {
			m = selectionScroller(m)
		}

	case focusSidebar:
		if m.cursorSide < len(m.playlists)+albumOffset-1 { // + because of the Album offset
			m.cursorSide += steps

			if m.cursorSide > len(m.playlists)+albumOffset-1 {
				m.cursorSide = len(m.playlists) + albumOffset - 1
			}

			headerHeight := 1

			footerHeight := int(float64(m.height) * 0.10)
			if footerHeight < 5 {
				footerHeight = 5
			}

			mainHeight := m.height - headerHeight - footerHeight - (3 * 2) // 3 sections with each 2 borders (top and bottom)
			if mainHeight < 0 {
				mainHeight = 0
			}

			visibleRows := mainHeight - 6 // Conservative estimate for headers
			if visibleRows < 1 {
				visibleRows = 1
			}

			if m.cursorSide >= m.sideOffset+visibleRows {
				m.sideOffset = m.cursorSide - visibleRows + 1
			}
		}
	}

	// Check to see if more has to be loaded
	return loadMore(m)
}

func displayAlbumFromSelected(m model) (tea.Model, tea.Cmd) {
	var id string

	switch m.focus {
	case focusMain:
		if m.viewMode == viewList && m.displayMode == displaySongs && len(m.songs) > 0 {
			id = m.songs[m.cursorMain].AlbumID // album id of selected song
		} else if m.viewMode == viewQueue && len(m.queue) > 0 {
			id = m.queue[m.cursorMain].AlbumID // album id of a queued song
		}

	case focusSong:
		if len(m.queue) > 0 {
			id = m.queue[m.queueIndex].AlbumID // album id of playing song
		}
	}

	// Return on no ID
	if id == "" {
		return m, nil
	}

	// Reset model
	m.loading = true

	// Save state
	m.displayModePrev = m.displayMode
	m.songsPrev = m.songs
	m.cursorMainPrev = m.cursorMain
	m.mainOffsetPrev = m.mainOffset

	// Set new state
	m.viewMode = viewList
	m.displayMode = displaySongs
	m.songs = nil
	m.cursorMain = 0
	m.mainOffset = 0
	m.lastSearchQuery = ""

	return m, getAlbumSongs(id, false)
}

func displayArtistFromSelected(m model) (tea.Model, tea.Cmd) {
	var id string

	switch m.focus {
	case focusMain:
		switch m.displayMode {
		case displaySongs:
			if m.viewMode == viewList && len(m.songs) > 0 {
				id = m.songs[m.cursorMain].ArtistID // artist id of selected songs
			} else if m.viewMode == viewQueue && len(m.queue) > 0 {
				id = m.songs[m.queueIndex].ArtistID // artist id of queued song
			}

		case displayAlbums:
			if len(m.albums) > 0 {
				id = m.albums[m.cursorMain].ArtistID
			}
		}

	case focusSong:
		if len(m.queue) > 0 {
			id = m.queue[m.queueIndex].ArtistID
		}

	}

	// Return on no ID
	if id == "" {
		return m, nil
	}

	// Reset model
	m.loading = true

	// Save state
	m.displayModePrev = m.displayMode
	m.songsPrev = m.songs
	m.albumsPrev = m.albums
	m.mainOffsetPrev = m.mainOffset
	m.cursorMainPrev = m.cursorMain
	m.mainOffsetPrev = m.mainOffset

	// New state
	m.viewMode = viewList
	m.displayMode = displayAlbums
	m.songs = nil
	m.albums = nil
	m.cursorMain = 0
	m.mainOffset = 0
	m.lastSearchQuery = ""

	return m, getArtistAlbums(id)
}

func cycleFilter(m model, forward bool) model {
	if m.focus == focusSearch {
		if forward {
			m.filterMode = (m.filterMode + 1) % 3
		} else {
			m.filterMode = ((m.filterMode-1)%3 + 3) % 3
		}

		switch m.filterMode {
		case filterSongs:
			m.textInput.Placeholder = "Search songs..."
		case filterAlbums:
			m.textInput.Placeholder = "Search albums..."
		case filterArtist:
			m.textInput.Placeholder = "Search artists..."
		}
	}

	return m
}

func toggleQueue(m model) model {
	if m.focus != focusSearch {
		switch m.viewMode {
		case viewList:
			m.viewMode = viewQueue
			m.displayModePrev = m.displayMode
			m.displayMode = displaySongs
			m.cursorMain = m.queueIndex
			if m.cursorMain > 2 {
				m.mainOffset = m.cursorMain - 2
			} else {
				m.mainOffset = 0
			}
		case viewQueue:
			m.viewMode = viewList
			m.displayMode = m.displayModePrev
			m.cursorMain = 0
			m.mainOffset = 0
		}

		// Reset selection
		m = resetSelection(m)
	}

	return m
}

func mediaTogglePlay(m model, msg tea.Msg) model {
	_, isMpris := msg.(integration.PlayPauseMsg)
	if m.focus != focusSearch || isMpris {
		player.TogglePause()
		m.playerStatus.Paused = !m.playerStatus.Paused

		if m.dbusInstance != nil {
			var newStatus string
			if m.playerStatus.Paused {
				newStatus = "Paused"
			} else {
				newStatus = "Playing"
			}

			m.dbusInstance.UpdateStatus(newStatus)
		}
	}

	return m
}

func mediaSongSkip(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	_, isMpris := msg.(integration.NextSongMsg)
	if m.focus != focusSearch || isMpris {
		return m, tea.Batch(
			m.playNext(),
		)
	} else {
		return typeInput(m, msg)
	}
}

func mediaSongPrev(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	_, isMpris := msg.(integration.PreviousSongMsg)
	if m.focus != focusSearch || isMpris {
		return m, m.playPrev()
	} else {
		return typeInput(m, msg)
	}
}

func mediaVolumeUp(m model, _ tea.Msg) (tea.Model, tea.Cmd) {
	player.VolumeUp()
	m.playerStatus.Volume = player.GetVolume()
	return m, nil
}

func mediaVolumeDown(m model, _ tea.Msg) (tea.Model, tea.Cmd) {
	player.VolumeDown()
	m.playerStatus.Volume = player.GetVolume()
	return m, nil
}

func mediaQueueNext(m model) model {
	if m.focus == focusMain && (m.displayMode == displaySongs || m.displayMode == displayAlbums) {
		selectedSongs := getSelectedSongs(m)

		if selectedSongs != nil {
			if len(m.queue) == 0 { // Create a new queue
				m.queue = selectedSongs
				m.queueIndex = 0
			} else { // Add to current queue
				insertAt := m.queueIndex + 1
				tail := append([]api.Song{}, m.queue[insertAt:]...)
				m.queue = append(m.queue[:insertAt], append(selectedSongs, tail...)...)
			}

			if m.viewMode == viewQueue && m.cursorMain > m.queueIndex {
				m.cursorMain++
			}
		}
	}

	// Sync MPV's Queue
	m.syncNextSong()

	return m
}

func mediaQueueLast(m model) model {
	if m.focus == focusMain && (m.displayMode == displaySongs || m.displayMode == displayAlbums) {
		selectedSongs := getSelectedSongs(m)

		if selectedSongs != nil {
			m.queue = append(m.queue, selectedSongs...)
		}

	}

	// Sync MPV's Queue
	m.syncNextSong()

	return m
}

func mediaDeleteSongFromQueue(m model) model {
	if m.focus != focusMain || m.viewMode != viewQueue || len(m.queue) == 0 {
		return m
	}

	// Get songs to delete
	toDelete := make(map[int]bool)
	if m.showSelection && len(m.selectionMap) > 0 {
		for k := range m.selectionMap {
			toDelete[k] = true
		}
	} else {
		toDelete[m.cursorMain] = true
	}

	// Do not remove current playing song
	if toDelete[m.queueIndex] {
		delete(toDelete, m.queueIndex)
	}

	// Return if nothing to delete
	if len(toDelete) == 0 {
		return m
	}

	var newQueue []api.Song
	newQueueIndex := m.queueIndex

	for i, song := range m.queue {
		if toDelete[i] {
			// Shift index if song before index is deleted
			if i < m.queueIndex {
				newQueueIndex--
			}
		} else {
			newQueue = append(newQueue, song) // Keep the song
		}
	}

	// Set new queue
	m.queue = newQueue
	m.queueIndex = newQueueIndex

	// Update cursor if out of bounds
	if m.cursorMain >= len(m.queue) {
		m.cursorMain = len(m.queue) - 1
		if m.cursorMain < 0 {
			m.cursorMain = 0
		}
	}

	if m.showSelection {
		m = resetSelection(m)
	}

	// Sync MPV's Queue
	m.syncNextSong()

	return m
}
func mediaClearQueue(m model) model {
	if m.focus == focusMain {
		m.queue = nil
		m.queueIndex = 0
	}

	if m.playerStatus.Paused {
		player.Stop()
	}

	// Sync MPV's Queue
	m.syncNextSong()

	return m
}

func mediaSongUpQueue(m model) model {
	if m.focus != focusMain || m.viewMode != viewQueue {
		return m
	}

	minIndex, maxIndex := m.cursorMain, m.cursorMain
	if m.showSelection && len(m.selectionMap) > 0 {
		minIndex, maxIndex = -1, -1
		for i := range m.selectionMap {
			if minIndex == -1 || i < minIndex {
				minIndex = i
			}
			if maxIndex == -1 || i > maxIndex {
				maxIndex = i
			}
		}
	}

	if minIndex > 0 {
		target := m.queue[minIndex-1]

		copy(m.queue[minIndex-1:maxIndex], m.queue[minIndex:maxIndex+1])
		m.queue[maxIndex] = target

		if m.queueIndex == minIndex-1 {
			m.queueIndex = maxIndex
		} else if m.queueIndex >= minIndex && m.queueIndex <= maxIndex {
			m.queueIndex--
		}

		if m.showSelection {
			newMap := make(map[int]bool)
			for k := range m.selectionMap {
				newMap[k-1] = true
			}
			m.selectionMap = newMap
			m.selectionAnchor--
		}

		m.cursorMain--
	}

	m.syncNextSong()
	return m
}

func mediaSongDownQueue(m model) model {
	if m.focus != focusMain || m.viewMode != viewQueue {
		return m
	}

	minIndex, maxIndex := m.cursorMain, m.cursorMain
	if m.showSelection && len(m.selectionMap) > 0 {
		minIndex, maxIndex = -1, -1
		for i := range m.selectionMap {
			if minIndex == -1 || i < minIndex {
				minIndex = i
			}
			if maxIndex == -1 || i > maxIndex {
				maxIndex = i
			}
		}
	}

	if maxIndex < len(m.queue)-1 {
		target := m.queue[maxIndex+1]

		copy(m.queue[minIndex+1:maxIndex+2], m.queue[minIndex:maxIndex+1])
		m.queue[minIndex] = target

		if m.queueIndex == maxIndex+1 {
			m.queueIndex = minIndex
		} else if m.queueIndex >= minIndex && m.queueIndex <= maxIndex {
			m.queueIndex++
		}

		if m.showSelection {
			newMap := make(map[int]bool)
			for k := range m.selectionMap {
				newMap[k+1] = true
			}
			m.selectionMap = newMap
			m.selectionAnchor++
		}

		m.cursorMain++
	}

	m.syncNextSong()
	return m
}
func mediaRestartSong(m model) model {
	if m.focus != focusSearch {
		player.RestartSong()
	}

	return m
}

func mediaSeekForward(m model) model {
	if m.focus != focusSearch {
		player.Forward10Seconds()
	}

	return m
}

func mediaToggleMediaPlayer(m model) model {
	if m.focus != focusSearch {
		if m.showMediaPlayer {
			m.showMediaPlayer = false
		} else {
			m.showMediaPlayer = true
		}

		if m.coverArt != nil {
			resModel, _ := m.handleCoverArt(coverArtMsg{
				img: m.coverArt,
			})
			if updatedModel, ok := resModel.(model); ok {
				m = updatedModel
			}
		}
	}

	return m
}

func mediaSeekRewind(m model) model {
	if m.focus != focusSearch {
		player.Back10Seconds()
	}

	return m
}

func mediaShuffle(m model) model {
	if m.focus != focusSearch {
		if len(m.queue) < 2 {
			return m
		}

		newQueue := make([]api.Song, len(m.queue))
		copy(newQueue, m.queue)
		m.queue = newQueue

		currentSongID := ""
		if m.queueIndex >= 0 && m.queueIndex < len(m.queue) {
			currentSongID = m.queue[m.queueIndex].ID
		}

		rand.Shuffle(len(m.queue), func(i, j int) {
			m.queue[i], m.queue[j] = m.queue[j], m.queue[i]
		})

		if currentSongID != "" {
			for i, song := range m.queue {
				if song.ID == currentSongID {
					// Swap the song at i with 0 to set current song to first
					m.queue[0], m.queue[i] = m.queue[i], m.queue[0]
					m.queueIndex = 0
					break
				}
			}
		} else {
			m.queueIndex = 0
		}
	}

	// Sync MPV's Queue
	m.syncNextSong()

	return m
}

func mediaToggleLoop(m model) model {
	if m.focus != focusSearch {
		m.loopMode = (m.loopMode + 1) % 3
	}

	// Sync MPV's Queue
	m.syncNextSong()

	return m
}

func mediaToggleFavorite(m model) (model, tea.Cmd) {
	var ids []string
	var idsToStar []string
	var idsToUnstar []string

	switch m.focus {
	case focusMain: // Focus main view
		ids = selectionIdsGetter(m)

	case focusSong: // Focus footer
		if len(m.queue) > 0 {
			ids = []string{m.queue[m.queueIndex].ID}
		}
	}

	// Return on no ID
	if len(ids) == 0 {
		return m, nil
	}

	// Toggle favorite status
	for i := range ids {
		id := ids[i]
		isStarred := m.starredMap[id]
		if isStarred {
			delete(m.starredMap, id)
			idsToUnstar = append(idsToUnstar, id)
		} else {
			m.starredMap[id] = true
			idsToStar = append(idsToStar, id)
		}
	}

	return m, toggleStarCmd(idsToStar, idsToUnstar)
}

func mediaShowFavorites(m model, msg tea.Msg) (model, tea.Cmd) {
	if m.focus == focusSearch {
		return typeInput(m, msg)
	}

	m.displayMode = displaySongs

	m.songs = nil
	m.viewMode = viewList
	m.focus = focusMain

	return m, openLikedSongsCmd()
}

func toggleAddToPlaylistPopup(m model) model {
	if m.focus == focusMain && m.displayMode == displaySongs &&
		((m.viewMode == viewList && len(m.songs) > 0) || (m.viewMode == viewQueue && len(m.queue) > 0)) {
		m.showPlaylists = !m.showPlaylists

		if m.showPlaylists {
			m.cursorPopup = 0
		}

	}

	return m
}

func toggleAddRatingPopup(m model) model {
	if m.focus != focusMain {
		return m
	}

	switch m.displayMode {
	case displaySongs:
		if m.viewMode == viewList && len(m.songs) > 0 && m.songs[m.cursorMain].ID != "" {
			m.cursorPopup = m.songs[m.cursorMain].Rating
			m.showRating = !m.showRating
		} else if m.viewMode == viewQueue && len(m.queue) > 0 && m.queue[m.cursorMain].ID != "" {
			m.cursorPopup = m.queue[m.cursorMain].Rating
			m.showRating = !m.showRating
		}
	case displayAlbums:
		if len(m.albums) > 0 && m.albums[m.cursorMain].ID != "" {
			m.cursorPopup = m.albums[m.cursorMain].Rating
			m.showRating = !m.showRating
		}
	case displayArtist:
		if len(m.artists) > 0 && m.artists[m.cursorMain].ID != "" {
			m.cursorPopup = m.artists[m.cursorMain].Rating
			m.showRating = !m.showRating
		}
	}

	return m
}

func mediaCreateShare(m model) tea.Cmd {
	if m.focus != focusMain {
		return nil
	}

	var ids []string
	switch m.displayMode {
	case displaySongs:
		var targetList []api.Song

		switch m.viewMode {
		case viewList:
			targetList = m.songs
		case viewQueue:
			targetList = m.queue
		}

		if len(targetList) > 0 {
			if m.showSelection { // Add selection
				for i := range m.selectionMap {
					ids = append(ids, targetList[i].ID)
				}
			} else { // Add single song
				ids = []string{targetList[m.cursorMain].ID}
			}
		}

	case displayAlbums: // Albums
		if len(m.albums) > 0 {
			if m.showSelection { // Add selection
				for i := range m.selectionMap {
					ids = append(ids, m.albums[i].ID)
				}
			} else { // Add single album
				ids = []string{m.albums[m.cursorMain].ID}
			}
		}
	}

	if len(ids) > 0 {
		return createMediaShareCmd(ids)
	}

	return nil
}

func toggleNotifications(m model) model {
	if m.focus != focusSearch {
		m.notify = !m.notify
	}

	return m
}

func lyricsUp(m model) model {
	if m.songLinesOffset > 0 {
		m.songLinesOffset = m.songLinesOffset - 1
	}

	return m
}

func lyricsDown(m model) model {
	if len(m.songLyrics) > 0 && m.songLinesOffset < len(m.songLyrics[0].Lines) {
		m.songLinesOffset = m.songLinesOffset + 1
	}

	return m
}

func (m *model) updateLoginInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.loginInputs))
	for i := range m.loginInputs {
		m.loginInputs[i], cmds[i] = m.loginInputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func login(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter":
		m.loading = true
		m.loginErr = ""

		domain := m.loginInputs[0].Value()
		if !strings.Contains(domain, "http") {
			m.loginErr = "Please include the protocol at the start 'http(s)'"
			return m, nil
		}
		api.AppServerConfig.Server.URL = strings.TrimSuffix(domain, "/")

		switch m.loginType {
		case loginPassword:
			username := m.loginInputs[1].Value()
			password := m.loginInputs[2].Value()

			if domain == "" || username == "" || password == "" {
				m.loginErr = "All fields are required"
				return m, nil
			}

			api.AppServerConfig.Server.AuthMethod = "plaintext"
			api.AppServerConfig.Server.Username = username
			api.AppServerConfig.Server.Password = password
			api.AppServerConfig.Server.PasswordToken = ""
			api.AppServerConfig.Server.PasswordSalt = ""
			api.AppServerConfig.Server.ApiKey = ""

		case loginPasswordHashed:
			username := m.loginInputs[1].Value()
			passwordToken := m.loginInputs[2].Value()
			passwordsalt := m.loginInputs[3].Value()

			if domain == "" || username == "" || passwordToken == "" || passwordsalt == "" {
				m.loginErr = "All fields are required"
				return m, nil
			}

			api.AppServerConfig.Server.AuthMethod = "hashed"
			api.AppServerConfig.Server.Username = username
			api.AppServerConfig.Server.Password = ""
			api.AppServerConfig.Server.PasswordToken = passwordToken
			api.AppServerConfig.Server.PasswordSalt = passwordsalt
			api.AppServerConfig.Server.ApiKey = ""

		case loginApi:
			username := m.loginInputs[1].Value()
			apiKey := m.loginInputs[2].Value()

			if domain == "" || username == "" || apiKey == "" {
				m.loginErr = "All fields are required"
				return m, nil
			}

			api.AppServerConfig.Server.AuthMethod = "api_key"
			api.AppServerConfig.Server.Username = username
			api.AppServerConfig.Server.Password = ""
			api.AppServerConfig.Server.PasswordToken = ""
			api.AppServerConfig.Server.PasswordSalt = ""
			api.AppServerConfig.Server.ApiKey = apiKey
		}

		return m, tea.Batch(
			attemptLoginCmd(),
		)

	case "up", "down", "tab", "shift+tab":
		if key == "up" || key == "shift+tab" {
			switch m.loginType {
			case loginPassword:
				m.loginFocus = ((m.loginFocus-1)%3 + 3) % 3

			case loginPasswordHashed:
				m.loginFocus = ((m.loginFocus-1)%4 + 4) % 4

			case loginApi:
				m.loginFocus = ((m.loginFocus-1)%3 + 3) % 3
			}
		} else {
			switch m.loginType {
			case loginPassword:
				m.loginFocus = (m.loginFocus + 1) % 3

			case loginPasswordHashed:
				m.loginFocus = (m.loginFocus + 1) % 4

			case loginApi:
				m.loginFocus = (m.loginFocus + 1) % 3
			}
		}

	case "ctrl+t":
		m.loginType = (m.loginType + 1) % 3

		// Clear old inputs
		m.loginInputs[1].SetValue("")
		m.loginInputs[2].SetValue("")
		m.loginInputs[3].SetValue("")

		// Correct the inputs with fitting values
		switch m.loginType {
		case loginPassword:
			m.loginInputs[1].Prompt = "Username: "
			m.loginInputs[1].Placeholder = "username"
			m.loginInputs[1].EchoMode = textinput.EchoNormal

			m.loginInputs[2].Prompt = "Password: "
			m.loginInputs[2].Placeholder = "password"
			m.loginInputs[2].EchoMode = textinput.EchoPassword

		case loginPasswordHashed:
			m.loginInputs[1].Prompt = "Username: "
			m.loginInputs[1].Placeholder = "username"
			m.loginInputs[1].EchoMode = textinput.EchoNormal

			m.loginInputs[2].Prompt = "Token: "
			m.loginInputs[2].Placeholder = "md5 hash"
			m.loginInputs[2].EchoMode = textinput.EchoNormal

			m.loginInputs[3].Prompt = "Salt: "
			m.loginInputs[3].Placeholder = "random string"
			m.loginInputs[3].EchoMode = textinput.EchoNormal

		case loginApi:
			m.loginInputs[1].Prompt = "Username: "
			m.loginInputs[1].Placeholder = "username"
			m.loginInputs[1].EchoMode = textinput.EchoNormal

			m.loginInputs[2].Prompt = "API Key: "
			m.loginInputs[2].Placeholder = "api key"
			m.loginInputs[2].EchoMode = textinput.EchoPassword

			if m.loginFocus > 1 {
				m.loginFocus = 1
			}
		}
	}

	// Focus the correct login input
	for i := 0; i <= len(m.loginInputs)-1; i++ {
		if i == m.loginFocus {
			m.loginInputs[i].Focus()
		} else {
			m.loginInputs[i].Blur()
		}
	}

	return m, m.updateLoginInputs(msg)
}

func playerMenu(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if keyMatches(key, api.AppConfig.Keybinds.Global.Help) {
		m.showHelp = !m.showHelp
		return m, nil
	} else if m.showHelp {
		return m, nil
	}

	// GLOBAL KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Global.Quit) {
		return quit(m, msg)
	}

	// NAVIGATION KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Media.ToggleMediaPlayer) || keyMatches(key, api.AppConfig.Keybinds.Global.Back) {
		return mediaToggleMediaPlayer(m), nil
	}

	// MEDIA KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Media.PlayPause) {
		return mediaTogglePlay(m, msg), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Next) {
		return mediaSongSkip(m, msg)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Prev) {
		return mediaSongPrev(m, msg)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.VolumeUp) {
		return mediaVolumeUp(m, msg)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.VolumeDown) {
		return mediaVolumeDown(m, msg)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Shuffle) {
		return mediaShuffle(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Loop) {
		return mediaToggleLoop(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Restart) {
		return mediaRestartSong(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Rewind) {
		return mediaSeekRewind(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Media.Forward) {
		return mediaSeekForward(m), nil
	}

	// OTHER KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Other.ToggleNotifications) {
		return toggleNotifications(m), nil
	}

	// Navigation
	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Up) {
		return lyricsUp(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Down) {
		return lyricsDown(m), nil
	}

	return m, nil
}

func playlistsMenu(key string, m model) (model, tea.Cmd) {
	if keyMatches(key, api.AppConfig.Keybinds.Global.Back) || keyMatches(key, api.AppConfig.Keybinds.Library.AddToPlaylist) {
		m.showPlaylists = false
		m.cursorPopup = 0
		return m, nil
	}

	if len(m.playlists) == 0 {
		return m, nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Up) {
		if m.cursorPopup > 0 {
			m.cursorPopup--
		}
	} else if keyMatches(key, api.AppConfig.Keybinds.Navigation.Down) {
		if m.cursorPopup < len(m.playlists)-1 {
			m.cursorPopup++
		}
	} else if keyMatches(key, api.AppConfig.Keybinds.Navigation.Select) {
		var targetList []api.Song
		var ids []string
		inBounds := cursorInBounds(m)

		// Cursor out of bounds
		if !inBounds {
			return m, nil
		}

		switch m.viewMode {
		case viewList:
			targetList = m.songs
		case viewQueue:
			targetList = m.queue
		}

		// No target list
		if len(targetList) == 0 {
			return m, nil
		}

		if m.showSelection { // Add selection
			for i := range m.selectionMap {
				ids = append(ids, targetList[i].ID)
			}
		} else { // Add single song
			ids = []string{targetList[m.cursorMain].ID}
		}

		// Toggle playlist view
		m.showPlaylists = !m.showPlaylists
		return m, addSongToPlaylistCmd(m.playlists[m.cursorPopup].ID, ids)
	}

	return m, nil
}

func ratingMenu(key string, m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	if keyMatches(key, api.AppConfig.Keybinds.Global.Back) || keyMatches(key, api.AppConfig.Keybinds.Library.AddRating) {
		m.showRating = false
		m.cursorPopup = 0
		return m, nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Up) && m.cursorPopup > 0 {
		m.cursorPopup--
	} else if keyMatches(key, api.AppConfig.Keybinds.Navigation.Down) && m.cursorPopup < 5 {
		m.cursorPopup++
	} else if keyMatches(key, api.AppConfig.Keybinds.Navigation.Select) && cursorInBounds(m) {
		m, cmd = setRating(m, m.cursorPopup)

		// Reset popup
		m.cursorPopup = 0
		m.showRating = !m.showRating

		return m, cmd
	}

	return m, nil
}

func setRating(m model, rating int) (model, tea.Cmd) {
	if !cursorInBounds(m) {
		return m, nil
	}

	var ids []string
	switch m.displayMode {
	case displaySongs:
		var targetList []api.Song
		switch m.viewMode {
		case viewList:
			targetList = m.songs
		case viewQueue:
			targetList = m.queue
		}

		if m.showSelection { // Add selection
			for i := range m.selectionMap {
				targetList[i].Rating = rating
				ids = append(ids, targetList[i].ID)
			}
		} else { // Add single album
			targetList[m.cursorMain].Rating = rating
			ids = []string{targetList[m.cursorMain].ID}
		}

	case displayAlbums:
		if m.showSelection { // Add selection
			for i := range m.selectionMap {
				m.albums[i].Rating = rating
				ids = append(ids, m.albums[i].ID)
			}
		} else { // Add single album
			m.albums[m.cursorMain].Rating = rating
			ids = []string{m.albums[m.cursorMain].ID}
		}

	case displayArtist:
		if m.showSelection { // Add selection
			for i := range m.selectionMap {
				m.artists[i].Rating = rating
				ids = append(ids, m.artists[i].ID)
			}
		} else { // Add single artist
			m.artists[m.cursorMain].Rating = rating
			ids = []string{m.artists[m.cursorMain].ID}
		}
	}

	return m, addRatingCmd(ids, rating)
}

// Helper for infinte scrolling
func loadMore(m model) (model, tea.Cmd) {
	if m.focus == focusMain && m.pageHasMore && !m.loading {
		// Songs
		if m.displayMode == displaySongs && len(m.songs)-m.cursorMain <= 10 && m.lastSearchQuery != "" {
			m.loading = true
			m.pageOffset += 150
			return m, searchCmd(m.lastSearchQuery, filterSongs, m.pageOffset)
		}

		// Albums
		if m.displayMode == displayAlbums && len(m.albums)-m.cursorMain <= 10 {
			m.loading = true
			m.pageOffset += 150

			// Check if search or sidebar loading
			if m.lastSearchQuery != "" {
				return m, searchCmd(m.lastSearchQuery, filterAlbums, m.pageOffset)
			} else {
				return m, getAlbumList(m.albumListType, m.pageOffset)
			}
		}

		// Artists
		if m.displayMode == displayArtist && len(m.artists)-m.cursorMain <= 10 && m.lastSearchQuery != "" {
			m.loading = true
			m.pageOffset += 150
			return m, searchCmd(m.lastSearchQuery, filterArtist, m.pageOffset)
		}
	}

	return m, nil
}

// Helper for checking if the cursor in bounds
func cursorInBounds(m model) bool {
	switch m.displayMode {
	case displaySongs:
		switch m.viewMode {
		case viewList:
			return m.cursorMain >= 0 && m.cursorMain < len(m.songs)

		case viewQueue:
			return m.cursorMain >= 0 && m.cursorMain < len(m.queue)
		}

	case displayAlbums:
		return m.cursorMain >= 0 && m.cursorMain < len(m.albums)

	case displayArtist:
		return m.cursorMain >= 0 && m.cursorMain < len(m.artists)
	}

	return false
}

// Helper for highlighting selections
func selectionScroller(m model) model {
	if m.showSelection {
		start := m.selectionAnchor
		end := m.cursorMain

		// Swap if scrolling up
		if start > end {
			start, end = end, start
		}

		// Clear selection
		m.selectionMap = make(map[int]bool)

		// Select everything in range
		for i := start; i <= end; i++ {
			m.selectionMap[i] = true
		}
	}
	return m
}

// Helper for getting all ID's from a selection
func selectionIdsGetter(m model) []string {
	var ids []string

	switch m.displayMode {
	case displaySongs: // Songs
		var targetList []api.Song
		switch m.viewMode {
		case viewList:
			targetList = m.songs
		case viewQueue:
			targetList = m.queue
		}

		if len(targetList) > 0 {
			if m.showSelection { // Add selection
				for i := range m.selectionMap {
					ids = append(ids, targetList[i].ID)
				}
			} else { // Add single song
				ids = []string{targetList[m.cursorMain].ID}
			}
		}

	case displayAlbums: // Albums
		if len(m.albums) > 0 {
			if m.showSelection { // Add selection
				for i := range m.selectionMap {
					ids = append(ids, m.albums[i].ID)
				}
			} else { // Add single album
				ids = []string{m.albums[m.cursorMain].ID}
			}
		}

	case displayArtist: // Artists
		if len(m.artists) > 0 {
			if m.showSelection { // Add selection
				for i := range m.selectionMap {
					ids = append(ids, m.artists[i].ID)
				}
			} else { // Add single artist
				ids = []string{m.artists[m.cursorMain].ID}
			}
		}
	}

	return ids
}

// Helper for resetting selection state
func resetSelection(m model) model {
	m.showSelection = false
	m.selectionAnchor = -1
	m.selectionMap = make(map[int]bool)

	return m
}
