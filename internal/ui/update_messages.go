package ui

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/MattiaPun/SubTUI/v2/internal/integration"
	"github.com/MattiaPun/SubTUI/v2/internal/player"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/mosaic"
	"github.com/gen2brain/beeep"
	zone "github.com/lrstanley/bubblezone"
)

const doubleClickThreshold = 500 * time.Millisecond

func (m model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	return m, nil
}

func (m model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionRelease || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}

	headerHeight := 1
	footerHeight := int(float64(m.height) * 0.10)
	if footerHeight < 5 {
		footerHeight = 5
	}
	mainHeight := m.height - headerHeight - footerHeight - (3 * 2) // borders
	sidebarWidth := int(float64(m.width) * 0.25)

	listStartY := headerHeight + 2
	if msg.Y < listStartY { // Header
		m.focus = focusSearch
		m.textInput.Focus()

		if zone.Get("filter_prev").InBounds(msg) {
			return cycleFilter(m, false), nil
		}

		if zone.Get("filter_next").InBounds(msg) {
			return cycleFilter(m, true), nil
		}

		return m, nil
	} else if msg.Y > listStartY+mainHeight { // Footer
		m.focus = focusSong
		m.textInput.Blur()

		return m, nil
	}

	if msg.X < sidebarWidth { // Sidebar
		m.focus = focusSidebar
		m.textInput.Blur()

		totalItems := len(albumTypes) + len(m.playlists)
		endIndex := m.sideOffset + mainHeight
		if endIndex > totalItems {
			endIndex = totalItems
		}

		for i := m.sideOffset; i < endIndex; i++ {
			id := fmt.Sprintf("sidebar_item_%d", i)

			if zone.Get(id).InBounds(msg) {
				m.cursorSide = i

				if isDoubleClick(m, id) {
					return enter(m)
				}

				m.lastClickTime = time.Now()
				m.lastClickId = id
				return m, nil
			}
		}
	} else if msg.X >= sidebarWidth { // Main view
		m.focus = focusMain
		m.textInput.Blur()

		var mainListItemsCount int
		switch m.displayMode {
		case displaySongs:
			mainListItemsCount = len(m.songs)
			if m.viewMode == viewQueue {
				mainListItemsCount = len(m.queue)
			}
		case displayAlbums:
			mainListItemsCount = len(m.albums)
		case displayArtist:
			mainListItemsCount = len(m.artists)
		}

		endIndex := m.mainOffset + mainHeight
		if endIndex > mainListItemsCount {
			endIndex = mainListItemsCount
		}

		for i := m.mainOffset; i < endIndex; i++ {
			id := fmt.Sprintf("mainview_item_%d", i)
			if zone.Get(id).InBounds(msg) {
				m.cursorMain = i

				if isDoubleClick(m, id) {
					return enter(m)
				}

				m.lastClickTime = time.Now()
				m.lastClickId = id
				return m, nil
			}
		}
	}

	return m, nil
}

// Helper for checking for double click's
func isDoubleClick(m model, clickedId string) bool {
	return time.Since(m.lastClickTime) < doubleClickThreshold && clickedId == m.lastClickId
}

