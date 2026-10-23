package lyricscache

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
)

func lyrics(value string) []api.StructuredLyrics {
	return []api.StructuredLyrics{{
		Synced: false,
		Lines:  []api.LyricLine{{Value: value}},
	}}
}

func TestGetFetchesOnceAndRemembers(t *testing.T) {
	fetched := 0
	cache := New(Config{Dir: t.TempDir()}, func(id string) ([]api.StructuredLyrics, error) {
		fetched++
		return lyrics("line for " + id), nil
	})

	first, err := cache.Get("song-1")
	if err != nil {
		t.Fatalf("first Get: %v", err)
	}
	if first[0].Lines[0].Value != "line for song-1" {
		t.Fatalf("unexpected lyrics: %v", first)
	}

	second, err := cache.Get("song-1")
	if err != nil {
		t.Fatalf("second Get: %v", err)
	}
	if fetched != 1 {
		t.Fatalf("fetched %d times, want 1", fetched)
	}
	if len(second) != len(first) {
		t.Fatalf("second Get returned %v, want %v", second, first)
	}
}

func TestGetUsesDiskAfterRestart(t *testing.T) {
	dir := t.TempDir()

	fetched := 0
	fetch := func(string) ([]api.StructuredLyrics, error) {
		fetched++
		return lyrics("cached line"), nil
	}

	if _, err := New(Config{Dir: dir}, fetch).Get("song-1"); err != nil {
		t.Fatalf("warm up: %v", err)
	}

	restarted := New(Config{Dir: dir}, func(string) ([]api.StructuredLyrics, error) {
		t.Fatal("disk cache miss, fetch must not be called")
		return nil, nil
	})

	got, err := restarted.Get("song-1")
	if err != nil {
		t.Fatalf("disk hit: %v", err)
	}
	if got[0].Lines[0].Value != "cached line" {
		t.Fatalf("unexpected lyrics: %v", got)
	}
	if fetched != 1 {
		t.Fatalf("fetched %d times, want 1", fetched)
	}
}

func TestEmptyResultIsCachedButExpires(t *testing.T) {
	dir := t.TempDir()

	fetched := 0
	fetch := func(string) ([]api.StructuredLyrics, error) {
		fetched++
		return nil, nil
	}

	cache := New(Config{Dir: dir}, fetch)
	cache.negativeTTL = 10 * time.Millisecond

	if got, err := cache.Get("song-1"); err != nil || len(got) != 0 {
		t.Fatalf("Get = %v, %v", got, err)
	}
	if _, err := cache.Get("song-1"); err != nil {
		t.Fatalf("second Get: %v", err)
	}
	if fetched != 1 {
		t.Fatalf("empty result was fetched %d times, want 1 within the ttl", fetched)
	}

	time.Sleep(20 * time.Millisecond)

	if _, err := cache.Get("song-1"); err != nil {
		t.Fatalf("Get after ttl: %v", err)
	}
	if fetched != 2 {
		t.Fatalf("fetched %d times, want 2 after the ttl expired", fetched)
	}
}

func TestFoundLyricsDoNotExpire(t *testing.T) {
	cache := New(Config{Dir: t.TempDir()}, func(string) ([]api.StructuredLyrics, error) {
		return lyrics("kept"), nil
	})
	cache.negativeTTL = time.Nanosecond

	if _, err := cache.Get("song-1"); err != nil {
		t.Fatalf("first Get: %v", err)
	}
	time.Sleep(time.Millisecond)

	if _, err := cache.Get("song-1"); err != nil {
		t.Fatalf("second Get: %v", err)
	}
}

func TestGetReplacesCorruptFile(t *testing.T) {
	dir := t.TempDir()
	cache := New(Config{Dir: dir}, func(string) ([]api.StructuredLyrics, error) {
		return lyrics("fresh"), nil
	})

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache.path("song-1"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := cache.Get("song-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got[0].Lines[0].Value != "fresh" {
		t.Fatalf("unexpected lyrics: %v", got)
	}

	data, err := os.ReadFile(cache.path("song-1"))
	if err != nil {
		t.Fatalf("cache file not rewritten: %v", err)
	}
	if len(data) == 0 || data[0] != '{' {
		t.Fatalf("cache file holds %q, want JSON", data)
	}
}

func TestMemoryEviction(t *testing.T) {
	cache := New(Config{Dir: t.TempDir(), MaxMemoryItems: 2}, func(string) ([]api.StructuredLyrics, error) {
		return lyrics("line"), nil
	})

	for _, id := range []string{"a", "b", "c"} {
		if _, err := cache.Get(id); err != nil {
			t.Fatalf("Get(%q): %v", id, err)
		}
	}

	cache.mu.Lock()
	stored := len(cache.items)
	_, hasA := cache.items["a"]
	cache.mu.Unlock()

	if stored != 2 {
		t.Fatalf("memory cache holds %d entries, want 2", stored)
	}
	if hasA {
		t.Fatal("least recently used lyrics were not evicted")
	}
}

func TestDiskEviction(t *testing.T) {
	dir := t.TempDir()

	cache := New(Config{Dir: dir, MaxDiskBytes: 600}, func(string) ([]api.StructuredLyrics, error) {
		return lyrics("a line long enough to take some space on disk"), nil
	})

	for _, id := range []string{"a", "b", "c", "d"} {
		if _, err := cache.Get(id); err != nil {
			t.Fatalf("Get(%q): %v", id, err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	var total int64
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			t.Fatal(err)
		}
		total += info.Size()
	}

	if total > cache.maxDisk {
		t.Fatalf("disk cache uses %d bytes, limit is %d", total, cache.maxDisk)
	}
	if len(entries) == 0 {
		t.Fatal("disk cache evicted everything")
	}
}

func TestGetErrors(t *testing.T) {
	cache := New(Config{Dir: t.TempDir()}, func(string) ([]api.StructuredLyrics, error) {
		return nil, errors.New("network down")
	})

	if _, err := cache.Get(""); err == nil {
		t.Fatal("empty id did not return an error")
	}
	if _, err := cache.Get("song-1"); err == nil {
		t.Fatal("fetch error was swallowed")
	}

	noFetcher := New(Config{Dir: t.TempDir()}, nil)
	if _, err := noFetcher.Get("song-1"); err == nil {
		t.Fatal("missing fetcher did not return an error")
	}
}

func TestCacheFilesDoNotUseSongIDs(t *testing.T) {
	cache := New(Config{Dir: t.TempDir()}, nil)

	for _, id := range []string{"al-1/2", "al-1_2", "../../etc/passwd"} {
		if dir := filepath.Dir(cache.path(id)); dir != cache.dir {
			t.Fatalf("path %q escapes the cache directory", cache.path(id))
		}
	}

	if cache.path("al-1/2") == cache.path("al-1_2") {
		t.Fatal("different song ids map to the same file")
	}
}
