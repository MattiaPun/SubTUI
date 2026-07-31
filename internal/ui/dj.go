package ui

import (
	"math/rand"
	"strings"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

var moodGenres = map[int][]string{
	DjMoodChill:  {"ambient", "chillout", "downtempo", "lounge", "jazz", "classical", "acoustic", "folk", "lo-fi", "soul", "r&b"},
	DjMoodEnergy: {"rock", "metal", "punk", "electronic", "dance", "hip-hop", "pop", "alternative", "indie", "techno", "house"},
	DjMoodFocus:  {"classical", "ambient", "instrumental", "jazz", "post-rock", "minimal", "cinematic", "soundtrack"},
}

type djRefillResultMsg struct {
	songs []api.Song
}

func (m model) djToggle() model {
	m.djEnabled = !m.djEnabled
	if m.djEnabled {
		m.djMood = DjMoodAny
		if len(m.queue) > 0 && m.queueIndex < len(m.queue) {
			m.djSeed = m.queue[m.queueIndex].ID
		}
	} else {
		m.djHistory = make(map[string]bool)
	}
	return m
}

func (m model) djCycleMood() model {
	if !m.djEnabled {
		return m
	}
	m.djMood = (m.djMood + 1) % len(DjMoodLabels)
	return m
}

func (m model) djRefillIfNeeded() tea.Cmd {
	if !m.djEnabled {
		return nil
	}
	remaining := len(m.queue) - m.queueIndex - 1
	if remaining > 5 {
		return nil
	}
	return m.djRefillCmd()
}

func (m model) djRefillCmd() tea.Cmd {
	seedID := m.djSeed
	if seedID == "" && len(m.queue) > 0 && m.queueIndex < len(m.queue) {
		seedID = m.queue[m.queueIndex].ID
	}

	return func() tea.Msg {
		if seedID == "" {
			songs, err := api.SubsonicGetRandomSongs(20, "")
			if err != nil {
				return djRefillResultMsg{}
			}
			return djRefillResultMsg{songs: songs}
		}

		similar, err := api.SubsonicGetSimilarSongs(seedID, 50)
		if err != nil {
			similar = nil
		}

		if len(similar) < 15 {
			random, err := api.SubsonicGetRandomSongs(30, "")
			if err == nil {
				seen := make(map[string]bool, len(similar))
				for _, s := range similar {
					seen[s.ID] = true
				}
				for _, s := range random {
					if !seen[s.ID] {
						similar = append(similar, s)
						seen[s.ID] = true
					}
				}
			}
		}

		return djRefillResultMsg{songs: similar}
	}
}

func (m model) handleDjRefill(msg djRefillResultMsg) (tea.Model, tea.Cmd) {
	if len(msg.songs) == 0 {
		return m, nil
	}

	type weighted struct {
		song   api.Song
		weight float64
	}
	var candidates []weighted

	for _, song := range msg.songs {
		w := 1.0

		if m.djSeed != "" && song.ArtistID != "" {
			var seedArtistID string
			if len(m.queue) > 0 && m.queueIndex < len(m.queue) {
				seedArtistID = m.queue[m.queueIndex].ArtistID
				if song.ArtistID == seedArtistID {
					w *= 3.0
				}
			}
		}

		w *= 1.0 + float64(song.Rating)*0.5

		if song.Filtered || song.ID == "" {
			w = 0
		}

		if m.djHistory[song.ID] {
			w *= 0.05
		}

		if m.djMood != DjMoodAny {
			genre := strings.ToLower(song.Genre)
			validGenres := moodGenres[m.djMood]
			match := false
			for _, g := range validGenres {
				if strings.Contains(genre, g) {
					match = true
					break
				}
			}
			if !match {
				w *= 0.3
			}
		}

		if w > 0 {
			candidates = append(candidates, weighted{song, w})
		}
	}

	if len(candidates) == 0 {
		return m, nil
	}

	totalWeight := 0.0
	for _, c := range candidates {
		totalWeight += c.weight
	}

	count := 15
	if count > len(candidates) {
		count = len(candidates)
	}
	selected := make([]api.Song, 0, count)
	selectedIDs := make(map[string]bool)

	for len(selected) < count {
		pick := rand.Float64() * totalWeight
		cumulative := 0.0
		found := -1
		for i, c := range candidates {
			if selectedIDs[c.song.ID] {
				continue
			}
			cumulative += c.weight
			if pick <= cumulative {
				found = i
				break
			}
		}
		if found == -1 {
			break
		}
		s := candidates[found].song
		selected = append(selected, s)
		selectedIDs[s.ID] = true
		m.djHistory[s.ID] = true
		if len(selectedIDs) >= len(candidates) {
			break
		}
	}

	m.queue = append(m.queue, selected...)

	// DJ-дозаповнення має потрапити в MPRIS TrackList: перегенеровуємо
	// id (ReplaceTrackList згенерує TrackListReplaced) і пере-пушимо
	// metadata поточного треку, бо його старий id став невалідним.
	m = m.generateTrackIDs()
	m.syncTrackListToDBus()
	m.pushCurrentMetadata()

	m.syncNextSong()

	return m, nil
}
