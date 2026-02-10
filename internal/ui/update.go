package ui

import (
	"github.com/MattiaPun/SubTUI/v2/internal/integration"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		return m.handleWindowResize(msg)

	case tea.KeyMsg:
		return m.handlesKeys(msg)

	case loginResultMsg:
		return m.handleLoginResult(msg)

	case playlistResultMsg:
		return m.handlePlaylistResult(msg)

	case errMsg:
		return m.handleErr(msg)

	case statusMsg:
		return m.handleStatus(msg)

	case songsResultMsg:
		return m.handleSongResult(msg)

	case albumsResultMsg:
		return m.handleAlbumResult(msg)

	case artistsResultMsg:
		return m.handleArtistsResult(msg)

	case starredResultMsg:
		return m.handleStarredResult(msg)

	case viewStarredSongsMsg:
		return m.handleViewStarredSongs(msg)

	case createShareMsg:
		return m.handleCreateShare(msg)

	case playQueueResultMsg:
		return m.handlePlayQueueResult(msg)

	case SetDBusMsg:
		return m.handleSetDBUS(msg)

	case integration.PlayPauseMsg:
		return m.handleIntegrationPlayPause(msg)

	case integration.NextSongMsg:
		return m.handleIntegrationNextSong(msg)

	case integration.PreviousSongMsg:
		return m.handleIntegrationPreviousSong(msg)

	case SetDiscordMsg:
		return m.handleSetDiscord(msg)

	}

	return m, cmd
}
