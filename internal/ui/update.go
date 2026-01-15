package ui

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/MattiaPun/SubTUI/internal/api"
	"github.com/MattiaPun/SubTUI/internal/integration"
	"github.com/MattiaPun/SubTUI/internal/player"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/beeep"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	// Handle Window Resize
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.viewMode == viewLogin {
			return m, nil
		}

	// Key Presses
	case tea.KeyMsg:

		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.viewMode == viewLogin {
			return login(m, msg)
		}

		if msg.String() == "?" {
			m.showHelp = !m.showHelp
			return m, nil
		} else if m.showHelp {
			return m, nil
		}

		if (msg.String() == "g" || m.lastKey == "g") && (m.focus == focusMain || m.focus == focusSidebar) {
			switch msg.String() {
			case "g":
				if m.lastKey == "g" {
					return navigateTop(m), nil
				} else {
					m.lastKey = "g"
					return m, nil
				}
			case "a":
				return displaySongAlbum(m)
			case "r":
				return displaySongArtist(m)
			default:
				m.lastKey = ""
			}
		}

		switch msg.String() {
		case "q":
			return quit(m, msg)

		case "/":
			return focusSearchBar(m), nil

		case "tab":
			m = cycleFocus(m, true)

		case "shift+tab":
			m = cycleFocus(m, false)

		case "enter":
			return enter(m)

		case "backspace":
			return goBack(m, msg)

		case "G":
			m = navigateBottom(m)

		case "up", "k":
			m = navigateUp(m)

		case "down", "j":
			m = navigateDown(m)

		case "ctrl+n":
			m = cycleFilter(m, true)

		case "ctrl+b":
			m = cycleFilter(m, false)

		case "Q":
			m = toggleQueue(m)

		case "p", "P":
			m = mediaTogglePlay(m, msg)

		case "n":
			return mediaSongSkip(m, msg)

		case "b":
			return mediaSongPrev(m, msg)

		case "N":
			m = mediaAddSongNext(m)

		case "a":
			m = mediaAddSongToQueue(m)

		case "d":
			m = mediaDeleteSongFromQueue(m)

		case "D":
			m = mediaDeleteQueue(m)

		case "K":
			m = mediaSongUpQueue(m)

		case "J":
			m = mediaSongDownQueue(m)

		case "w":
			m = mediaRestartSong(m)

		case ",":
			m = mediaSeekRewind(m)

		case ";":
			m = mediaSeekForward(m)

		case "S":
			m = mediaShuffle(m)

		case "L":
			m = mediaToggleLoop(m)

		case "f":
			return mediaToggleFavorite(m, msg)

		case "F":
			return mediaShowFavorites(m, msg)

		case "s":
			m = toggleNotifications(m)
		}

	case loginResultMsg:
		if msg.err != nil {
			log.Printf("[Login] Failure: %v", msg.err)
		} else {
			log.Printf("[Login] Success. Switching to Main View.")
		}

		m.loading = false
		if msg.err != nil { // login failed
			errMsg := msg.err.Error()

			if strings.Contains(strings.ToLower(errMsg), "network") || strings.Contains(strings.ToLower(errMsg), "tls") || strings.Contains(strings.ToLower(errMsg), "remote") {
				m.loginErr = "Host not found. Please check URL/Connection."
			} else if strings.Contains(errMsg, "Wrong username") {
				m.loginErr = "Invalid Credentials"
			} else {
				m.loginErr = errMsg
			}

			m.viewMode = viewLogin
			m.loginInputs[0].SetValue(api.AppConfig.URL)
			m.loginInputs[1].SetValue(api.AppConfig.Username)
			m.loginInputs[2].SetValue(api.AppConfig.Password)

			m.loginFocus = 0
			m.loginInputs[0].Focus()
			m.loginInputs[1].Blur()
			m.loginInputs[2].Blur()
		} else { // login success
			if err := player.InitPlayer(); err != nil {
				m.loginErr = fmt.Sprintf("Audio Engine Error: %v", err)
				return m, nil
			}

			m.viewMode = viewList
			m.focus = focusMain
			m.loginErr = ""

			return m, tea.Batch(
				syncPlayerCmd(),
				getPlaylists(),
				getPlayQueue(),
				getStarredCmd(),
			)
		}

	case playlistResultMsg:
		m.playlists = msg.playlists

	case errMsg:
		m.loading = false
		m.err = msg.err

	case statusMsg:
		if len(m.queue) > 0 {
			currentSong := m.queue[m.queueIndex]

			if currentSong.ID != m.lastPlayedSongID {
				m.lastPlayedSongID = currentSong.ID
				m.scrobbled = false

				// Setup metadata
				metadata := integration.Metadata{
					Title:    currentSong.Title,
					Artist:   currentSong.Artist,
					Album:    currentSong.Album,
					Duration: float64(currentSong.Duration), // Cast int to float64
					ImageURL: api.SubsonicCoverArtUrl(currentSong.ID, 500),
				}

				// System notification
				if m.notify {
					go func() {
						artBytes, err := api.SubsonicCoverArt(currentSong.ID)

						title := "SubTUI"
						description := fmt.Sprintf("Playing %s - %s", currentSong.Title, currentSong.Artist)

						if err != nil {
							_ = beeep.Notify(title, description, "")
						} else {
							_ = beeep.Notify(title, description, artBytes)
						}
					}()
				}

				// MRPIS Update
				if m.dbusInstance != nil {
					m.dbusInstance.UpdateMetadata(metadata)
				}

				// Discord Update
				if m.discordInstance != nil {
					m.discordInstance.UpdateActivity(metadata)
				}
			}
		}

		if len(m.queue) > 0 && m.queueIndex >= 0 && !m.scrobbled {
			currentSong := m.queue[m.queueIndex]

			pos := m.playerStatus.Current
			dur := m.playerStatus.Duration

			if dur > 0 {
				target := math.Min(dur/2, 240)

				if pos >= target {
					m.scrobbled = true

					go api.SubsonicScrobble(currentSong.ID, true)
				}
			}
		}

		m.playerStatus = player.PlayerStatus(msg)

		if ((m.playerStatus.Duration > 0 && m.playerStatus.Current >= m.playerStatus.Duration-0.5) ||
			(m.playerStatus.Title == "<nil>" && m.queueIndex != len(m.queue)-1)) &&
			!m.playerStatus.Paused {

			switch m.loopMode {
			case LoopNone:
				//
			case LoopAll:
				if m.queueIndex == len(m.queue)-1 {
					m.queueIndex = -1
				}

			case LoopOne:
				m.queueIndex = m.queueIndex - 1

			}

			return m, tea.Batch(
				m.playNext(),
				syncPlayerCmd(),
			)
		}

		windowTitle := "SubTUI"
		if m.playerStatus.Title != "" && m.playerStatus.Title != "<nil>" && !strings.Contains(m.playerStatus.Title, "stream?c=SubTUI") {
			windowTitle = fmt.Sprintf("%s - %s", m.playerStatus.Title, m.playerStatus.Artist)
		}

		return m, tea.Batch(syncPlayerCmd(), tea.SetWindowTitle(windowTitle))

	case songsResultMsg:
		m.loading = false
		m.songs = msg.songs
		m.cursorMain = 0
		m.mainOffset = 0
		m.focus = focusMain

	case albumsResultMsg:
		m.loading = false
		m.albums = msg.albums
		m.cursorMain = 0
		m.mainOffset = 0
		m.focus = focusMain

	case artistsResultMsg:
		m.loading = false
		m.artists = msg.artists
		m.cursorMain = 0
		m.mainOffset = 0
		m.focus = focusMain

	case starredResultMsg:
		for _, s := range msg.result.Songs {
			m.starredMap[s.ID] = true
		}
		for _, a := range msg.result.Albums {
			m.starredMap[a.ID] = true
		}
		for _, r := range msg.result.Artists {
			m.starredMap[r.ID] = true
		}

	case viewLikedSongsMsg:
		for _, s := range msg.Songs {
			m.starredMap[s.ID] = true
		}
		for _, a := range msg.Albums {
			m.starredMap[a.ID] = true
		}

		m.songs = msg.Songs

		return m, nil

	case playQueueResultMsg:
		for index, song := range msg.result.Entries {
			m.queue = append(m.queue, song)

			if song.ID == msg.result.Current {
				m.queueIndex = index
			}
		}

		return m, m.playQueueIndex(m.queueIndex, true)

	case SetDBusMsg:
		m.dbusInstance = msg.Instance
		return m, nil

	case integration.PlayPauseMsg:
		m = mediaTogglePlay(m, msg)

	case integration.NextSongMsg:
		return mediaSongSkip(m, msg)

	case integration.PreviousSongMsg:
		return mediaSongPrev(m, msg)

	case SetDiscordMsg:
		m.discordInstance = msg.Instance
		return m, nil

	}

	// Update inputs
	if m.focus == focusSearch {
		m, cmd = typeInput(m, msg)
	}

	return m, cmd
}

