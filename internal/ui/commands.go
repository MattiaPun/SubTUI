package ui

import (
	"time"

	"github.com/MattiaPun/SubTUI/internal/api"
	"github.com/MattiaPun/SubTUI/internal/player"
	tea "github.com/charmbracelet/bubbletea"
)

func attemptLoginCmd() tea.Cmd {
	return func() tea.Msg {
		if err := api.SaveConfig(); err != nil {
			return loginResultMsg{err: err}
		}

		err := api.SubsonicLoginCheck()
		return loginResultMsg{err: err}
	}
}

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

func getAlbumList(searchType string) tea.Cmd {
	return func() tea.Msg {
		albums, err := api.SubsonicGetAlbumList(searchType)
		if err != nil {
			return errMsg{err}
		}
		return albumsResultMsg{albums}
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

func openLikedSongsCmd() tea.Cmd {
	return func() tea.Msg {
		result, err := api.SubsonicGetStarred()
		if err != nil {
			return errMsg{err}
		}

		return viewLikedSongsMsg(result)
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

func getPlayQueue() tea.Cmd {
	return func() tea.Msg {
		result, err := api.SubsonicGetQueue()
		if err != nil {
			return errMsg{err}
		}
		return playQueueResultMsg{result}
	}

}

func savePlayQueueCmd(ids []string, currentID string) tea.Cmd {
	return func() tea.Msg {

		if len(ids) != 0 {
			api.SubsonicSaveQueue(ids, currentID)
		}

		return nil
	}

}
