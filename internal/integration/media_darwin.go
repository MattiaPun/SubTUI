//go:build darwin

package integration

import tea "github.com/charmbracelet/bubbletea"

type Instance struct{}

func Init(p *tea.Program) *Instance {

	return nil
}

func (ins *Instance) GetStatus() string {
	return "Stopped"
}

func (ins *Instance) UpdateStatus(status string)     {}
func (ins *Instance) UpdatePosition(position int64)  {}
func (ins *Instance) UpdateMetadata(meta Metadata)   {}
func (ins *Instance) ClearMetadata()                 {}
func (ins *Instance) Close()                         {}
func (ins *Instance) ReplaceTrackList(trackIDs []string, metadatas map[string]map[string]interface{}, currentTrack string) {}
func (ins *Instance) EmitTrackAdded(trackIDs []string, trackID string, metadata map[string]interface{}, afterTrack string)  {}
func (ins *Instance) EmitTrackRemoved(trackIDs []string, trackID string)                                                   {}
func (ins *Instance) EmitTrackMetadataChanged(trackID string, metadata map[string]interface{})                              {}
