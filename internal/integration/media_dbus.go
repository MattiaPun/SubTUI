//go:build linux || freebsd

package integration

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
)

type Instance struct {
	props *prop.Properties
	conn  *dbus.Conn
}

func Init(p *tea.Program) *Instance {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil
	}

	ins := &Instance{conn: conn}

	mp2 := &MediaPlayer2{Program: p}
	err = conn.Export(mp2, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player")
	if err != nil {
		log.Printf("MPRIS Export Error: %v", err)
		return nil
	}

	err = conn.Export(mp2, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2")
	if err != nil {
		log.Printf("MPRIS Root Export Error: %v", err)
	}

	ins.props, _ = prop.Export(
		conn,
		"/org/mpris/MediaPlayer2",
		map[string]map[string]*prop.Prop{
			"org.mpris.MediaPlayer2":        rootProps,
			"org.mpris.MediaPlayer2.Player": playerProps,
		},
	)

	reply, err := conn.RequestName("org.mpris.MediaPlayer2.subtui", dbus.NameFlagReplaceExisting)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		log.Printf("MPRIS Name Error: %v", err)
	}

	return ins
}

func (ins *Instance) GetStatus() string {
	if ins == nil || ins.props == nil {
		return "Stopped"
	}

	variant, err := ins.props.Get("org.mpris.MediaPlayer2.Player", "PlaybackStatus")
	if err != nil {
		return "Stopped"
	}

	if s, ok := variant.Value().(string); ok {
		return s
	}

	return "Stopped"
}

func (ins *Instance) UpdateStatus(status string) {
	if ins == nil {
		return
	}

	_ = ins.props.Set("org.mpris.MediaPlayer2.Player", "PlaybackStatus", dbus.MakeVariant(string(status)))
}

func (ins *Instance) UpdateMetadata(meta Metadata) {
	if ins == nil {
		return
	}

	_ = ins.props.Set("org.mpris.MediaPlayer2.Player", "Metadata", dbus.MakeVariant(meta.ToMap()))
}

func (ins *Instance) ClearMetadata() {
	if ins == nil {
		return
	}

	emptyMeta := make(map[string]dbus.Variant)
	emptyMeta["mpris:trackid"] = dbus.MakeVariant(dbus.ObjectPath("/org/mpris/MediaPlayer2/TrackList/NoTrack"))

	_ = ins.props.Set("org.mpris.MediaPlayer2.Player", "Metadata", dbus.MakeVariant(emptyMeta))
	_ = ins.props.Set("org.mpris.MediaPlayer2.Player", "PlaybackStatus", dbus.MakeVariant("Stopped"))
}

func (ins *Instance) Close() {
	if ins != nil && ins.conn != nil {
		_ = ins.conn.Close()
	}
}
