package ui

import (
	"github.com/MattiaPun/SubTUI/internal/api"
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
			albums, err := api.SubsonicSearchAlbum(query, 0)
			if err != nil {
				return errMsg{err}
			}
			return albumsResultMsg{albums}

		case filterArtist:
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

		return viewStarredSongsMsg(result)
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

func addSongToPlaylistCmd(songID string, playlistID string) tea.Cmd {
	return func() tea.Msg {

		if songID != "" && playlistID != "" {
			api.SubsonicAddToPlaylist(songID, playlistID)
		}

		return nil
	}
}

func addRatingCmd(ID string, rating int) tea.Cmd {
	return func() tea.Msg {

		if ID != "" && rating >= 0 && rating <= 5 {
			api.SubsonicRate(ID, rating)
		}

		return nil
	}
}

func createMediaShareCmd(ID string) tea.Cmd {
	return func() tea.Msg {

		if ID != "" {
			url, err := api.SubsonicCreateShare(ID)
			if err != nil {
				return errMsg{err}
			}

			return createShareMsg{url: url}
		}
		return nil
	}
}
