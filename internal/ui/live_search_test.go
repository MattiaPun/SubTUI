package ui

import (
	"testing"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/charmbracelet/bubbles/textinput"
)

func newSearchModel(query string) model {
	ti := textinput.New()
	ti.Focus()
	ti.SetValue(query)

	return model{
		textInput:  ti,
		focus:      focusSearch,
		filterMode: filterSongs,
	}
}

func TestLiveSearchIgnoresShortQueries(t *testing.T) {
	m := newSearchModel("a")

	cmd := m.scheduleLiveSearch("")
	if cmd != nil {
		t.Fatal("a single character scheduled a search")
	}
	if m.liveSearchSeq != 1 {
		t.Fatalf("sequence = %d, want 1", m.liveSearchSeq)
	}

	m.textInput.SetValue("ab")
	if cmd := m.scheduleLiveSearch("a"); cmd == nil {
		t.Fatal("two characters did not schedule a search")
	}
}

func TestLiveSearchIgnoresUnchangedQuery(t *testing.T) {
	m := newSearchModel("abc")

	if cmd := m.scheduleLiveSearch("abc"); cmd != nil {
		t.Fatal("unchanged query scheduled a search")
	}
	if m.liveSearchSeq != 0 {
		t.Fatalf("sequence = %d, want 0", m.liveSearchSeq)
	}
}

func TestLiveSearchTickDropsStaleQueries(t *testing.T) {
	m := newSearchModel("abc")
	m.liveSearchSeq = 2

	// An older keystroke fires.
	next, cmd := m.handleLiveSearchTick(liveSearchTickMsg{query: "ab", seq: 1})
	if cmd != nil {
		t.Fatal("stale sequence number ran a search")
	}
	if updated := next.(model); updated.loading {
		t.Fatal("stale tick set the loading state")
	}

	// The timer for the current text but an older keystroke fires after the
	// query was edited again.
	m.textInput.SetValue("abcd")
	_, cmd = m.handleLiveSearchTick(liveSearchTickMsg{query: "abc", seq: 2})
	if cmd != nil {
		t.Fatal("outdated query ran a search")
	}
}

func TestLiveSearchTickKeepsFocusInSearch(t *testing.T) {
	m := newSearchModel("abc")
	m.liveSearchSeq = 1
	m.cursorMain = 7
	m.pageOffset = 3

	next, cmd := m.handleLiveSearchTick(liveSearchTickMsg{query: "abc", seq: 1})
	if cmd == nil {
		t.Fatal("valid tick did not run a search")
	}

	updated := next.(model)
	if updated.focus != focusSearch {
		t.Fatalf("focus = %d, want focusSearch", updated.focus)
	}
	if updated.pageOffset != 0 {
		t.Fatalf("page offset = %d, want 0", updated.pageOffset)
	}
	if !updated.pageHasMore {
		t.Fatal("pageHasMore was not reset")
	}
	if updated.lastSearchQuery != "abc" {
		t.Fatalf("last search query = %q, want abc", updated.lastSearchQuery)
	}
	if !updated.loading {
		t.Fatal("live search did not set the loading state")
	}
}

func TestLiveSearchTickSelectsDisplayMode(t *testing.T) {
	m := newSearchModel("abc")
	m.liveSearchSeq = 1
	m.filterMode = filterAlbums

	next, _ := m.handleLiveSearchTick(liveSearchTickMsg{query: "abc", seq: 1})
	if got := next.(model).displayMode; got != displayAlbums {
		t.Fatalf("display mode = %d, want displayAlbums", got)
	}
}

func TestSearchResultsKeepFocusWhileTyping(t *testing.T) {
	InitStyles()

	// Live search: the user is still in the search field.
	m := newSearchModel("abc")
	m.pageOffset = 0

	next, _ := m.handleSongResult(songsResultMsg{songs: []api.Song{{ID: "1", Title: "Song"}}})
	if got := next.(model).focus; got != focusSearch {
		t.Fatalf("live search result moved focus to %d", got)
	}

	// Explicit search: the focus was moved to the list before the request.
	m = newSearchModel("abc")
	m.focus = focusMain

	next, _ = m.handleSongResult(songsResultMsg{songs: []api.Song{{ID: "1", Title: "Song"}}})
	if got := next.(model).focus; got != focusMain {
		t.Fatalf("explicit search result moved focus to %d", got)
	}
}