func (m model) handleLoginResult(msg loginResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		log.Printf("[Login] Failure: %v", msg.err)
	} else {
		log.Printf("[Login] Success. Switching to Main View.")
	}

	m.loading = false

	// login failed
	if msg.err != nil {
		errMsg := msg.err.Error()

		if strings.Contains(strings.ToLower(errMsg), "network") || strings.Contains(strings.ToLower(errMsg), "tls") || strings.Contains(strings.ToLower(errMsg), "remote") {
			m.loginErr = "Host not found. Please check URL/Connection."
		} else if strings.Contains(errMsg, "Wrong username") {
			m.loginErr = "Invalid Credentials"
		} else {
			m.loginErr = errMsg
		}

		m.loginInputs[0].SetValue(api.AppServerConfig.Server.URL)

		m.viewMode = viewLogin
		switch strings.ToLower(api.AppServerConfig.Server.AuthMethod) {
		case "plaintext":
			m.loginType = 0
			m.loginInputs[1].Prompt = "Username: "
			m.loginInputs[1].Placeholder = "username"
			m.loginInputs[1].EchoMode = textinput.EchoNormal
			m.loginInputs[1].SetValue(api.AppServerConfig.Server.Username)

			m.loginInputs[2].Prompt = "Password: "
			m.loginInputs[2].Placeholder = "password"
			m.loginInputs[2].EchoMode = textinput.EchoPassword
			m.loginInputs[2].SetValue(api.AppServerConfig.Server.Password)

		case "hashed":
			m.loginType = 1
			m.loginInputs[1].Prompt = "Username: "
			m.loginInputs[1].Placeholder = "username"
			m.loginInputs[1].EchoMode = textinput.EchoNormal
			m.loginInputs[1].SetValue(api.AppServerConfig.Server.Username)

			m.loginInputs[2].Prompt = "Token: "
			m.loginInputs[2].Placeholder = "md5 hash"
			m.loginInputs[2].EchoMode = textinput.EchoNormal
			m.loginInputs[2].SetValue(api.AppServerConfig.Server.PasswordToken)

			m.loginInputs[3].Prompt = "Salt: "
			m.loginInputs[3].Placeholder = "random string"
			m.loginInputs[3].EchoMode = textinput.EchoNormal
			m.loginInputs[3].SetValue(api.AppServerConfig.Server.PasswordSalt)

		case "api_key":
			m.loginType = 2
			m.loginInputs[1].Prompt = "Username: "
			m.loginInputs[1].Placeholder = "username"
			m.loginInputs[1].EchoMode = textinput.EchoNormal
			m.loginInputs[1].SetValue(api.AppServerConfig.Server.Username)

			m.loginInputs[2].Prompt = "API Key: "
			m.loginInputs[2].Placeholder = "api key"
			m.loginInputs[2].EchoMode = textinput.EchoPassword
			m.loginInputs[2].SetValue(api.AppServerConfig.Server.ApiKey)

			if m.loginFocus > 1 {
				m.loginFocus = 1
			}
		}

		return m, nil
	}

	// Login Success
	if err := player.InitPlayer(); err != nil {
		m.loginErr = fmt.Sprintf("Audio Engine Error: %v", err)
		return m, nil
	}

	m.viewMode = viewList
	m.focus = focusSearch
	m.loginErr = ""

	return m, tea.Batch(
		syncPlayerCmd(),
		getPlaylists(),
		getPlayQueue(),
		getStarredCmd(),
	)
}

func (m model) handlePlaylistResult(msg playlistResultMsg) (tea.Model, tea.Cmd) {
	m.playlists = msg.playlists
	return m, nil
}

func (m model) handleErr(msg errMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	m.err = msg.err
	return m, nil
}

