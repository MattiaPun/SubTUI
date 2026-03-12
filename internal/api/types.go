package api

import (
	"encoding/json"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"
)

type SubsonicResponse struct {
	Response struct {
		Status            string         `json:"status"`
		User              *SubsonicUser  `json:"user,omitempty"`
		Error             *SubsonicError `json:"error,omitempty"`
		SearchResult      SearchResult3  `json:"searchResult3"`
		PlaylistContainer struct {
			Playlists []Playlist `json:"playlist"`
		} `json:"playlists"`
		PlaylistDetail struct {
			Entries []Song `json:"entry"`
		} `json:"playlist"`
		Album struct {
			Songs []Song `json:"song"`
		} `json:"album"`
		AlbumList struct {
			Albums []Album `json:"album"`
		} `json:"albumList"`
		Artist struct {
			Albums []Album `json:"album"`
		} `json:"artist"`
		Starred2 struct {
			Artist []Artist `json:"artist"`
			Album  []Album  `json:"album"`
			Song   []Song   `json:"song"`
		} `json:"starred2"`
		PlayQueue PlayQueue `json:"playQueue"`
		Shares    struct {
			ShareList []struct {
				URL string `json:"url"`
			} `json:"share"`
		} `json:"shares"`
	} `json:"subsonic-response"`
}

type SubsonicUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type SubsonicError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PlayQueue struct {
	Current string `json:"current"`
	Entries []Song `json:"entry"`
}

type SearchResult3 struct {
	Artists []Artist `json:"artist"`
	Albums  []Album  `json:"album"`
	Songs   []Song   `json:"song"`
}

type Artist struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	AlbumCount int    `json:"albumCount"`
	Rating     int    `json:"userRating"`
}

type Album struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Artist    string `json:"artist"`
	ArtistID  string `json:"artistId"`
	SongCount int    `json:"songCount"`
	Genre     string `json:"genre"`
	Year      int    `json:"year"`
	Rating    int    `json:"userRating"`
	Duration  int    `json:"duration"`
	Note      string `json:"comment"`
}

type Song struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Artist       string   `json:"artist"`
	ArtistID     string   `json:"artistId"`
	AlbumArtists []Artist `json:"albumArtists"`
	Album        string   `json:"album"`
	AlbumID      string   `json:"albumId"`
	Duration     int      `json:"duration"`
	Rating       int      `json:"userRating"`
	Genre        string   `json:"genre"`
	Year         int      `json:"year"`
	Note         string   `json:"comment"`
	Path         string   `json:"path"`
	PlayCount    int      `json:"playCount"`
	TrackNumber  int      `json:"track"`
	DiscNumber   int      `json:"discNumber"`
	Filtered     bool
}

type Playlist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Helper: Unmarshal Song for sanitization
func (s *Song) UnmarshalJSON(data []byte) error {
	type Alias Song
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	s.Title = SanitizeDisplayString(s.Title)
	s.Artist = SanitizeDisplayString(s.Artist)
	s.Album = SanitizeDisplayString(s.Album)

	return nil
}

// Helper: Unmarshal Album for sanitization
func (a *Album) UnmarshalJSON(data []byte) error {
	type Alias Album
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	a.Name = SanitizeDisplayString(a.Name)
	a.Artist = SanitizeDisplayString(a.Artist)

	return nil
}

// Helper: Unmarshal Artists for sanitization
func (a *Artist) UnmarshalJSON(data []byte) error {
	type Alias Artist
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	a.Name = SanitizeDisplayString(a.Name)

	return nil
}

// Helper: Unmarshal Playlists for sanitization
func (p *Playlist) UnmarshalJSON(data []byte) error {
	type Alias Playlist
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.Name = SanitizeDisplayString(p.Name)

	return nil
}

// Helper: Sanitize string
func SanitizeDisplayString(s string) string {
	s = width.Fold.String(s)
	s = norm.NFC.String(s)

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteRune(r)
		case unicode.IsControl(r):
			continue
		case r == '\u200B' || r == '\u200C' || r == '\u200D' || r == '\uFEFF' || r == '\u00AD':
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
