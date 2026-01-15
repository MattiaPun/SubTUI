package ui

import (
	"fmt"

	"github.com/MattiaPun/SubTUI/internal/api"
	"github.com/MattiaPun/SubTUI/internal/player"
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
