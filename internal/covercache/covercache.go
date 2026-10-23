// Package covercache keeps original cover art images around so each artwork is
// downloaded and decoded once, no matter how often it is redrawn on screen.
//
// The originals are what gets cached: raw image files on disk and decoded
// images in memory. Rendering them into terminal art stays the caller's job,
// because the same artwork is drawn at different sizes.
package covercache

import (
	"bytes"
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	_ "golang.org/x/image/webp"
)

const (
	// maxMemoryItems bounds the decoded images kept in memory. A 500x500 RGBA
	// cover takes about 1 MiB, so this keeps the cache within a few dozen MiB.
	maxMemoryItems = 32
	// maxDiskBytes bounds the original artwork files written to disk.
	maxDiskBytes = 256 << 20
	fileExt      = ".img"
)

// Fetcher downloads the original image bytes for an artwork id.
type Fetcher func(id string) ([]byte, error)

type entry struct {
	id  string
	img image.Image
}

// Cache is a two level cache: decoded images in memory, original files on disk.
// It is safe for concurrent use.
type Cache struct {
	dir      string
	fetch    Fetcher
	maxDisk  int64
	maxItems int

	mu    sync.Mutex
	order *list.List
	items map[string]*list.Element
}

// New creates a cache storing artwork in dir. Directories are created on the
// first write; a missing or unwritable directory only disables the disk layer.
func New(dir string, fetch Fetcher, opts ...Option) *Cache {
	c := &Cache{
		dir:      dir,
		fetch:    fetch,
		maxDisk:  maxDiskBytes,
		maxItems: maxMemoryItems,
		order:    list.New(),
		items:    make(map[string]*list.Element),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Option customises a Cache.
type Option func(*Cache)

// WithMaxDiskBytes bounds the original artwork files kept on disk. A value of
// zero or less keeps the cache in memory only.
func WithMaxDiskBytes(bytes int64) Option {
	return func(c *Cache) {
		c.maxDisk = bytes
	}
}

// WithMaxMemoryItems bounds the decoded images kept in memory.
func WithMaxMemoryItems(items int) Option {
	return func(c *Cache) {
		c.maxItems = items
	}
}

// Get returns the artwork for id, in order of preference: memory, disk, network.
func (c *Cache) Get(id string) (image.Image, error) {
	if id == "" {
		return nil, errors.New("covercache: empty artwork id")
	}

	if img, ok := c.fromMemory(id); ok {
		return img, nil
	}

	path := c.path(id)
	if data, err := os.ReadFile(path); err == nil {
		img, decErr := decode(data)
		if decErr == nil {
			c.toMemory(id, img)
			return img, nil
		}

		// Truncated or unsupported file: drop it and fetch a fresh copy.
		_ = os.Remove(path)
	}

	if c.fetch == nil {
		return nil, fmt.Errorf("covercache: no fetcher configured")
	}

	data, err := c.fetch(id)
	if err != nil {
		return nil, err
	}

	img, err := decode(data)
	if err != nil {
		return nil, err
	}

	c.toMemory(id, img)
	c.write(id, data)

	return img, nil
}

func decode(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

// path derives a filesystem safe file name from the artwork id.
func (c *Cache) path(id string) string {
	sum := sha256.Sum256([]byte(id))

	return filepath.Join(c.dir, hex.EncodeToString(sum[:16])+fileExt)
}

func (c *Cache) fromMemory(id string) (image.Image, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[id]
	if !ok {
		return nil, false
	}

	c.order.MoveToFront(el)
	return el.Value.(*entry).img, true
}

func (c *Cache) toMemory(id string, img image.Image) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if el, ok := c.items[id]; ok {
		c.order.MoveToFront(el)
		el.Value.(*entry).img = img
		return
	}

	if c.maxItems <= 0 {
		return
	}

	c.items[id] = c.order.PushFront(&entry{id: id, img: img})

	for c.order.Len() > c.maxItems {
		last := c.order.Back()
		if last == nil {
			return
		}
		c.order.Remove(last)
		delete(c.items, last.Value.(*entry).id)
	}
}

// write stores the original bytes, ignoring failures: the disk layer is an
// optimisation and must never break playback.
func (c *Cache) write(id string, data []byte) {
	if c.maxDisk <= 0 || c.dir == "" {
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
