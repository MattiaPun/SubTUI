package ui

import (
	"fmt"

	"git.punjwani.pm/Mattia/DepthTUI/internal/api"
	"git.punjwani.pm/Mattia/DepthTUI/internal/player"
	tea "github.com/charmbracelet/bubbletea"
)

func formatDuration(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

func (m *model) playQueueIndex(index int) tea.Cmd {
	if index < 0 || index >= len(m.queue) {
		return nil
	}

	m.queueIndex = index
	song := m.queue[m.queueIndex]

	return func() tea.Msg {
		err := player.PlaySong(song.ID)
		if err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func (m *model) playNext() tea.Cmd {
	if len(m.queue) == 0 {
		return nil
	}

	newIndex := m.queueIndex + 1

	if newIndex >= len(m.queue) {
		return nil
	}

	return m.playQueueIndex(newIndex)
}

func (m *model) playPrev() tea.Cmd {
	if len(m.queue) == 0 {
		return nil
	}

	newIndex := m.queueIndex - 1
	if newIndex < 0 {
		newIndex = 0
	}

	return m.playQueueIndex(newIndex)
}

func (m *model) setQueue(songs []api.Song, startIndex int) tea.Cmd {
	m.queue = songs
	return m.playQueueIndex(startIndex)
}
