package ui

import (
	"fmt"
	"strings"

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
		switch m.loopMode {
		case LoopAll:
			newIndex = 0
		case LoopOne:
			newIndex = m.queueIndex
		default:
			return nil
		}
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
	var newQueue []api.Song
	newStartIndex := 0

	filters := api.AppConfig.Filters

	for i, song := range m.songs {
		if i == startIndex || !isSongExcluded(*m, song, filters) {
			if i == startIndex {
				newStartIndex = len(newQueue)
			}

			newQueue = append(newQueue, song)
		}
	}

	m.queue = newQueue
	return m.playQueueIndex(newStartIndex, false)
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
	if m.focus == focusMain && cursorInBounds(m) {
		switch m.viewMode {
		case viewList:
			switch m.displayMode {
			case displaySongs:
				return []api.Song{m.songs[m.cursorMain]}

			case displayAlbums:
				songs, err := api.SubsonicGetAlbum(m.albums[m.cursorMain].ID)

				if err != nil {
					return []api.Song{}
				}

				songs = applyExclusionFilters(m, songs)

				var filteredSongs []api.Song
				for _, song := range songs {

					if !song.Filtered {
						filteredSongs = append(filteredSongs, song)
					}
				}

				return filteredSongs
			}
		case viewQueue:
			return []api.Song{m.queue[m.cursorMain]}
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

func applyExclusionFilters(m model, songs []api.Song) []api.Song {
	filters := api.AppConfig.Filters
	if len(filters.Titles) == 0 && len(filters.Artists) == 0 && len(filters.AlbumArtists) == 0 &&
		filters.MinDuration == 0 && len(filters.Genres) == 0 && len(filters.Notes) == 0 &&
		len(filters.Paths) == 0 && filters.MaxPlayCount == 0 && !filters.ExcludeFavorites && filters.MaxRating == 0 {
		return songs
	}

	for i := range songs {
		songs[i].Filtered = isSongExcluded(m, songs[i], filters)
	}

	return songs
}

func isSongExcluded(m model, song api.Song, filters api.Filters) bool {
	if isTitleExcluded(song.Title, filters.Titles) {
		return true
	}

	if isArtistExcluded(song.Artist, filters.Artists) {
		return true
	}

	if isAlbumArtistExcluded(song.AlbumArtists, filters.AlbumArtists) {
		return true
	}

	if isDurationExcluded(song.Duration, filters.MinDuration) {
		return true
	}

	if isGenreExcluded(song.Genre, filters.Genres) {
		return true
	}

	if isNoteExcluded(song.Note, filters.Notes) {
		return true
	}

	if isPathExcluded(song.Path, filters.Paths) {
		return true
	}

	if isPlayCountExcluded(song.PlayCount, filters.MaxPlayCount) {
		return true
	}

	if isFavoriteExcluded(song.ID, m.starredMap, filters.ExcludeFavorites) {
		return true
	}

	if isRatingExcluded(song.Rating, filters.MaxRating) {
		return true
	}

	return false
}

// Helper: Check if title is in the filters
func isTitleExcluded(title string, filters []string) bool {
	if len(filters) == 0 || title == "" {
		return false
	}

	for _, f := range filters {
		if strings.Contains(strings.ToLower(title), strings.ToLower(f)) {
			return true
		}
	}

	return false
}

// Helper: Check if artists is in the filters
func isArtistExcluded(artist string, filters []string) bool {
	if len(filters) == 0 || artist == "" {
		return false
	}

	for _, f := range filters {
		if strings.EqualFold(artist, f) {
			return true
		}
	}

	return false
}

// Helper: Check if album artists is in the filters
func isAlbumArtistExcluded(albumArtists []api.Artist, filters []string) bool {
	if len(filters) == 0 || len(albumArtists) == 0 {
		return false
	}

	for _, f := range filters {
		for _, artist := range albumArtists {
			if strings.EqualFold(artist.Name, f) {
				return true
			}
		}
	}

	return false
}

// Helper: Check if duration is below filter
func isDurationExcluded(duration int, minDuration int) bool {
	if minDuration <= 0 {
		return false
	}

	return duration <= minDuration
}

// Helper: Check if genre is in the filters
func isGenreExcluded(genre string, filters []string) bool {
	if len(filters) == 0 || genre == "" {
		return false
	}

	for _, f := range filters {
		if strings.EqualFold(genre, f) {
			return true
		}
	}

	return false
}

func isNoteExcluded(note string, filters []string) bool {
	if len(filters) == 0 || note == "" {
		return false
	}

	for _, f := range filters {
		if strings.Contains(strings.ToLower(note), strings.ToLower(f)) {
			return true
		}
	}

	return false
}

// Helper: Check if path is in the filters
func isPathExcluded(path string, filters []string) bool {
	if len(filters) == 0 || path == "" {
		return false
	}

	for _, f := range filters {
		if strings.Contains(strings.ToLower(path), strings.ToLower(f)) {
			return true
		}
	}

	return false
}

// Helper: Check if playcount is below filters
func isPlayCountExcluded(playCount int, maxPlayCount int) bool {
	if maxPlayCount == 0 {
		return false
	}

	return playCount <= maxPlayCount
}

// Helper: Check if favorites get filtered
func isFavoriteExcluded(songID string, starredMap map[string]bool, excludeFavorites bool) bool {
	if !excludeFavorites {
		return false
	}

	isStarred := starredMap[songID]
	return isStarred
}

// Helper: Check if rating is below filters
func isRatingExcluded(rating int, maxRating int) bool {
	if maxRating <= 0 || maxRating > 5 {
		return false
	}

	if rating > 0 && rating <= maxRating {
		return true
	}

	return false
}