func typeInput(m model, msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
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

// Cycles Focus: Search -> Sidebar -> Main -> Song -> Search
func cycleFocus(m model, forward bool) model {
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

			switch m.filterMode {
			case filterSongs:
				m.displayMode = displaySongs
			case filterAlbums:
				m.displayMode = displayAlbums
			case filterArtist:
				m.displayMode = displayArtist
			}

			return m, searchCmd(query, m.filterMode)
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
					m.displayMode = displaySongs
					m.songs = nil

					return m, getAlbumSongs(selectedAlbum.ID)
				}

			// Open albums of artist
			case filterArtist:
				if len(m.artists) > 0 {
					selectedArtist := m.artists[m.cursorMain]
					m.loading = true
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
			switch m.cursorSide {
			case 0:
				return m, getAlbumList("random")
			case 1:
				return m, getAlbumList("starred")
			case 2:
				return m, getAlbumList("newest")
			case 3:
				return m, getAlbumList("recent")
			case 4:
				return m, getAlbumList("frequent")
			}

		} else {
			m.displayMode = displaySongs
			return m, getPlaylistSongs((m.playlists[m.cursorSide-albumOffset]).ID) // - because of the Album offset

		}

	}

	return m, nil
}

func goBack(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.focus == focusSearch {
		return typeInput(m, msg)
	}

	if m.viewMode == viewQueue {
		return toggleQueue(m), nil
	}

	m.displayMode = m.displayModePrev
	m.displayModePrev = m.displayMode

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
	}

	return m
}

