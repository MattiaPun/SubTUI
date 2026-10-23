package ui

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/MattiaPun/SubTUI/v2/internal/covercache"
	"github.com/MattiaPun/SubTUI/v2/internal/lyricscache"
	tea "github.com/charmbracelet/bubbletea"
	_ "golang.org/x/image/webp"
)

func attemptLoginCmd() tea.Cmd {
	return func() tea.Msg {
		if err := api.SaveConfig(filepath.Join(api.ConfigDir, "credentials.toml"), api.AppServerConfig, 0600); err != nil {
			return loginResultMsg{err: err}
		}

		err := api.SubsonicLoginCheck()
		return loginResultMsg{err: err}
	}
}

func searchCmd(query string, mode int, offset int) tea.Cmd {
	return func() tea.Msg {

		switch mode {
		case filterSongs:
			songs, err := api.SubsonicSearchSong(query, offset)
			if err != nil {
				return errMsg{err}
			}
			return songsResultMsg{songs}

		case filterAlbums:
			albums, err := api.SubsonicSearchAlbum(query, offset)
			if err != nil {
				return errMsg{err}
			}
			return albumsResultMsg{albums}

		case filterArtist:
			artists, err := api.SubsonicSearchArtist(query, offset)
			if err != nil {
				return errMsg{err}
			}
			return artistsResultMsg{artists}
		}

		return nil
	}
}

func getAlbumSongs(albumID string, shuffled bool) tea.Cmd {
	return func() tea.Msg {
		songs, err := api.SubsonicGetAlbum(albumID)
		if err != nil {
			return errMsg{err}
		}

		if shuffled {
			return shuffledSongsMsg{songs, false}
		} else {
			return songsResultMsg{songs}
		}
	}
}

func getAlbumList(searchType string, offset int) tea.Cmd {
	return func() tea.Msg {
		albums, err := api.SubsonicGetAlbumList(searchType, offset)
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

func getPlaylistSongs(id string, shuffled bool) tea.Cmd {
	return func() tea.Msg {
		songs, err := api.SubsonicGetPlaylistSongs(id)
		if err != nil {
			return errMsg{err}
		}

		if shuffled {
			return shuffledSongsMsg{songs, true}
		} else {
			return songsResultMsg{songs}
		}
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

func toggleStarCmd(idsToStar []string, idsToUnstar []string) tea.Cmd {
	return func() tea.Msg {
		go api.SubsonicStar(idsToStar)
		go api.SubsonicUnstar(idsToUnstar)
		return nil
	}
}

// coverArtSize is the artwork resolution requested from the server. It is
// rendered at varying terminal sizes, so it is fetched large once and cached.
const coverArtSize = 500

var (
	coverCacheOnce sync.Once
	coverCache     *covercache.Cache
)

// getCoverArtCache lazily creates the shared artwork cache from the settings in
// the [cache.cover_art] section.
func getCoverArtCache() *covercache.Cache {
	coverCacheOnce.Do(func() {
		settings := api.AppConfig.Cache.CoverArt

		dir := settings.Dir
		if dir == "" {
			userCacheDir, err := os.UserCacheDir()
			if err != nil {
				userCacheDir = os.TempDir()
			}
			dir = filepath.Join(userCacheDir, "subtui", "covers")
		}

		coverCache = covercache.New(
			filepath.Join(dir),
			func(id string) ([]byte, error) {
				return api.SubsonicCoverArt(id, coverArtSize)
			},
			covercache.WithMaxDiskBytes(int64(settings.MaxSizeMB)<<20),
		)
	})

	return coverCache
}

func getCoverArtCmd(coverArtID string) tea.Cmd {
	return func() tea.Msg {
		if !api.AppConfig.Cache.CoverArt.Enabled {
			return fetchCoverArt(coverArtID)
		}

		img, err := getCoverArtCache().Get(coverArtID)
		if err != nil {
			return nil
		}

		return coverArtMsg{
			img: img,
		}
	}
}

// fetchCoverArt downloads and decodes artwork without touching the cache.
func fetchCoverArt(coverArtID string) tea.Msg {
	imgData, err := api.SubsonicCoverArt(coverArtID, coverArtSize)
	if err != nil {
		return nil
	}

	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil
	}

	return coverArtMsg{img: img}
}

var (
	lyricsCacheOnce sync.Once
	lyricsCache     *lyricscache.Cache
)

// getLyricsCache lazily creates the shared lyrics cache from the settings in
// the [cache.lyrics] section.
func getLyricsCache() *lyricscache.Cache {
	lyricsCacheOnce.Do(func() {
		settings := api.AppConfig.Cache.Lyrics

		dir := settings.Dir
		if dir == "" {
			userCacheDir, err := os.UserCacheDir()
			if err != nil {
				userCacheDir = os.TempDir()
			}
			dir = filepath.Join(userCacheDir, "subtui", "lyrics")
		}

		lyricsCache = lyricscache.New(
			lyricscache.Config{
				Dir:          dir,
				MaxDiskBytes: int64(settings.MaxSizeMB) << 20,
			},
			api.SubsonicGetLyrics,
		)
	})

	return lyricsCache
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

func addSongToPlaylistCmd(playlistID string, songIds []string) tea.Cmd {
	return func() tea.Msg {

		if playlistID == "" || len(songIds) == 0 {
			return nil
		}

		api.SubsonicAddToPlaylist(playlistID, songIds)
		return nil
	}
}

func addRatingCmd(ids []string, rating int) tea.Cmd {
	return func() tea.Msg {
		if rating >= 00 && rating <= 5 && len(ids) > 0 {
			for i := range ids {
				go api.SubsonicRate(ids[i], rating)
			}
		}

		return nil
	}
}

func createMediaShareCmd(ids []string) tea.Cmd {
	return func() tea.Msg {

		if len(ids) > 0 {
			url, err := api.SubsonicCreateShare(ids)
			if err != nil {
				return errMsg{err}
			}

			return createShareMsg{url: url}
		}
		return nil
	}
}

func getLyricsCmd(ID string) tea.Cmd {
	return func() tea.Msg {
		if ID == "" {
			return nil
		}

		if !api.AppConfig.Cache.Lyrics.Enabled {
			return fetchLyrics(ID)
		}

		result, err := getLyricsCache().Get(ID)
		if err != nil {
			return errMsg{err}
		}

		return getLyricsMsg{result}
	}
}

// fetchLyrics requests lyrics without touching the cache.
func fetchLyrics(ID string) tea.Msg {
	result, err := api.SubsonicGetLyrics(ID)
	if err != nil {
		return errMsg{err}
	}

	return getLyricsMsg{result}
}

// prefetchLyricsCmd fetches lyrics ahead of time for the track that is queued
// next and stores them in the cache. Nothing is displayed.
func prefetchLyricsCmd(ID string) tea.Cmd {
	return func() tea.Msg {
		// Without the cache the lyrics would be downloaded and thrown away.
		if ID == "" || !api.AppConfig.Cache.Lyrics.Enabled {
			return nil
		}

		_, _ = getLyricsCache().Get(ID)

		return nil
	}
}

// prefetchCoverArtCmd downloads the artwork of the next track into the cache
// without touching what is on screen.
func prefetchCoverArtCmd(coverArtID string) tea.Cmd {
	return func() tea.Msg {
		if coverArtID == "" {
			return nil
		}

		_, _ = getCoverArtCache().Get(coverArtID)

		return nil
	}
}
