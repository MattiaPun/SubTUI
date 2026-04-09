package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type LrcLibResponse struct {
	ID           int     `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

func fetchLrcLib(artist, title string) (*LrcLibResponse, error) {
	reqURL, err := url.Parse("https://lrclib.net/api/get")
	if err != nil {
		return nil, err
	}
	q := reqURL.Query()
	q.Add("track_name", title)
	q.Add("artist_name", artist)
	reqURL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "SubTUI (https://github.com/MattiaPun/SubTUI)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, nil // Not found
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("lrclib status code %d", resp.StatusCode)
	}

	var data LrcLibResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func parseSyncedLyrics(synced string) []LyricLine {
	var lines []LyricLine
	for _, l := range strings.Split(synced, "\n") {
		l = strings.TrimSpace(l)
		if l == "" || !strings.HasPrefix(l, "[") {
			continue
		}

		closingIdx := strings.Index(l, "]")
		if closingIdx < 0 {
			continue
		}

		timeStr := l[1:closingIdx]
		text := strings.TrimSpace(l[closingIdx+1:])

		parts := strings.Split(timeStr, ":")
		if len(parts) == 2 {
			min, _ := strconv.Atoi(parts[0])
			secParts := strings.Split(parts[1], ".")
			sec, _ := strconv.Atoi(secParts[0])
			ms := 0
			if len(secParts) > 1 {
				ms, _ = strconv.Atoi(secParts[1])
				// .xx means hundredths of a sec, .xxx means thousandths.
				// The unit in `LyricLine` `Start` is milliseconds.
				if len(secParts[1]) == 2 {
					ms *= 10
				} else if len(secParts[1]) == 1 {
					ms *= 100
				}
			}
			lines = append(lines, LyricLine{
				Start: min*60*1000 + sec*1000 + ms,
				Value: text,
			})
		}
	}
	return lines
}
