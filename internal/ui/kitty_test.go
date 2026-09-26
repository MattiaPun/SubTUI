package ui

import (
	"image"
	"os"
	"strings"
	"testing"

	"github.com/MattiaPun/SubTUI/v2/internal/api"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

func TestKittyPlaceholderSize(t *testing.T) {
	s := kittyPlaceholder(16, 8)
	if w := lipgloss.Width(s); w != 16 {
		t.Errorf("width = %d, want 16", w)
	}
	if h := lipgloss.Height(s); h != 8 {
		t.Errorf("height = %d, want 8", h)
	}
}

func clearTermEnv(t *testing.T) {
	for _, k := range []string{"TMUX", "STY", "TERM", "TERM_PROGRAM", "KITTY_WINDOW_ID"} {
		t.Setenv(k, "")
	}
}

func TestResolveAlbumArtRenderer(t *testing.T) {
	clearTermEnv(t)
	if r := resolveAlbumArtRenderer("auto"); r != AlbumArtRendererMosaic {
		t.Errorf("auto without kitty = %q", r)
	}
	if r := resolveAlbumArtRenderer("kitty"); r != AlbumArtRendererKitty {
		t.Errorf("explicit kitty = %q", r)
	}
	if r := resolveAlbumArtRenderer("bogus"); r != AlbumArtRendererMosaic {
		t.Errorf("unknown = %q", r)
	}

	t.Setenv("TERM", "xterm-kitty")
	if r := resolveAlbumArtRenderer("auto"); r != AlbumArtRendererKitty {
		t.Errorf("auto in kitty = %q", r)
	}
	if r := resolveAlbumArtRenderer("mosaic"); r != AlbumArtRendererMosaic {
		t.Errorf("explicit mosaic = %q", r)
	}

	t.Setenv("TMUX", "/tmp/tmux-1000/default,1,0")
	if r := resolveAlbumArtRenderer("auto"); r != AlbumArtRendererMosaic {
		t.Errorf("auto in tmux = %q", r)
	}
}

func TestKittySequences(t *testing.T) {
	if s := kittyDeleteSeq(); !strings.HasPrefix(s, "\x1b_G") || !strings.Contains(s, "a=d,d=I,i=") {
		t.Errorf("delete seq = %q", s)
	}

	place := kittyPlaceSeq(16, 8)
	if !strings.Contains(place, "a=p,U=1,") || !strings.Contains(place, "c=16,r=8") {
		t.Errorf("place seq = %q", place)
	}

	seq, err := kittyTransmitSeq(image.NewRGBA(image.Rect(0, 0, 4, 4)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(seq, "\x1b_Gf=100,") {
		t.Errorf("transmit seq missing PNG format: %q", seq[:min(len(seq), 80)])
	}
}

func TestKittyTransmitHasPayload(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 300, 300))
	for i := range img.Pix {
		img.Pix[i] = byte(i * 7)
	}

	seq, err := kittyTransmitSeq(img)
	if err != nil {
		t.Fatal(err)
	}
	// First image chunk must carry base64 PNG data, and a multi-chunk transfer must end with m=0
	if !strings.Contains(seq, "i=371569,m=1;iVBOR") {
		t.Errorf("transmit seq has no PNG payload: %q", seq[:min(len(seq), 120)])
	}
	if !strings.Contains(seq, "m=0;") {
		t.Error("transmit seq has no final chunk")
	}
}

// Kitty art must take exactly the space the mosaic takes, in the footer and the media player
func TestKittyLayoutMatchesMosaic(t *testing.T) {
	zone.NewGlobal()
	clearTermEnv(t)
	api.AppConfig.Theme.DisplayAlbumArt = true
	defer func() { api.AppConfig.Theme.DisplayAlbumArt = false }()

	// Keep kitty sequences out of the test output
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer devNull.Close()
	prevOut := TermOut
	TermOut = &lockedTerm{f: devNull}
	defer func() { TermOut = prevOut }()
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))

	for _, mediaPlayer := range []bool{false, true} {
		render := func(renderer string) string {
			m := InitialModel()
			m.viewMode = viewList
			m.width, m.height = 120, 40
			m.showMediaPlayer = mediaPlayer
			m.albumArtRenderer = renderer
			res, _ := m.handleCoverArt(coverArtMsg{img: img, resize: true})
			return res.(model).View()
		}

		mosaicView := strings.Split(render(AlbumArtRendererMosaic), "\n")
		kittyView := strings.Split(render(AlbumArtRendererKitty), "\n")

		if len(kittyView) != len(mosaicView) {
			t.Fatalf("mediaPlayer=%v: kitty view has %d lines, mosaic %d", mediaPlayer, len(kittyView), len(mosaicView))
		}
		for i := range kittyView {
			if kw, mw := lipgloss.Width(kittyView[i]), lipgloss.Width(mosaicView[i]); kw != mw {
				t.Errorf("mediaPlayer=%v line %d: kitty width %d, mosaic width %d", mediaPlayer, i, kw, mw)
			}
		}
	}
}
