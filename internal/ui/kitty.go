package ui

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/charmbracelet/x/mosaic"
)

const (
	AlbumArtRendererAuto   = "auto"
	AlbumArtRendererKitty  = "kitty"
	AlbumArtRendererMosaic = "mosaic"
)

// Fixed image ID for the cover art (must fit in 24 bits, it is encoded as a fg color)
const kittyImageID = 0x5AB71

// Terminal output shared by bubbletea and the kitty graphics sequences
var TermOut = &lockedTerm{f: os.Stdout}

var kittyUsed bool

// Helper: Serialize writes to the terminal so escape sequences don't interleave with frames
type lockedTerm struct {
	mu sync.Mutex
	f  *os.File
}

func (t *lockedTerm) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.f.Write(p)
}

func (t *lockedTerm) Read(p []byte) (int, error) { return t.f.Read(p) }
func (t *lockedTerm) Close() error               { return nil }
func (t *lockedTerm) Fd() uintptr                { return t.f.Fd() }

// Helper: Check if the terminal supports kitty unicode placeholders
func kittySupported() bool {
	if os.Getenv("TMUX") != "" || os.Getenv("STY") != "" {
		return false
	}

	term := os.Getenv("TERM")
	return os.Getenv("KITTY_WINDOW_ID") != "" ||
		term == "xterm-kitty" ||
		term == "xterm-ghostty" ||
		strings.EqualFold(os.Getenv("TERM_PROGRAM"), "ghostty")
}

// Helper: Resolve the configured renderer to 'kitty' or 'mosaic'
func resolveAlbumArtRenderer(cfg string) string {
	switch strings.ToLower(strings.TrimSpace(cfg)) {
	case AlbumArtRendererKitty:
		return AlbumArtRendererKitty
	case AlbumArtRendererAuto, "":
		if kittySupported() {
			return AlbumArtRendererKitty
		}
	}

	return AlbumArtRendererMosaic
}

func kittyDeleteSeq() string {
	return ansi.KittyGraphics(nil, "a=d", "d=I", fmt.Sprintf("i=%d", kittyImageID), "q=2")
}

func kittyPlaceSeq(cols, rows int) string {
	id := fmt.Sprintf("i=%d", kittyImageID)
	return ansi.KittyGraphics(nil, "a=d", "d=i", id, "q=2") +
		ansi.KittyGraphics(nil, "a=p", "U=1", id, fmt.Sprintf("c=%d", cols), fmt.Sprintf("r=%d", rows), "q=2")
}

func kittyTransmitSeq(img image.Image) (string, error) {
	var buf bytes.Buffer
	buf.WriteString(kittyDeleteSeq())

	err := kitty.EncodeGraphics(&buf, img, &kitty.Options{
		Action:       kitty.Transmit,
		Transmission: kitty.Direct, // must be explicit, otherwise no image data is sent
		ID:           kittyImageID,
		Format:       kitty.PNG,
		Quite:        2,
		Chunk:        true,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Send the cover art to the terminal
func kittyTransmit(img image.Image) {
	seq, err := kittyTransmitSeq(img)
	if err != nil {
		log.Printf("[KITTY] Failed to encode cover art: %v", err)
		return
	}

	kittyUsed = true
	_, _ = TermOut.Write([]byte(seq))
}

// Create a virtual placement of cols x rows cells for the cover art
func kittyPlace(cols, rows int) {
	_, _ = TermOut.Write([]byte(kittyPlaceSeq(cols, rows)))
}

// Free the cover art from the terminal
func kittyDelete() {
	if kittyUsed {
		_, _ = TermOut.Write([]byte(kittyDeleteSeq()))
	}
}

// Free the cover art on exit
func KittyCleanup() {
	kittyDelete()
}

// Generate the grid of unicode placeholders where the terminal draws the image
func kittyPlaceholder(cols, rows int) string {
	fg := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", (kittyImageID>>16)&0xFF, (kittyImageID>>8)&0xFF, kittyImageID&0xFF)

	lines := make([]string, rows)
	for r := range rows {
		var sb strings.Builder
		sb.WriteString(fg)
		for c := range cols {
			sb.WriteRune(kitty.Placeholder)
			sb.WriteRune(kitty.Diacritic(r))
			sb.WriteRune(kitty.Diacritic(c))
		}
		sb.WriteString("\x1b[39m")
		lines[r] = sb.String()
	}

	return strings.Join(lines, "\n")
}

// Helper: (Re)place the cover art if its size changed
func (m model) syncKittyPlacement() model {
	if m.coverCols == m.kittyPlacedCols && m.coverRows == m.kittyPlacedRows {
		return m
	}

	kittyPlace(m.coverCols, m.coverRows)
	m.kittyPlacedCols, m.kittyPlacedRows = m.coverCols, m.coverRows
	return m
}

// Helper: Render the cover art with the active renderer
func (m model) renderCoverArt() string {
	if m.albumArtRenderer == AlbumArtRendererKitty {
		return kittyPlaceholder(m.coverCols, m.coverRows)
	}

	return m.coverMosaic.Render(m.coverArt)
}

// Helper: Cell size of the rendered mosaic, so both renderers take the same space
func mosaicCellSize(mo mosaic.Mosaic, img image.Image) (int, int) {
	art := strings.TrimRight(mo.Render(img), "\n")
	return lipgloss.Width(art), lipgloss.Height(art)
}
