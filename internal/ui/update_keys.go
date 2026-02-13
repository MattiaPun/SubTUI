package ui

import (
	"math/rand"
	"strings"
	"time"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/MattiaPun/SubTUI/v2/internal/integration"
	"github.com/MattiaPun/SubTUI/v2/internal/player"
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

	if (key == "g" || m.lastKey == "g") && (m.focus == focusMain || m.focus == focusSidebar) {
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
	if keyMatches(key, api.AppConfig.Keybinds.Global.Back) {
		return goBack(m)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Global.Quit) {
		return quit(m, msg)
	}

	// NAVIGATION KEYBINDS
	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Up) {
		return navigateUp(m), nil
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Down) {
		return navigateDown(m)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Bottom) {
		return navigateBottom(m)
	}

	if keyMatches(key, api.AppConfig.Keybinds.Navigation.Select) {
		return enter(m)
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
		return mediaToggleFavorite(m, msg)
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
					m.displayModePrev = m.displayMode
					m.displayMode = displaySongs
					m.songs = nil

					return m, getAlbumSongs(selectedAlbum.ID)
				}

				// Open albums of artist
			case filterArtist:
				if len(m.artists) > 0 {
					selectedArtist := m.artists[m.cursorMain]
					m.loading = true
					m.displayModePrev = m.displayMode
					m.displayMode = displayAlbums
					m.albums = nil

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
			return m, getPlaylistSongs((m.playlists[m.cursorSide-albumOffset]).ID) // - because of the Album offset

		}

	}

	return m, nil
}

func goBack(m model) (tea.Model, tea.Cmd) {
	if m.viewMode == viewQueue {
		return toggleQueue(m), nil
	}

	tempDisplay := m.displayMode
	m.displayMode = m.displayModePrev
	m.displayModePrev = tempDisplay

	m.viewMode = viewList

	return m, nil
}

