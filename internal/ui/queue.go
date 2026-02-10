package ui

import (
	"fmt"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/MattiaPun/SubTUI/v2/internal/player"
	tea "github.com/charmbracelet/bubbletea"
)

func formatDuration(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

func (m *model) playQueueIndex(index int, startPaused bool) tea.Cmd {
	if index < 0 || index >= len(m.queue) {
		return nil
	}

	m.queueIndex = index
	song := m.queue[m.queueIndex]

	playCmd := func() tea.Msg {
		err := player.PlaySong(song.ID, startPaused)
		if err != nil {
			return errMsg{err}
		}

		nextIndex := -1
		if m.loopMode == LoopOne {
			nextIndex = index
		} else if index+1 < len(m.queue) {
			nextIndex = index + 1
		} else if m.loopMode == LoopAll && len(m.queue) > 0 {
			nextIndex = 0
		}

		if nextIndex != -1 {
			_ = player.EnqueueSong(m.queue[nextIndex].ID)
		}

		return nil
	}

	return tea.Batch(
		playCmd,
		m.savePlayQueue(),
	)
}

func (m *model) playNext() tea.Cmd {
	if len(m.queue) == 0 {
		return nil
	}

	newIndex := m.queueIndex + 1

	if newIndex >= len(m.queue) {
		return nil
	}

	return m.playQueueIndex(newIndex, false)
}

func (m *model) playPrev() tea.Cmd {
	if len(m.queue) == 0 {
		return nil
	}

	newIndex := m.queueIndex - 1
	if newIndex < 0 {
		newIndex = 0
	}

	return m.playQueueIndex(newIndex, false)
}

func (m *model) setQueue(startIndex int) tea.Cmd {
	newQueue := make([]api.Song, len(m.songs))
	copy(newQueue, m.songs)
	m.queue = newQueue
	return m.playQueueIndex(startIndex, false)
}

func (m *model) savePlayQueue() tea.Cmd {
	ids := []string{}
	currentID := ""

	if len(m.queue) != 0 {
		currentID = m.queue[m.queueIndex].ID
		for _, song := range m.queue {
			ids = append(ids, song.ID)
		}
	}

	return savePlayQueueCmd(ids, currentID)
}

func getSelectedSongs(m model) []api.Song {
	if m.focus == focusMain {
		switch m.viewMode {
		case viewList:
			if m.displayMode == displaySongs && len(m.songs) > 0 {
				return []api.Song{m.songs[m.cursorMain]}
			} else if m.displayMode == displayAlbums || m.displayMode == displayArtist {
				songs, err := api.SubsonicGetAlbum(m.albums[m.cursorMain].ID)
				if err != nil {
					return []api.Song{}
				}
				return songs
			}
		case viewQueue:
			if len(m.queue) > 0 {
				return []api.Song{m.queue[m.cursorMain]}
			}
		}
	}

	return []api.Song{}
}

func (m model) syncNextSong() {
	if len(m.queue) == 0 {
		go player.UpdateNextSong("")
		return
	}

	nextIndex := -1
	switch m.loopMode {
	case LoopOne:
		nextIndex = m.queueIndex
	case LoopNone:
		if m.queueIndex == len(m.queue)-1 {
			nextIndex = -1
		} else {
			nextIndex = m.queueIndex + 1
		}
	case LoopAll:
		if m.queueIndex == len(m.queue)-1 {
			nextIndex = 0
		} else {
			nextIndex = m.queueIndex + 1
		}
	}

	if nextIndex != -1 {
		go player.UpdateNextSong(m.queue[nextIndex].ID)
	} else {
		go player.UpdateNextSong("")
	}
}
