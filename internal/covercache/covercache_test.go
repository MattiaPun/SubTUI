package covercache

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func pngImage(t *testing.T, size int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 128, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	return buf.Bytes()
}

func TestGetFetchesOnceAndKeepsOriginal(t *testing.T) {
	fetched := 0
	imageData := pngImage(t, 500)

	cache := New(t.TempDir(), func(id string) ([]byte, error) {
		fetched++
		return imageData, nil
	})

	first, err := cache.Get("cover-1")
	if err != nil {
		t.Fatalf("first Get: %v", err)
	}
	if got := first.Bounds().Dx(); got != 500 {
		t.Fatalf("cached image width = %d, want the original 500", got)
	}

	second, err := cache.Get("cover-1")
	if err != nil {
		t.Fatalf("second Get: %v", err)
	}
	if fetched != 1 {
		t.Fatalf("fetched %d times, want 1", fetched)
	}
	if second != first {
		t.Fatal("second Get did not reuse the decoded image")
	}
}

func TestGetUsesDiskAfterRestart(t *testing.T) {
	dir := t.TempDir()
	imageData := pngImage(t, 120)

	fetched := 0
	fetch := func(string) ([]byte, error) {
		fetched++
		return imageData, nil
	}

	if _, err := New(dir, fetch).Get("cover-1"); err != nil {
		t.Fatalf("warm up: %v", err)
	}
	if fetched != 1 {
		t.Fatalf("fetched %d times during warm up, want 1", fetched)
	}

	// A fresh cache process should read the original from disk.
	restarted := New(dir, func(string) ([]byte, error) {
		t.Fatal("disk cache miss, fetch must not be called")
		return nil, nil
	})

	img, err := restarted.Get("cover-1")
	if err != nil {
		t.Fatalf("disk hit: %v", err)
	}
	if got := img.Bounds().Dx(); got != 120 {
		t.Fatalf("image width = %d, want 120", got)
	}
	if fetched != 1 {
		t.Fatalf("fetched %d times, want 1", fetched)
	}
}

func TestGetReplacesCorruptFile(t *testing.T) {
	dir := t.TempDir()
	cache := New(dir, func(string) ([]byte, error) {
		return pngImage(t, 60), nil
	})

	// Simulate a truncated download from a previous run.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cache.path("cover-1"), []byte("not an image"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := cache.Get("cover-1"); err != nil {
		t.Fatalf("Get: %v", err)
	}

	data, err := os.ReadFile(cache.path("cover-1"))
	if err != nil {
		t.Fatalf("cache file not rewritten: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("\x89PNG")) {
		t.Fatalf("cache file holds %q, want the refetched PNG", data[:8])
	}
}

func TestMemoryEviction(t *testing.T) {
	cache := New(t.TempDir(), func(string) ([]byte, error) {
		return pngImage(t, 16), nil
	})
	cache.maxItems = 2

	for _, id := range []string{"a", "b", "c"} {
		if _, err := cache.Get(id); err != nil {
			t.Fatalf("Get(%q): %v", id, err)
		}
	}

	cache.mu.Lock()
	got := len(cache.items)
	_, hasA := cache.items["a"]
	cache.mu.Unlock()

	if got != 2 {
		t.Fatalf("memory cache holds %d images, want 2", got)
	}
	if hasA {
		t.Fatal("least recently used image was not evicted")
	}
}

func TestDiskEviction(t *testing.T) {
	dir := t.TempDir()
	imageData := pngImage(t, 200)

	cache := New(dir, func(string) ([]byte, error) {
		return imageData, nil
	})
	// Leave room for roughly one and a half images.
	cache.maxDisk = int64(len(imageData)) + int64(len(imageData))/2

	for _, id := range []string{"a", "b", "c"} {
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
	cache := New(t.TempDir(), func(string) ([]byte, error) {
		return nil, errors.New("network down")
	})

	if _, err := cache.Get(""); err == nil {
		t.Fatal("empty id did not return an error")
	}

	if _, err := cache.Get("cover-1"); err == nil {
		t.Fatal("fetch error was swallowed")
	}

	broken := New(t.TempDir(), func(string) ([]byte, error) {
		return []byte("not an image"), nil
	})
	if _, err := broken.Get("cover-1"); err == nil {
		t.Fatal("undecodable artwork was accepted")
	}
}

func TestCacheFilesAreNotHumanReadableIDs(t *testing.T) {
	// Artwork ids come from the server and may contain path separators or
	// case-insensitive collisions, so they must not be used as file names.
	cache := New(t.TempDir(), nil)

	for _, id := range []string{"al-1/2", "al-1_2", "../../etc/passwd"} {
		path := cache.path(id)
		if filepath.Dir(path) != cache.dir {
			t.Fatalf("path %q escapes the cache directory", path)
		}
	}

	if cache.path("al-1/2") == cache.path("al-1_2") {
		t.Fatal("different artwork ids map to the same file")
	}
}
