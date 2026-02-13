package ui

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/MattiaPun/SubTUI/v2/internal/integration"
	"github.com/MattiaPun/SubTUI/v2/internal/player"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/beeep"
)

func (m model) handleWindowResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	return m, nil
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

		m.viewMode = viewLogin
		m.loginInputs[0].SetValue(api.AppServerConfig.Server.URL)
		m.loginInputs[1].SetValue(api.AppServerConfig.Server.Username)
		m.loginInputs[2].SetValue(api.AppServerConfig.Server.Password)

		m.loginFocus = 0
		m.loginInputs[0].Focus()
		m.loginInputs[1].Blur()
		m.loginInputs[2].Blur()

		return m, nil
	}

	// Login Success
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

	if m.playerStatus.Path == "" || m.playerStatus.Path == "<nil>" || len(m.queue) == 0 {

		m.queue = []api.Song{}
		m.lastPlayedSongID = ""

		// MRPIS Update
		if m.dbusInstance != nil {
			m.dbusInstance.ClearMetadata()
		}

		return m, syncPlayerCmd()
	}

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
				Rating:   math.Round(float64(currentSong.Rating*10)) / 10,
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
			if m.discordRPC && m.discordInstance != nil {
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

	if m.playerStatus.Path != "" &&
		m.playerStatus.Path != "<nil>" &&
		len(m.queue) > 0 &&
		!strings.Contains(m.playerStatus.Path, "id="+m.queue[m.queueIndex].ID) {

		nextIndex := m.queueIndex + 1
		m.scrobbled = false

		// Queue next song
		if nextIndex < len(m.queue) {
			m.queueIndex = nextIndex
		}

		nextNextIndex := -1
		switch m.loopMode {
		case LoopOne:
			nextNextIndex = nextIndex
		case LoopNone:
			nextNextIndex = nextIndex + 1
		case LoopAll:
			if nextIndex == len(m.queue)-1 {
				nextNextIndex = 0
			} else {
				nextNextIndex = nextIndex + 1
			}
		}

		// Queue next next song
		if nextNextIndex < len(m.queue) {
			player.UpdateNextSong(m.queue[nextNextIndex].ID)
		} else { // End of queue, clear MPV
			go player.UpdateNextSong("")
		}
	}

	windowTitle := "SubTUI"
	if m.playerStatus.Title != "" && m.playerStatus.Title != "<nil>" && !strings.Contains(m.playerStatus.Title, "stream?c=SubTUI") {
		windowTitle = fmt.Sprintf("%s - %s", m.playerStatus.Title, m.playerStatus.Artist)
	}

	return m, tea.Batch(syncPlayerCmd(), tea.SetWindowTitle(windowTitle))
}

func (m model) handleSongResult(msg songsResultMsg) (tea.Model, tea.Cmd) {
	m.loading = false
	m.focus = focusMain

	if m.pageOffset > 0 { // Append: paging
		m.songs = append(m.songs, msg.songs...)
	} else { // Replace: no paging
		m.songs = msg.songs
		m.cursorMain = 0
		m.mainOffset = 0
	}

	m.pageHasMore = (len(msg.songs) == 150)

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