func navigateTop(m model) model {
	switch m.focus {
	case focusMain:
		m.cursorMain = 0
		m.mainOffset = 0
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
			listLen = len(m.songs)
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

func navigateUp(m model) model {
	if m.focus == focusMain && m.cursorMain > 0 {
		m.cursorMain--
		if m.cursorMain < m.mainOffset {
			m.mainOffset = m.cursorMain
		}
	} else if m.focus == focusSidebar && m.cursorSide > 0 {
		m.cursorSide--
		if m.cursorSide < m.sideOffset {
			m.sideOffset = m.cursorSide
		}
	}

	return m
}

func navigateDown(m model) (model, tea.Cmd) {
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
	if m.focus == focusMain && m.cursorMain < listLen-1 {
		m.cursorMain++

		// Height - Search(3) - Footer(6) - Margins(4) - TableHeader(2) = 17
		visibleRows := m.height - 17
		if m.cursorMain >= m.mainOffset+visibleRows {
			m.mainOffset++
		}
	} else if m.focus == focusSidebar && m.cursorSide < len(m.playlists)+albumOffset-1 { // + because of the Album offset
		m.cursorSide++

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
			m.sideOffset++
		}
	}

	// Check to see if more has to be loaded
	return loadMore(m)
}

func displayAlbumFromSelected(m model) (tea.Model, tea.Cmd) {
	// Only execute when focused on a song
	if m.focus != focusMain || (m.focus == focusMain && m.displayMode != displaySongs) {
		return m, nil
	}

	albumID := ""
	if m.viewMode == viewList && len(m.songs) != 0 {
		// album of a songs
		albumID = m.songs[m.cursorMain].AlbumID
	} else if m.viewMode == viewQueue && len(m.queue) != 0 {
		// album of a queued song
		albumID = m.queue[m.cursorMain].AlbumID
	}

	if albumID == "" {
		return m, nil
	}

	m.loading = true
	m.mainOffset = 0
	m.cursorMain = 0
	m.lastSearchQuery = ""

	m.viewMode = viewList
	m.displayModePrev = m.displayMode
	m.displayMode = displaySongs

	return m, getAlbumSongs(albumID)
}

func displayArtistFromSelected(m model) (tea.Model, tea.Cmd) {
	// Only execute when focused on a song or album
	if m.focus != focusMain || (m.focus == focusMain && m.displayMode == displayArtist) {
		return m, nil
	}

	artistID := ""
	if m.viewMode == viewList {
		if m.displayMode == displaySongs && len(m.songs) != 0 {
			// artist of a songs
			artistID = m.songs[m.cursorMain].ArtistID
		} else if m.displayMode == displayAlbums && len(m.albums) != 0 {
			// artist of an album
			artistID = m.albums[m.cursorMain].ArtistID
		}
	} else if m.viewMode == viewQueue && len(m.queue) != 0 {
		// artist of a queued song
		artistID = m.queue[m.cursorMain].ArtistID
	}

	if artistID == "" {
		return m, nil
	}

	m.loading = true
	m.mainOffset = 0
	m.cursorMain = 0
	m.lastSearchQuery = ""

	m.viewMode = viewList
	m.displayModePrev = m.displayMode
	m.displayMode = displayAlbums

	return m, getArtistAlbums(artistID)
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
	if m.focus == focusMain {
		selectedSongs := getSelectedSongs(m)

		if selectedSongs != nil {
			if len(m.queue) == 0 {
				m.queue = selectedSongs
				m.queueIndex = 0
			} else {
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
	if m.focus == focusMain {
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
	if m.focus == focusMain && m.viewMode == viewQueue && len(m.queue) > 0 {
		if m.cursorMain != m.queueIndex {
			m.queue = append(m.queue[:m.cursorMain], m.queue[m.cursorMain+1:]...)
			if m.cursorMain < m.queueIndex {
				m.queueIndex--
			}
		}
	}

	if m.cursorMain >= len(m.queue) && m.cursorMain > 0 {
		m.cursorMain--
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

	// Sync MPV's Queue
	m.syncNextSong()

	return m
}

func mediaSongUpQueue(m model) model {
	if m.focus == focusMain && m.viewMode == viewQueue && m.cursorMain > 0 {
		tempSong := m.queue[m.cursorMain]

		m.queue[m.cursorMain] = m.queue[m.cursorMain-1]
		m.queue[m.cursorMain-1] = tempSong

		m.cursorMain--
	}

	// Sync MPV's Queue
	m.syncNextSong()

	return m
}

func mediaSongDownQueue(m model) model {
	if m.focus == focusMain && m.viewMode == viewQueue && m.cursorMain < len(m.queue)-1 {
		tempSong := m.queue[m.cursorMain]

		m.queue[m.cursorMain] = m.queue[m.cursorMain+1]
		m.queue[m.cursorMain+1] = tempSong

		m.cursorMain++
	}

	// Sync MPV's Queue
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

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(m.queue), func(i, j int) {
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

func mediaToggleFavorite(m model, msg tea.Msg) (model, tea.Cmd) {
	if m.focus == focusSearch {
		return typeInput(m, msg)
	}

	id := ""

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
			id = targetList[m.cursorMain].ID
		}
	case displayAlbums:
		if len(m.albums) > 0 {
			id = m.albums[m.cursorMain].ID
		}
	case displayArtist:
		if len(m.artists) > 0 {
			id = m.artists[m.cursorMain].ID
		}
	}

	if id == "" {
		return m, nil
	}

	isStarred := m.starredMap[id]

	if isStarred {
		delete(m.starredMap, id)
	} else {
		m.starredMap[id] = true
	}

	return m, toggleStarCmd(id, isStarred)
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
		if (m.viewMode == viewList && len(m.songs) > 0 && m.songs[m.cursorMain].ID != "") || (m.viewMode == viewQueue && len(m.queue) > 0 && m.queue[m.cursorMain].ID != "") {
			m.showRating = !m.showRating
		}
	case displayAlbums:
		if len(m.albums) > 0 && m.albums[m.cursorMain].ID != "" {
			m.showRating = !m.showRating
		}
	case displayArtist:
		if len(m.artists) > 0 && m.artists[m.cursorMain].ID != "" {
			m.showRating = !m.showRating
		}
	}

	return m
}

func mediaCreateShare(m model) tea.Cmd {
	if m.focus != focusMain {
		return nil
	}

	var id string

	switch {
	case m.viewMode == viewList && m.displayMode == displaySongs && len(m.songs) > 0:
		id = m.songs[m.cursorMain].ID

	case m.viewMode == viewList && m.displayMode == displayAlbums && len(m.albums) > 0:
		id = m.albums[m.cursorMain].ID

	case m.viewMode == viewQueue && len(m.queue) > 0:
		id = m.queue[m.cursorMain].ID
	}

	if id != "" {
		return createMediaShareCmd(id)
	}

	return nil
}

func toggleNotifications(m model) model {
	if m.focus != focusSearch {
		m.notify = !m.notify
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

func login(m model, msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Cycle focus logic
			if s == "enter" && m.loginFocus == len(m.loginInputs)-1 {
				m.loading = true
				m.loginErr = ""

				domain := m.loginInputs[0].Value()
				username := m.loginInputs[1].Value()
				password := m.loginInputs[2].Value()

				if domain == "" || username == "" || password == "" {
					m.loginErr = "All fields are required"
					return m, nil
				}

				if !strings.Contains(domain, "http") {
					m.loginErr = "Please include the protocol at the start 'http(s)'"
					return m, nil
				}

				api.AppServerConfig.Server.URL = strings.TrimSuffix(domain, "/")
				api.AppServerConfig.Server.Username = username
				api.AppServerConfig.Server.Password = password

				return m, tea.Batch(
					attemptLoginCmd(),
				)
			}

			if s == "up" || s == "shift+tab" {
				m.loginFocus--
			} else {
				m.loginFocus++
			}

			if m.loginFocus > len(m.loginInputs)-1 {
				m.loginFocus = 0
			} else if m.loginFocus < 0 {
				m.loginFocus = len(m.loginInputs) - 1
			}

			for i := 0; i <= len(m.loginInputs)-1; i++ {
				if i == m.loginFocus {
					m.loginInputs[i].Focus()
				} else {
					m.loginInputs[i].Blur()
				}
			}
			return m, nil
		}
	}

	return m, m.updateLoginInputs(msg)
}

func playlistsMenu(key string, m model) (model, tea.Cmd) {
	var cmd tea.Cmd
	if keyMatches(key, api.AppConfig.Keybinds.Global.Back) || keyMatches(key, api.AppConfig.Keybinds.Library.AddToPlaylist) {
		m.showPlaylists = false
		m.cursorPopup = 0
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
		if m.viewMode == viewList {
			cmd = addSongToPlaylistCmd(m.songs[m.cursorMain].ID, m.playlists[m.cursorPopup].ID)
		} else {
			cmd = addSongToPlaylistCmd(m.queue[m.cursorMain].ID, m.playlists[m.cursorPopup].ID)
		}
		m.showPlaylists = !m.showPlaylists
		return m, cmd
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
	} else if keyMatches(key, api.AppConfig.Keybinds.Navigation.Select) {
		switch m.displayMode {
		case displaySongs:
			if m.viewMode == viewList {
				cmd = addRatingCmd(m.songs[m.cursorMain].ID, m.cursorPopup)
			} else {
				cmd = addRatingCmd(m.queue[m.cursorMain].ID, m.cursorPopup)
			}
		case displayAlbums:
			cmd = addRatingCmd(m.albums[m.cursorMain].ID, m.cursorPopup)
		case displayArtist:
			cmd = addRatingCmd(m.artists[m.cursorMain].ID, m.cursorPopup)
		}
		m.cursorPopup = 0
		m.showRating = !m.showRating
		return m, cmd
	}

	return m, nil
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
