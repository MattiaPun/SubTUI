package ui

import (
	"git.punjwani.pm/Mattia/DepthTUI/internal/api"
	"git.punjwani.pm/Mattia/DepthTUI/internal/player"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	// Handle Window Resize
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// Key Presses
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		// Tab Cycles Focus: Search -> Sidebar -> Main -> Song -> Search
		case "tab":
			m.focus = (m.focus + 1) % 4
			if m.focus == focusSearch {
				m.textInput.Focus()
			} else {
				m.textInput.Blur()
			}

		case "enter":
			if m.focus == focusSearch {
				// Trigger Search
				query := m.textInput.Value()
				if query != "" {
					m.loading = true
					m.focus = focusMain
					m.textInput.Blur()
					return m, searchSongsCmd(query)
				}
			} else if m.focus == focusMain {
				if m.viewMode == viewList {
					// List View: Load list into queue
					if len(m.songs) > 0 {
						return m, m.setQueue(m.songs, m.cursorMain)
					}
				} else {
					// Queue View: Jump to selected song
					if len(m.queue) > 0 {
						return m, m.playQueueIndex(m.cursorMain)
					}
				}
			} else if m.focus == focusSidebar {
				// Open playlist
				m.loading = true
				m.focus = focusMain
				m.viewMode = viewList
				return m, getPlaylistSongs((m.playlists[m.cursorSide]).ID)
			}

		case "up", "k": // Navigation up
			if m.focus == focusMain && m.cursorMain > 0 {
				m.cursorMain--
				if m.cursorMain < m.mainOffset {
					m.mainOffset = m.cursorMain
				}
			} else if m.focus == focusSidebar && m.cursorSide > 0 {
				m.cursorSide--
			}

		case "down", "j": // Navigation down
			if m.focus == focusMain && m.cursorMain < len(m.songs)-1 {
				m.cursorMain++

				// Height - Search(3) - Footer(6) - Margins(4) - TableHeader(2)
				visibleRows := m.height - 17

				if m.cursorMain >= m.mainOffset+visibleRows {
					m.mainOffset++
				}
			} else if m.focus == focusSidebar && m.cursorSide < len(m.playlists)-1 {
				m.cursorSide++
			}

		case "Q":
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

		case "p": // Play/pause
			if m.focus != focusSearch {
				player.TogglePause()
			}

		case "n": // Next song
			return m, m.playNext()

		case "b": // Previous song
			return m, m.playPrev()

		case "N": // Play next
			selectedSong := m.songs[m.cursorMain]

			if len(m.queue) == 0 {
				m.queue = []api.Song{selectedSong}
				m.queueIndex = 0
			} else {
				insertAt := m.queueIndex + 1
				tail := append([]api.Song{}, m.queue[insertAt:]...)
				m.queue = append(m.queue[:insertAt], append([]api.Song{selectedSong}, tail...)...)
			}

		case "a": // Add to queue
			m.queue = append(m.queue, m.songs[m.cursorMain])

		case "d": // Delete from queue
			m.queue = append(m.queue[:m.queueIndex], m.queue[m.queueIndex+1:]...)

		case "D": // Clear queue
			m.queue = nil

		case "ctrl+k": // Move up in queue
			tempSong := m.queue[m.cursorMain]

			m.queue[m.cursorMain] = m.queue[m.cursorMain-1]
			m.queue[m.cursorMain-1] = tempSong

		case "ctrl+j": // Move down in queue
			tempSong := m.queue[m.cursorMain]

			m.queue[m.cursorMain] = m.queue[m.cursorMain+1]
			m.queue[m.cursorMain+1] = tempSong

		case ",": // -10sec
			player.Back10Seconds()

		case ";": // +10sec
			player.Forward10Seconds()
		}

	case songsResultMsg:
		m.loading = false
		m.songs = msg.songs
		m.cursorMain = 0
		m.mainOffset = 0
		m.focus = focusMain
		m.viewMode = viewList
		m.textInput.Blur()

	case playlistResultMsg:
		m.playlists = msg.playlists

	case errMsg:
		m.loading = false
		m.err = msg.err

	case statusMsg:
		m.playerStatus = player.PlayerStatus(msg)
		if m.playerStatus.Duration > 0 &&
			m.playerStatus.Current >= m.playerStatus.Duration-1 &&
			!m.playerStatus.Paused {

			return m, tea.Batch(
				m.playNext(),
				syncPlayerCmd(),
			)
		}

		return m, syncPlayerCmd()
	}

	// Update inputs
	if m.focus == focusSearch {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}