func navigateBottom(m model) model {
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
		m.cursorSide = len(albumTypes) - 1 + len(m.playlists) - 1
	}

	return m
}

func navigateUp(m model) model {
	if m.focus == focusMain && m.cursorMain > 0 {
		m.cursorMain--
		if m.cursorMain < m.mainOffset {
			m.mainOffset = m.cursorMain
		}
	} else if m.focus == focusSidebar && m.cursorSide > 0 {
		m.cursorSide--
	}

	return m
}

func navigateDown(m model) model {
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
	}

	return m
}

func displaySongAlbum(m model) (tea.Model, tea.Cmd) {
	var targetList []api.Song
	switch m.viewMode {
	case viewList:
		targetList = m.songs
	case viewQueue:
		targetList = m.queue
	}
	albumCmd := getAlbumSongs(targetList[m.cursorMain].AlbumID)

	m.viewMode = viewList
	m.displayModePrev = m.displayMode
	m.displayMode = displaySongs
	m.mainOffset = 0
	m.cursorMain = 0
	m.loading = true

	return m, albumCmd
}

func displaySongArtist(m model) (tea.Model, tea.Cmd) {
	var targetList []api.Song
	switch m.viewMode {
	case viewList:
		targetList = m.songs
	case viewQueue:
		targetList = m.queue
	}
	albumCmd := getArtistAlbums(targetList[m.cursorMain].ArtistID)

	m.viewMode = viewList
	m.displayModePrev = m.displayMode
	m.displayMode = displayAlbums
	m.mainOffset = 0
	m.cursorMain = 0
	m.loading = true

	return m, albumCmd
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

func mediaAddSongNext(m model) model {
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

	return m
}

func mediaAddSongToQueue(m model) model {
	if m.focus == focusMain {
		selectedSongs := getSelectedSongs(m)

		if selectedSongs != nil {
			m.queue = append(m.queue, selectedSongs...)
		}

	}

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

	return m
}

func mediaDeleteQueue(m model) model {
	if m.focus == focusMain {
		m.queue = nil
		m.queueIndex = 0
	}

	return m
}

func mediaSongUpQueue(m model) model {
	if m.focus == focusMain && m.viewMode == viewQueue && m.cursorMain > 0 {
		tempSong := m.queue[m.cursorMain]

		m.queue[m.cursorMain] = m.queue[m.cursorMain-1]
		m.queue[m.cursorMain-1] = tempSong

		m.cursorMain--
	}

	return m
}

func mediaSongDownQueue(m model) model {
	if m.focus == focusMain && m.viewMode == viewQueue && m.cursorMain < len(m.queue)-1 {
		tempSong := m.queue[m.cursorMain]

		m.queue[m.cursorMain] = m.queue[m.cursorMain+1]
		m.queue[m.cursorMain+1] = tempSong

		m.cursorMain++
	}

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

	return m
}

func mediaToggleLoop(m model) model {
	if m.focus != focusSearch {
		m.loopMode = (m.loopMode + 1) % 3
	}

	return m
}

func mediaToggleFavorite(m model, msg tea.Msg) (model, tea.Cmd) {
	if m.focus == focusSearch {
		return typeInput(m, msg)
	}

	id := ""

	switch m.filterMode {
	case filterSongs:

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
	case filterAlbums:
		if len(m.albums) > 0 {
			id = m.albums[m.cursorMain].ID
		}
	case filterArtist:
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

				api.AppConfig.URL = strings.TrimSuffix(domain, "/")
				api.AppConfig.Username = username
				api.AppConfig.Password = password

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
