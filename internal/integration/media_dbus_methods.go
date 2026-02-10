//go:build linux || freebsd

package integration

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
)

type MediaPlayer2 struct {
	Program *tea.Program
}

func (m *MediaPlayer2) Play() *dbus.Error {
	if m.Program != nil {
		m.Program.Send(PlayPauseMsg{})
	}
	return nil
}

func (m *MediaPlayer2) Pause() *dbus.Error {
	if m.Program != nil {
		m.Program.Send(PlayPauseMsg{})
	}
	return nil
}

func (m *MediaPlayer2) PlayPause() *dbus.Error {
	if m.Program != nil {
		m.Program.Send(PlayPauseMsg{})
	}
	return nil
}

func (m *MediaPlayer2) Next() *dbus.Error {
	if m.Program != nil {
		m.Program.Send(NextSongMsg{})
	}
	return nil
}

func (m *MediaPlayer2) Previous() *dbus.Error {
	if m.Program != nil {
		m.Program.Send(PreviousSongMsg{})
	}
	return nil
}

func (m *MediaPlayer2) Stop() *dbus.Error {
	if m.Program != nil {
		m.Program.Send(PlayPauseMsg{})
	}
	return nil
}

func (m *MediaPlayer2) Quit() *dbus.Error {
	if m.Program != nil {
		m.Program.Quit()
	}
	return nil
}

func (m *MediaPlayer2) Raise() *dbus.Error { return nil }
