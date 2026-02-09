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
type NextSongMsg struct{}
type PreviousSongMsg struct{}

func (m Metadata) LengthInMicroseconds() int64 {
	return int64(m.Duration * 1000000)
}
