//go:build linux || freebsd

package integration

import (
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
)

func newProp(value interface{}, cb func(*prop.Change) *dbus.Error) *prop.Prop {
	return &prop.Prop{
		Value:    value,
		Writable: true,
		Emit:     prop.EmitTrue,
		Callback: cb,
	}
}

var playerProps = map[string]*prop.Prop{
	"PlaybackStatus": newProp("Paused", nil),
	"Rate":           newProp(1.0, nil),
	"Metadata":       newProp(map[string]interface{}{}, nil),
	"Volume":         newProp(float64(100), nil),
	"Position":       newProp(int64(0), nil),
	"MinimumRate":    newProp(1.0, nil),
	"MaximumRate":    newProp(1.0, nil),
	"CanGoNext":      newProp(true, nil),
	"CanGoPrevious":  newProp(true, nil),
	"CanPlay":        newProp(true, nil),
	"CanPause":       newProp(true, nil),
	"CanSeek":        newProp(false, nil),
	"CanControl":     newProp(true, nil),
}

var rootProps = map[string]*prop.Prop{
	"CanQuit":             newProp(true, nil),
	"CanRaise":            newProp(false, nil),
	"HasTrackList":        newProp(false, nil),
	"Identity":            newProp("SubTUI", nil),
	"SupportedUriSchemes": newProp([]string{}, nil),
	"SupportedMimeTypes":  newProp([]string{}, nil),
}

func (m Metadata) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"mpris:length":      m.LengthInMicroseconds(),
		"xesam:title":       m.Title,
		"xesam:artist":      []string{m.Artist},
		"xesam:album":       m.Album,
		"xesam:albumArtist": []string{m.Artist},
		"mpris:artUrl":      m.ImageURL,
		"xesam:userRating":  m.Rating / 5.0,
	}
}
