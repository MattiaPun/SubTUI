package ui

import (
	"strings"
	"testing"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/MattiaPun/SubTUI/v2/internal/player"
	"github.com/charmbracelet/x/ansi"
)

func lyricsWith(value string) []api.StructuredLyrics {
	return []api.StructuredLyrics{{
		Synced: false,
		Lines:  []api.LyricLine{{Value: value}},
	}}
}

func playingModel() model {
	return model{
		queue: []api.Song{
			{ID: "A", Title: "Song A"},
			{ID: "B", Title: "Song B"},
		},
		queueIndex:         0,
		lastPlayedSongPath: "stream?id=A",
		playerStatus:       player.PlayerStatus{Path: "stream?id=A"},
		lyricsSongID:       "A",
		songLyrics:         lyricsWith("old lyrics"),
	}
}

func TestSongChangeResetsLyricsAndFetches(t *testing.T) {
	m := playingModel()

	next, cmd := m.handleStatus(statusMsg(player.PlayerStatus{Path: "stream?id=B"}))
	updated, ok := next.(model)
	if !ok {
		t.Fatalf("unexpected model type %T", next)
	}

	if cmd == nil {
		t.Fatal("song change did not schedule a lyrics fetch")
	}
	if updated.queueIndex != 1 {
		t.Fatalf("queue index = %d, want 1", updated.queueIndex)
	}
	if updated.lyricsSongID != "B" {
		t.Fatalf("lyrics song ID = %q, want B", updated.lyricsSongID)
	}
	if !updated.lyricsLoading {
		t.Fatal("lyrics are not marked as loading")
	}
	if updated.songLyrics != nil {
		t.Fatalf("lyrics of the previous song were kept: %v", updated.songLyrics)
	}

	// A stale response for the previous song must not stop the loading state.
	next, _ = updated.handleLyrics(getLyricsMsg{id: "A", result: lyricsWith("stale")})
	updated = next.(model)
	if !updated.lyricsLoading || updated.songLyrics != nil {
		t.Fatal("stale lyrics response was applied")
	}

	// The response for the current song is applied and ends the loading state.
	next, _ = updated.handleLyrics(getLyricsMsg{id: "B", result: lyricsWith("new lyrics")})
	updated = next.(model)
	if updated.lyricsLoading {
		t.Fatal("lyrics are still marked as loading")
	}
	if len(updated.songLyrics) != 1 || updated.songLyrics[0].Lines[0].Value != "new lyrics" {
		t.Fatalf("unexpected lyrics: %v", updated.songLyrics)
	}
}

func TestPlayQueueIndexResetsLyrics(t *testing.T) {
	m := playingModel()

	cmd := m.playQueueIndex(1, false)
	if cmd == nil {
		t.Fatal("playQueueIndex returned no command")
	}
	if m.queueIndex != 1 {
		t.Fatalf("queue index = %d, want 1", m.queueIndex)
	}
	if m.lyricsSongID != "B" || !m.lyricsLoading || m.songLyrics != nil {
		t.Fatalf("lyrics were not reset for the new song: id=%q loading=%v lyrics=%v",
			m.lyricsSongID, m.lyricsLoading, m.songLyrics)
	}

	// Restarting the same song keeps its lyrics.
	m.songLyrics = lyricsWith("B lyrics")
	m.lyricsLoading = false
	m.playQueueIndex(1, false)
	if m.lyricsLoading || len(m.songLyrics) != 1 {
		t.Fatal("restarting the current song dropped its lyrics")
	}
}

func TestQueueEndClearsLyrics(t *testing.T) {
	m := playingModel()
	m.playerStatus = player.PlayerStatus{Path: ""}

	next, _ := m.handleStatus(statusMsg(player.PlayerStatus{}))
	updated := next.(model)

	if updated.songLyrics != nil || updated.lyricsLoading || updated.lyricsSongID != "" {
		t.Fatalf("lyrics state was not cleared: %+v", updated)
	}
}

func TestLyricsViewStatus(t *testing.T) {
	InitStyles()

	loading := model{lyricsLoading: true}
	if out := ansi.Strip(mediaPlayerLyricsContent(loading, 80, 20)); !strings.Contains(out, "Fetching lyrics...") {
		t.Fatalf("loading state not rendered:\n%s", out)
	}

	empty := model{}
	if out := ansi.Strip(mediaPlayerLyricsContent(empty, 80, 20)); !strings.Contains(out, "No lyrics found for this song") {
		t.Fatalf("empty state not rendered:\n%s", out)
	}

	found := model{songLyrics: lyricsWith("a line of lyrics")}
	out := ansi.Strip(mediaPlayerLyricsContent(found, 80, 20))
	if !strings.Contains(out, "a line of lyrics") {
		t.Fatalf("lyrics not rendered:\n%s", out)
	}
	if strings.Contains(out, "Fetching lyrics...") || strings.Contains(out, "No lyrics found") {
		t.Fatalf("status message shown next to fetched lyrics:\n%s", out)
	}
}
