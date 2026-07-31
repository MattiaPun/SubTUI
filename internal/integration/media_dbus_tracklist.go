//go:build linux || freebsd

package integration

import (
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
)

const NoTrack = "/org/mpris/MediaPlayer2/TrackList/NoTrack"

var trackListProps = map[string]*prop.Prop{
	"Tracks": {
		Value:    []dbus.ObjectPath{},
		Writable: true,
		Emit:     prop.EmitFalse,
	},
	"CanEditTracks": {
		Value:    true,
		Writable: true,
		Emit:     prop.EmitTrue,
	},
}

func (m *MediaPlayer2) GetTracksMetadata(trackIds []dbus.ObjectPath) (map[dbus.ObjectPath]map[string]interface{}, *dbus.Error) {
	if m.Instance == nil {
		return nil, nil
	}
	m.Instance.mu.RLock()
	defer m.Instance.mu.RUnlock()
	result := make(map[dbus.ObjectPath]map[string]interface{})
	for _, id := range trackIds {
		if meta, ok := m.Instance.trackMetas[id]; ok {
			result[id] = meta
		}
	}
	return result, nil
}

func (m *MediaPlayer2) AddTrack(uri string, afterTrack dbus.ObjectPath, setAsCurrent bool) *dbus.Error {
	if m.Program != nil {
		m.Program.Send(AddTrackRequestMsg{
			URI:          uri,
			AfterTrack:   string(afterTrack),
			SetAsCurrent: setAsCurrent,
		})
	}
	return nil
}

func (m *MediaPlayer2) RemoveTrack(trackId dbus.ObjectPath) *dbus.Error {
	if m.Program != nil {
		m.Program.Send(RemoveTrackRequestMsg{
			TrackID: string(trackId),
		})
	}
	return nil
}

func (m *MediaPlayer2) GoTo(trackId dbus.ObjectPath) *dbus.Error {
	if m.Program != nil {
		m.Program.Send(GoToRequestMsg{
			TrackID: string(trackId),
		})
	}
	return nil
}

func toDBusObjectPaths(ids []string) []dbus.ObjectPath {
	paths := make([]dbus.ObjectPath, len(ids))
	for i, id := range ids {
		paths[i] = dbus.ObjectPath(id)
	}
	return paths
}

func toDBusMetaMap(metas map[string]map[string]interface{}) map[dbus.ObjectPath]map[string]interface{} {
	result := make(map[dbus.ObjectPath]map[string]interface{}, len(metas))
	for k, v := range metas {
		result[dbus.ObjectPath(k)] = v
	}
	return result
}

func (ins *Instance) ReplaceTrackList(trackIDs []string, metadatas map[string]map[string]interface{}, currentTrack string) {
	if ins == nil || ins.conn == nil {
		return
	}

	dbusIDs := toDBusObjectPaths(trackIDs)
	dbusMetas := toDBusMetaMap(metadatas)
	dbusCurrent := dbus.ObjectPath(currentTrack)

	ins.mu.Lock()
	ins.trackMetas = dbusMetas
	ins.mu.Unlock()

	_ = ins.props.Set("org.mpris.MediaPlayer2.TrackList", "Tracks", dbus.MakeVariant(dbusIDs))
	_ = ins.conn.Emit("/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.TrackList.TrackListReplaced", dbusIDs, dbusCurrent)
}

func (ins *Instance) EmitTrackAdded(trackIDs []string, trackID string, metadata map[string]interface{}, afterTrack string) {
	if ins == nil || ins.conn == nil {
		return
	}

	dbusIDs := toDBusObjectPaths(trackIDs)
	dbusID := dbus.ObjectPath(trackID)
	dbusAfter := dbus.ObjectPath(afterTrack)

	ins.mu.Lock()
	ins.trackMetas[dbusID] = metadata
	ins.mu.Unlock()

	_ = ins.props.Set("org.mpris.MediaPlayer2.TrackList", "Tracks", dbus.MakeVariant(dbusIDs))
	_ = ins.conn.Emit("/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.TrackList.TrackAdded", toVariantMap(metadata), dbusAfter)
}

func (ins *Instance) EmitTrackRemoved(trackIDs []string, trackID string) {
	if ins == nil || ins.conn == nil {
		return
	}

	dbusIDs := toDBusObjectPaths(trackIDs)
	dbusID := dbus.ObjectPath(trackID)

	ins.mu.Lock()
	delete(ins.trackMetas, dbusID)
	ins.mu.Unlock()

	_ = ins.props.Set("org.mpris.MediaPlayer2.TrackList", "Tracks", dbus.MakeVariant(dbusIDs))
	_ = ins.conn.Emit("/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.TrackList.TrackRemoved", dbusID)
}

func (ins *Instance) EmitTrackMetadataChanged(trackID string, metadata map[string]interface{}) {
	if ins == nil || ins.conn == nil {
		return
	}

	dbusID := dbus.ObjectPath(trackID)

	ins.mu.Lock()
	ins.trackMetas[dbusID] = metadata
	ins.mu.Unlock()

	_ = ins.conn.Emit("/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.TrackList.TrackMetadataChanged", dbusID, toVariantMap(metadata))
}

func toVariantMap(meta map[string]interface{}) map[string]dbus.Variant {
	result := make(map[string]dbus.Variant, len(meta))
	for k, v := range meta {
		result[k] = dbus.MakeVariant(v)
	}
	return result
}


