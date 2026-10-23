// Package lyricscache keeps fetched lyrics around so a song does not have to
// ask the server again every time it is played.
//
// Lyrics are stored as JSON files in a directory of the user's cache and as
// decoded values in memory. Songs without lyrics are cached too, but only for
// a day, so lyrics added to the server later are picked up.
package lyricscache

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
)

const (
	// maxMemoryItems bounds the fetched lyrics kept in memory.
	maxMemoryItems = 32
	// maxDiskBytes bounds the lyrics files written to disk.
	maxDiskBytes = 32 << 20
	// negativeTTL is how long "this song has no lyrics" is trusted.
	negativeTTL = 24 * time.Hour
	fileExt     = ".json"
)

// Fetcher downloads the lyrics of a song.
type Fetcher func(id string) ([]api.StructuredLyrics, error)

// Config customises a Cache. Zero values fall back to the package defaults.
type Config struct {
	Dir            string
	MaxDiskBytes   int64
	MaxMemoryItems int
}

// entry is the on-disk and in-memory representation of cached lyrics.
type entry struct {
	Lyrics    []api.StructuredLyrics `json:"lyrics"`
	FetchedAt time.Time              `json:"fetched_at"`
}

// usable reports whether a cached entry may still be served. Lyrics found once
// stay valid; an empty answer expires so later additions show up.
func (e entry) usable(now time.Time, negativeTTL time.Duration) bool {
	if len(e.Lyrics) > 0 {
		return true
	}

	return now.Sub(e.FetchedAt) < negativeTTL
}

type memoryEntry struct {
	id  string
	ent entry
}

// Cache is a two level lyrics cache: values in memory, JSON files on disk. It
// is safe for concurrent use.
type Cache struct {
	dir         string
	fetch       Fetcher
	maxDisk     int64
	maxItems    int
	negativeTTL time.Duration

	mu    sync.Mutex
	order *list.List
	items map[string]*list.Element
}

// New creates a cache storing lyrics in cfg.Dir. Directories are created on the
// first write; a missing or unwritable directory only disables the disk layer.
func New(cfg Config, fetch Fetcher) *Cache {
	if cfg.MaxDiskBytes == 0 {
		cfg.MaxDiskBytes = maxDiskBytes
	}
	if cfg.MaxMemoryItems == 0 {
		cfg.MaxMemoryItems = maxMemoryItems
	}

	return &Cache{
		dir:         cfg.Dir,
		fetch:       fetch,
		maxDisk:     cfg.MaxDiskBytes,
		maxItems:    cfg.MaxMemoryItems,
		negativeTTL: negativeTTL,
		order:       list.New(),
		items:       make(map[string]*list.Element),
	}
}

// Get returns the lyrics for id, in order of preference: memory, disk, network.
func (c *Cache) Get(id string) ([]api.StructuredLyrics, error) {
	if id == "" {
		return nil, errors.New("lyricscache: empty song id")
	}

	now := time.Now()

	if lyrics, ok := c.fromMemory(id, now); ok {
		return lyrics, nil
	}

	path := c.path(id)
	if data, err := os.ReadFile(path); err == nil {
		var cached entry
		if err := json.Unmarshal(data, &cached); err == nil {
			if cached.usable(now, c.negativeTTL) {
				c.toMemory(id, cached)
				return cached.Lyrics, nil
			}
		}

		// Corrupt or expired: drop it and ask the server again.
		_ = os.Remove(path)
	}

	if c.fetch == nil {
		return nil, fmt.Errorf("lyricscache: no fetcher configured")
	}

	lyrics, err := c.fetch(id)
	if err != nil {
		return nil, err
	}

	fresh := entry{Lyrics: lyrics, FetchedAt: now}
	c.toMemory(id, fresh)
	c.write(id, fresh)

	return lyrics, nil
}

// path derives a filesystem safe file name from the song id.
func (c *Cache) path(id string) string {
	sum := sha256.Sum256([]byte(id))

	return filepath.Join(c.dir, hex.EncodeToString(sum[:16])+fileExt)
}

func (c *Cache) fromMemory(id string, now time.Time) ([]api.StructuredLyrics, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[id]
	if !ok {
		return nil, false
	}

	cached := el.Value.(*memoryEntry)
	if !cached.ent.usable(now, c.negativeTTL) {
		c.order.Remove(el)
		delete(c.items, id)
		return nil, false
	}

	c.order.MoveToFront(el)
	return cached.ent.Lyrics, true
}

func (c *Cache) toMemory(id string, cached entry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[id]; ok {
		c.order.MoveToFront(el)
		el.Value.(*memoryEntry).ent = cached
		return
	}

	if c.maxItems <= 0 {
		return
	}

	c.items[id] = c.order.PushFront(&memoryEntry{id: id, ent: cached})

	for c.order.Len() > c.maxItems {
		last := c.order.Back()
		if last == nil {
			return
		}
		c.order.Remove(last)
		delete(c.items, last.Value.(*memoryEntry).id)
	}
}

// write stores the entry, ignoring failures: the disk layer is an optimisation
// and must never break playback.
func (c *Cache) write(id string, cached entry) {
	if c.maxDisk <= 0 || c.dir == "" {
		return
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return
	}

	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return
	}

	tmp, err := os.CreateTemp(c.dir, "*.tmp")
	if err != nil {
		return
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return
	}
	if err := os.Rename(tmpName, c.path(id)); err != nil {
		_ = os.Remove(tmpName)
		return
	}

	c.trimDisk()
}

// trimDisk drops the least recently written files until the cache fits its limit.
func (c *Cache) trimDisk() {
	files, err := os.ReadDir(c.dir)
	if err != nil {
		return
	}

	type diskFile struct {
		path string
		size int64
		mod  time.Time
	}

	entries := make([]diskFile, 0, len(files))
	var total int64

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		info, err := f.Info()
		if err != nil {
			continue
		}

		entries = append(entries, diskFile{
			path: filepath.Join(c.dir, f.Name()),
			size: info.Size(),
			mod:  info.ModTime(),
		})
		total += info.Size()
	}

	if total <= c.maxDisk {
		return
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].mod.Before(entries[j].mod) })

	for _, e := range entries {
		if total <= c.maxDisk {
			return
		}
		if err := os.Remove(e.path); err == nil {
			total -= e.size
		}
	}
}
