package ui

import (
	"github.com/MattiaPun/SubTUI/v2/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

// nextSong returns the track that will play after the current one.
func (m model) nextSong() (api.Song, bool) {
	if len(m.queue) == 0 || m.queueIndex < 0 || m.queueIndex >= len(m.queue) {
		return api.Song{}, false
	}

	// Repeating the same track does not need a new prefetch.
	if m.loopMode == LoopOne {
		return api.Song{}, false
	}

	next := m.queueIndex + 1
	if next >= len(m.queue) {
		if m.loopMode != LoopAll {
			return api.Song{}, false
		}
		next = 0
	}

	if next == m.queueIndex {
		return api.Song{}, false
	}

	return m.queue[next], true
}

// prefetchNextTrackCmd warms the artwork and lyrics caches for the track that
// comes next, so switching to it does not wait for the network.
func (m model) prefetchNextTrackCmd() tea.Cmd {
	next, ok := m.nextSong()
	if !ok {
		return nil
	}

	var cmds []tea.Cmd

	// Without the cover art cache a prefetched image would be downloaded and
	// thrown away immediately.
	if api.AppConfig.Theme.DisplayAlbumArt && api.AppConfig.Cache.CoverArt.Enabled {
		cmds = append(cmds, prefetchCoverArtCmd(next.CoverArtID()))
	}

	if api.AppConfig.Cache.Lyrics.Enabled {
		cmds = append(cmds, prefetchLyricsCmd(next.ID))
	}

	return tea.Batch(cmds...)
}
