package ui

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"git.punjwani.pm/Mattia/DepthTUI/internal/api"
	"git.punjwani.pm/Mattia/DepthTUI/internal/player"
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

		switch msg.String() {
		case "q":
			return quit(m, msg)

		case "tab":
			m = cycleFocus(m, true)

		case "shift+tab":
			m = cycleFocus(m, false)

		case "enter":
			return enter(m)

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

		case "p":
			m = mediaTogglePlay(m)

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

		case ",":
			m = mediaSeekForward(m)

		case ";":
			m = mediaSeekRewind(m)

		case "S":
			m = mediaShuffle(m)

		case "L":
			m = mediaToggleLoop(m)

		case "F":
			return mediaToggleFavorite(m, msg)
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

				title := "DepthTUI"
				description := fmt.Sprintf("Playing %s - %s", currentSong.Title, currentSong.Artist)
				beeep.Notify(title, description, "")
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
		if m.playerStatus.Duration > 0 &&
			m.playerStatus.Current >= m.playerStatus.Duration-1 &&
			!m.playerStatus.Paused {

			switch m.loopMode {
			case LoopNone:
				//
			case LoopAll:
				m.queueIndex = -1
			case LoopOne:
				m.queueIndex = m.queueIndex - 1

			}

			return m, tea.Batch(
				m.playNext(),
				syncPlayerCmd(),
			)
		}

		return m, syncPlayerCmd()

	case songsResultMsg:
		m.loading = false
		m.songs = msg.songs
		m.albums = nil
		m.artists = nil
		m.cursorMain = 0
		m.mainOffset = 0
		m.focus = focusMain

	case albumsResultMsg:
		m.loading = false
		m.albums = msg.albums
		m.songs = nil
		m.artists = nil
		m.cursorMain = 0
		m.mainOffset = 0
		m.focus = focusMain

	case artistsResultMsg:
		m.loading = false
		m.artists = msg.artists
		m.songs = nil
		m.albums = nil
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
	if m.focus == focusSearch {
		query := m.textInput.Value()
		if query != "" {
			m.loading = true
			m.focus = focusMain
			m.viewMode = viewList
			m.textInput.Blur()
			m.songs = nil
			m.albums = nil
			m.artists = nil

			return m, searchCmd(query, m.filterMode)
		}
	} else if m.focus == focusMain {
		if m.viewMode == viewList {
			switch m.filterMode {
			// Play song
			case filterSongs:
				if len(m.songs) > 0 {
					return m, m.setQueue(m.songs, m.cursorMain)
				}

			// Open songs in album
			case filterAlbums:
				if len(m.albums) > 0 {
					selectedAlbum := m.albums[m.cursorMain]
					m.loading = true
					m.filterMode = filterSongs
					m.songs = nil

					return m, getAlbumSongs(selectedAlbum.ID)
				}

			// Open albums of artist
			case filterArtist:
				if len(m.artists) > 0 {
					selectedArtist := m.artists[m.cursorMain]
					m.loading = true
					m.filterMode = filterAlbums
					m.albums = nil

					return m, getArtistAlbums(selectedArtist.ID)
				}
			}

		} else {
			// Queue View: Jump to selected song
			if len(m.queue) > 0 {
				return m, m.playQueueIndex(m.cursorMain)
			}
		}
	} else if m.focus == focusSidebar {
		m.loading = true
		m.focus = focusMain
		m.viewMode = viewList
		m.filterMode = filterSongs
		return m, getPlaylistSongs((m.playlists[m.cursorSide]).ID)
	}

	return m, nil
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
	} else if m.filterMode == filterSongs {
		listLen = len(m.songs)
	} else if m.filterMode == filterAlbums {
		listLen = len(m.albums)
	} else if m.filterMode == filterArtist {
		listLen = len(m.artists)
	}

	if m.focus == focusMain && m.cursorMain < listLen-1 {
		m.cursorMain++

		// Height - Search(3) - Footer(6) - Margins(4) - TableHeader(2) = 17
		visibleRows := m.height - 17
		if m.cursorMain >= m.mainOffset+visibleRows {
			m.mainOffset++
		}
	} else if m.focus == focusSidebar && m.cursorSide < len(m.playlists)-1 {
		m.cursorSide++
	}

	return m
}

func cycleFilter(m model, forward bool) model {
	if m.focus == focusSearch {
		if forward {
			m.filterMode = (m.filterMode + 1) % 3
		} else {
			m.filterMode = ((m.filterMode-1)%3 + 3) % 3
		}
	}

	return m
}

func toggleQueue(m model) model {
	if m.focus != focusSearch {

		if m.viewMode == viewList {
			m.viewMode = viewQueue
			m.cursorMain = m.queueIndex
			if m.cursorMain > 2 {
				m.mainOffset = m.cursorMain - 2
			} else {
				m.mainOffset = 0
			}
		} else {
			m.viewMode = viewList
			m.cursorMain = 0
			m.mainOffset = 0
		}
	}

	return m
}

func mediaTogglePlay(m model) model {
	if m.focus != focusSearch {
		player.TogglePause()
	}

	return m
}

func mediaSongSkip(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.focus != focusSearch {
		return m, m.playNext()
	} else {
		return typeInput(m, msg)
	}
}

func mediaSongPrev(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.focus != focusSearch {
		return m, m.playPrev()
	} else {
		return typeInput(m, msg)
	}
}

func mediaAddSongNext(m model) model {
	if m.focus == focusMain {
		selectedSong := m.songs[m.cursorMain]

		if len(m.queue) == 0 {
			m.queue = []api.Song{selectedSong}
			m.queueIndex = 0
		} else {
			insertAt := m.queueIndex + 1
			tail := append([]api.Song{}, m.queue[insertAt:]...)
			m.queue = append(m.queue[:insertAt], append([]api.Song{selectedSong}, tail...)...)
		}
	}

	return m
}

func mediaAddSongToQueue(m model) model {
	if m.focus == focusMain {
		m.queue = append(m.queue, m.songs[m.cursorMain])
	}

	return m
}

func mediaDeleteSongFromQueue(m model) model {
	if m.focus == focusMain && m.viewMode == viewQueue && len(m.queue) > 0 {
		m.queue = append(m.queue[:m.cursorMain], m.queue[m.cursorMain+1:]...)
	}

	if m.cursorMain >= len(m.queue) && m.cursorMain > 0 {
		m.cursorMain--
	}

	return m
}

func mediaDeleteQueue(m model) model {
	if m.focus == focusMain {
		m.queue = nil
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

func mediaSeekForward(m model) model {
	if m.focus != focusSearch {
		player.Back10Seconds()
	}

	return m
}

func mediaSeekRewind(m model) model {
	if m.focus != focusSearch {
		player.Forward10Seconds()
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
		if len(m.songs) > 0 {
			id = m.songs[m.cursorMain].ID
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

				api.AppConfig.Domain = m.loginInputs[0].Value()
				api.AppConfig.Username = m.loginInputs[1].Value()
				api.AppConfig.Password = m.loginInputs[2].Value()

				if err := api.SaveConfig(); err != nil {
					m.err = err
					return m, nil
				}

				player.InitPlayer()
				m.viewMode = viewList
				m.focus = focusMain

				return m, tea.Batch(
					checkLoginCmd(),
					getPlaylists(),
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
