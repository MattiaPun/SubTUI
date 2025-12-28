package ui

import (
	"time"

	"git.punjwani.pm/Mattia/DepthTUI/internal/api"
	"git.punjwani.pm/Mattia/DepthTUI/internal/player"
	tea "github.com/charmbracelet/bubbletea"
)

func searchCmd(query string, mode int) tea.Cmd {
	return func() tea.Msg {

		switch mode {
		case filterSongs:
			songs, err := api.SubsonicSearchSong(query, 0)
			if err != nil {
				return errMsg{err}
			}
			return songsResultMsg{songs}

		case filterAlbums:
			// Ensure api.SubsonicSearchAlbum exists in your api package!
			albums, err := api.SubsonicSearchAlbum(query, 0)
			if err != nil {
				return errMsg{err}
			}
			return albumsResultMsg{albums}

		case filterArtist:
			// Ensure api.SubsonicSearchArtist exists in your api package!
			artists, err := api.SubsonicSearchArtist(query, 0)
			if err != nil {
				return errMsg{err}
			}
			return artistsResultMsg{artists}
		}

		return nil
	}
}

func getAlbumSongs(albumID string) tea.Cmd {
	return func() tea.Msg {
		songs, err := api.SubsonicGetAlbum(albumID)
		if err != nil {
			return errMsg{err}
		}
		return songsResultMsg{songs}
	}
}

func getArtistAlbums(artistID string) tea.Cmd {
	return func() tea.Msg {
		albums, err := api.SubsonicGetArtist(artistID)
		if err != nil {
			return errMsg{err}
		}
		return albumsResultMsg{albums}
	}
}

func getPlaylists() tea.Cmd {
	return func() tea.Msg {
		playlists, err := api.SubsonicGetPlaylists()
		if err != nil {
			return errMsg{err}
		}
		return playlistResultMsg{playlists}
	}
}

func getPlaylistSongs(id string) tea.Cmd {
	return func() tea.Msg {
		songs, err := api.SubsonicGetPlaylistSongs(id)
		if err != nil {
			return errMsg{err}
		}
		return songsResultMsg{songs}
	}
}

func syncPlayerCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return statusMsg(player.GetPlayerStatus())
	})
}

func getStarredCmd() tea.Cmd {
	return func() tea.Msg {
		result, err := api.SubsonicGetStarred()
		if err != nil {
			return errMsg{err}
		}
		return starredResultMsg{result}
	}
}

func toggleStarCmd(id string, isCurrentlyStarred bool) tea.Cmd {
	return func() tea.Msg {
		if isCurrentlyStarred {
			api.SubsonicUnstar(id)
		} else {
			api.SubsonicStar(id)
		}
		return nil
	}
}

func checkLoginCmd() tea.Cmd {
	return func() tea.Msg {
		if err := api.SubsonicPing(); err != nil {
			return errMsg{err}
		}

		return nil
	}
}
