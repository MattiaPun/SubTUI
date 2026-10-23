package ui

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

func solidImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	return img
}

func TestCoverArtCellSize(t *testing.T) {
	square := solidImage(500, 500)
	wide := solidImage(600, 300)
	tall := solidImage(300, 600)

	tests := []struct {
		name         string
		img          image.Image
		maxW, maxH   int
		wantW, wantH int
	}{
		{"square fits width", square, 10, 20, 10, 5},
		{"square clipped by height", square, 60, 17, 34, 17},
		{"wide fits width", wide, 60, 20, 60, 15},
		{"tall clipped by height", tall, 60, 20, 20, 20},
		{"tiny box", square, 1, 1, 1, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, h := coverArtCellSize(tc.img, tc.maxW, tc.maxH)
			if w != tc.wantW || h != tc.wantH {
				t.Fatalf("got %dx%d, want %dx%d", w, h, tc.wantW, tc.wantH)
			}
		})
	}
}

// Largest contiguous run of full blocks in a line (art is rendered as solid
// white blocks for the test image).
func blockRun(line string) int {
	best, cur := 0, 0
	for _, r := range line {
		if r == '█' {
			cur++
			if cur > best {
				best = cur
			}
		} else {
			cur = 0
		}
	}
	return best
}

func TestCoverArtFillsMediaPlayerPane(t *testing.T) {
	InitStyles()

	m := model{
		width:           160,
		height:          45,
		showMediaPlayer: true,
		coverArt:        solidImage(500, 500),
		queue:           []api.Song{{Title: "Song", Artist: "Artist", Album: "Album"}},
		queueIndex:      0,
	}

	content := ansi.Strip(mediaPlayerContent(m))
	lines := strings.Split(content, "\n")

	runs, rows := 0, 0
	for _, line := range lines {
		if run := blockRun(line); run > 0 {
			rows++
			if run > runs {
				runs = run
			}
		}
	}

	// availableWidth=156 -> side width=62; height=42 -> cover box content 17 rows
	wantW, wantH := coverArtCellSize(m.coverArt, 62, 17)
	if runs != wantW || rows != wantH {
		t.Fatalf("art block is %dx%d, want %dx%d", runs, rows, wantW, wantH)
	}
	if wantW < 30 || wantH < 15 {
		t.Fatalf("art did not grow to fill the pane: %dx%d", wantW, wantH)
	}

	// The whole media player must not overflow the terminal height.
	if got := len(lines); got > m.height {
		t.Fatalf("media player rendered %d lines, terminal has %d", got, m.height)
	}
	for _, line := range lines {
		if w := runewidth.StringWidth(line); w > m.width {
			t.Fatalf("line wider than terminal: %d > %d (%q)", w, m.width, line)
		}
	}
}

func TestFooterCoverArtSize(t *testing.T) {
	InitStyles()
	api.AppConfig.Theme.DisplayAlbumArt = true
	t.Cleanup(func() { api.AppConfig.Theme.DisplayAlbumArt = false })

	m := model{
		width:      160,
		height:     45,
		coverArt:   solidImage(500, 500),
		queue:      []api.Song{{Title: "Song", Artist: "Artist", Album: "Album"}},
		queueIndex: 0,
	}

	content := ansi.Strip(footerContent(m))
	rows, runs := 0, 0
	for _, line := range strings.Split(content, "\n") {
		if run := blockRun(line); run > 0 {
			rows++
			if run > runs {
				runs = run
			}
		}
	}

	if runs != footerCoverArtWidth || rows != footerCoverArtHeight {
		t.Fatalf("footer art block is %dx%d, want %dx%d", runs, rows, footerCoverArtWidth, footerCoverArtHeight)
	}
}
