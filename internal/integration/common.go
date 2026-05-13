package integration

type Metadata struct {
	Title    string
	Artist   string
	Album    string
	Duration float64
	Position float64
	ImageURL string
	Rating   float64
}

type Status string

const (
	StatusPlaying Status = "Playing"
	StatusPaused  Status = "Paused"
	StatusStopped Status = "Stopped"
)

type PlayPauseMsg struct{}
type StopMsg struct{}
type NextSongMsg struct{}
type PreviousSongMsg struct{}
type SetPositionMsg struct {
	Position int64
}

func (m Metadata) LengthInMicroseconds() int64 {
	return int64(m.Duration * 1000000)
}