func (m model) handleStatus(msg statusMsg) (tea.Model, tea.Cmd) {
	m.playerStatus = player.PlayerStatus(msg)
	var cmds []tea.Cmd
	cmds = append(cmds, syncPlayerCmd())

	// Queue ended
	if m.playerStatus.Path == "" || m.playerStatus.Path == "<nil>" || len(m.queue) == 0 {
		m.queue = []api.Song{}
		m.lastPlayedSongPath = ""

		// Clear MRPIS
		if m.dbusInstance != nil {
			m.dbusInstance.ClearMetadata()
		}

		// Clear album art
		if api.AppConfig.Theme.DisplayAlbumArt {
			m.coverArt = nil
		}

		cmds = append(cmds, tea.SetWindowTitle("SubTUI"))
		return m, tea.Batch(cmds...)
	}

	// Song changed
	if m.playerStatus.Path != m.lastPlayedSongPath {

		// Update queue index after mpv song change
		if !strings.Contains(m.playerStatus.Path, "id="+m.queue[m.queueIndex].ID) {
			nextIndex := m.queueIndex + 1

			if nextIndex < len(m.queue) {
				m.queueIndex = nextIndex
			} else if nextIndex >= len(m.queue) && m.loopMode == LoopAll {
				m.queueIndex = 0
			}
		}

		// Update queue
		m.syncNextSong()

		currentSong := m.queue[m.queueIndex]
		m.lastPlayedSongPath = m.playerStatus.Path // Update previous song
		m.scrobbled = false                        // Reset scrobble status

		// Setup metadata
		metadata := integration.Metadata{
			Title:    currentSong.Title,
			Artist:   currentSong.Artist,
			Album:    currentSong.Album,
			Duration: float64(currentSong.Duration),
			ImageURL: api.SubsonicCoverArtUrl(currentSong.ID, 500),
			Rating:   math.Round(float64(currentSong.Rating*10)) / 10,
		}

		// System notification
		if m.notify {
			go func() {
				artBytes, err := api.SubsonicCoverArt(currentSong.ID, 50)

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
		if m.discordRPC && m.discordInstance != nil {
			m.discordInstance.UpdateActivity(metadata)
		}

		// Album Art Update
		if api.AppConfig.Theme.DisplayAlbumArt {
			cmds = append(cmds, getCoverArtCmd(currentSong.ID))
		}

		windowTitle := fmt.Sprintf("%s - %s", metadata.Title, metadata.Artist)
		cmds = append(cmds, tea.SetWindowTitle(windowTitle)) // Update windows title
		cmds = append(cmds, m.savePlayQueue())               // Server queue update
	}

	// Scrobble after half of the song, or 4 minutes, whichever happens first
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

	return m, tea.Batch(cmds...)
}

func (m model) handleSongResult(msg songsResultMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	m.focus = focusMain

	songs := applyExclusionFilters(m, msg.songs)

	if m.pageOffset > 0 { // Append: paging
		m.songs = append(m.songs, msg.songs...)
	} else { // Replace: no paging
		m.songs = songs
		m.cursorMain = 0
		m.mainOffset = 0
	}

	m.pageHasMore = (len(songs) == 150)

	return m, nil
}

func (m model) handleAlbumResult(msg albumsResultMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	m.focus = focusMain
	m.pageHasMore = (len(msg.albums) == 150)

	if m.pageOffset > 0 { // Append: paging
		m.albums = append(m.albums, msg.albums...)
	} else { // Replace: no paging
		m.albums = msg.albums
		m.cursorMain = 0
		m.mainOffset = 0
	}

	return m, nil
}

func (m model) handleArtistsResult(msg artistsResultMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	m.focus = focusMain
	m.pageHasMore = (len(msg.artists) == 150)

	if m.pageOffset > 0 { // Append: paging
		m.artists = append(m.artists, msg.artists...)
	} else { // Replace: no paging
		m.artists = msg.artists
		m.cursorMain = 0
		m.mainOffset = 0
	}

	return m, nil
}

func (m model) handleStarredResult(msg starredResultMsg) (tea.Model, tea.Cmd) {
	for _, s := range msg.result.Songs {
		m.starredMap[s.ID] = true
	}
	for _, a := range msg.result.Albums {
		m.starredMap[a.ID] = true
	}
	for _, r := range msg.result.Artists {
		m.starredMap[r.ID] = true
	}

	return m, nil
}

func (m model) handleViewStarredSongs(msg viewStarredSongsMsg) (tea.Model, tea.Cmd) {
	for _, s := range msg.Songs {
		m.starredMap[s.ID] = true
	}
	for _, a := range msg.Albums {
		m.starredMap[a.ID] = true
	}

	m.songs = msg.Songs
	return m, nil
}

func (m model) handleCoverArt(msg coverArtMsg) (tea.Model, tea.Cmd) {
	m.coverArt = msg.img
	m.coverMosaic = mosaic.New().Width(16).Height(8)
	return m, nil
}

func (m model) handleShuffledSongs(msg shuffledSongsMsg) (tea.Model, tea.Cmd) {
	if msg.updateView {
		m.songs = msg.songs
	}

	songs := applyExclusionFilters(m, msg.songs)

	var filteredSongs []api.Song
	for _, song := range songs {
		if !song.Filtered {
			filteredSongs = append(filteredSongs, song)
		}
	}

	shuffledQueue := make([]api.Song, len(filteredSongs))
	copy(shuffledQueue, filteredSongs)

	rand.Shuffle(len(shuffledQueue), func(i, j int) {
		shuffledQueue[i], shuffledQueue[j] = shuffledQueue[j], shuffledQueue[i]
	})

	m.queue = shuffledQueue
	m.loading = false

	return m, m.playQueueIndex(0, false)
}

func (m model) handleCreateShare(msg createShareMsg) (tea.Model, tea.Cmd) {
	err := clipboard.WriteAll(msg.url)
	if err != nil {
		log.Printf("Failed to write to clipboard")
	}

	return m, nil
}

func (m model) handlePlayQueueResult(msg playQueueResultMsg) (tea.Model, tea.Cmd) {
	for index, song := range msg.result.Entries {
		m.queue = append(m.queue, song)

		if song.ID == msg.result.Current {
			m.queueIndex = index
		}
	}

	return m, m.playQueueIndex(m.queueIndex, true)
}

func (m model) handleSetDBUS(msg SetDBusMsg) (tea.Model, tea.Cmd) {
	m.dbusInstance = msg.Instance

	return m, nil
}

func (m model) handleIntegrationPlayPause(msg integration.PlayPauseMsg) (tea.Model, tea.Cmd) {
	m = mediaTogglePlay(m, msg)

	return m, nil
}

func (m model) handleIntegrationStop() (tea.Model, tea.Cmd) {
	m.queue = nil
	player.Stop()

	return m, nil
}

func (m model) handleIntegrationNextSong(msg integration.NextSongMsg) (tea.Model, tea.Cmd) {
	return mediaSongSkip(m, msg)
}

func (m model) handleIntegrationPreviousSong(msg integration.PreviousSongMsg) (tea.Model, tea.Cmd) {
	return mediaSongPrev(m, msg)
}

func (m model) handleSetDiscord(msg SetDiscordMsg) (tea.Model, tea.Cmd) {
	m.discordInstance = msg.Instance
	return m, nil
}
