package ui

import (
	"testing"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
)

func queueModel(loopMode int) model {
	return model{
		queue: []api.Song{
			{ID: "A", Title: "Song A"},
			{ID: "B", Title: "Song B"},
			{ID: "C", Title: "Song C"},
		},
		queueIndex: 0,
		loopMode:   loopMode,
	}
}

func TestNextSong(t *testing.T) {
	tests := []struct {
		name   string
		model  model
		wantID string
		wantOk bool
	}{
		{"middle of the queue", queueModel(LoopNone), "B", true},
		{"last with loop all", func() model { m := queueModel(LoopAll); m.queueIndex = 2; return m }(), "A", true},
		{"last without loop", func() model { m := queueModel(LoopNone); m.queueIndex = 2; return m }(), "", false},
		{"single track", model{queue: []api.Song{{ID: "A"}}, queueIndex: 0}, "", false},
		{"loop one", queueModel(LoopOne), "", false},
		{"empty queue", model{}, "", false},
		{"index out of range", model{queue: []api.Song{{ID: "A"}}, queueIndex: 4}, "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := tc.model.nextSong()
			if ok != tc.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOk)
			}
			if got.ID != tc.wantID {
				t.Fatalf("id = %q, want %q", got.ID, tc.wantID)
			}
		})
	}
}

func TestPrefetchNextTrackCmd(t *testing.T) {
	if cmd := (model{}).prefetchNextTrackCmd(); cmd != nil {
		t.Fatal("empty queue scheduled a prefetch")
	}

	if cmd := queueModel(LoopOne).prefetchNextTrackCmd(); cmd != nil {
		t.Fatal("loop one scheduled a prefetch for the same track")
	}

	// Nothing is prefetched when both caches are disabled, so no request is
	// made just to throw the result away.
	if cmd := queueModel(LoopNone).prefetchNextTrackCmd(); cmd != nil {
		t.Fatal("disabled caches scheduled a prefetch")
	}

	api.AppConfig.Cache.Lyrics.Enabled = true
	t.Cleanup(func() { api.AppConfig.Cache.Lyrics.Enabled = false })

	if cmd := queueModel(LoopNone).prefetchNextTrackCmd(); cmd == nil {
		t.Fatal("enabled lyrics cache did not schedule a prefetch")
	}
}
