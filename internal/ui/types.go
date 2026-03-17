package ui

const (
	focusSearch = iota
	focusSidebar
	focusMain
	focusSong
	focusPlaylist = 90
)

const (
	viewList = iota
	viewQueue
	viewLogin = 99
)

const (
	filterSongs = iota
	filterAlbums
	filterArtist
)

const (
	displaySongs = iota
	displayAlbums
	displayArtist
)

const (
	LoopNone = 0
	LoopAll  = 1
	LoopOne  = 2
)

const (
	loginPassword = iota
	loginPasswordHashed
	loginApi
)

const (
	pastLine = iota
	currentLine
	futureLine
)

type headerColumn[T any] struct {
	Title      string
	FixedWidth int
	Weight     float64
	Value      func(item T) string
}
