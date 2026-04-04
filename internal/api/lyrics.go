package api

import "fmt"

// LyricsResult is the unified result returned by fetching lyrics,
// regardless of which endpoint was used.
type LyricsResult struct {
	// Plain is set when using the classic getLyrics endpoint.
	Plain string
	// Structured is set when using getLyricsBySongId (OpenSubsonic).
	// Prefer the first entry that has Synced == true; fall back to unsynced.
	Structured []StructuredLyrics
}

// GetLyrics calls the classic Subsonic getLyrics endpoint.
// It takes the song artist and title as parameters.
// Available since Subsonic API 1.2.0. Supported by all servers.
func GetLyrics(artist, title string) (string, error) {
	params := map[string]string{
		"artist": artist,
		"title":  title,
	}

	data, err := subsonicGET("/getLyrics", params)
	if err != nil {
		return "", err
	}

	if data.Response.Status == "failed" && data.Response.Error != nil {
		return "", fmt.Errorf("api error: %s", data.Response.Error.Message)
	}

	return data.Response.Lyrics.Value, nil
}

// GetLyricsBySongID calls the OpenSubsonic getLyricsBySongId endpoint.
// It takes the song's Subsonic ID.
// Only available on OpenSubsonic-compatible servers (Navidrome ≥ 0.50, Gonic, etc.).
// Returns nil, nil if the server returns a "not supported" error code — the caller
// should treat this as a signal to fall back to GetLyrics.
func GetLyricsBySongID(songID string) ([]StructuredLyrics, error) {
	params := map[string]string{
		"id": songID,
	}

	data, err := subsonicGET("/getLyricsBySongId", params)
	if err != nil {
		return nil, err
	}

	if data.Response.Status == "failed" && data.Response.Error != nil {
		if data.Response.Error.Code == 70 {
			return nil, nil
		}
		return nil, fmt.Errorf("api error: %s", data.Response.Error.Message)
	}

	return data.Response.LyricsList.StructuredLyrics, nil
}

// FetchLyrics attempts to get lyrics for the given song.
// It first tries getLyricsBySongId (OpenSubsonic), falls back to LRCLIB, and then to getLyrics.
// Returns a LyricsResult and any error. A missing/empty result is not an error.
func FetchLyrics(songID, artist, title string) (LyricsResult, error) {
	if cached, ok := GetCachedLyrics(songID); ok {
		return cached, nil
	}

	structured, err := GetLyricsBySongID(songID)
	// If we got valid structured lyrics directly from Subsonic, use them.
	if err == nil && len(structured) > 0 {
		res := LyricsResult{Structured: structured}
		SetCachedLyrics(songID, res)
		return res, nil
	}

	// Try LRCLIB for synced lyrics fallback
	lrc, err := fetchLrcLib(artist, title)
	if err == nil && lrc != nil {
		if lrc.SyncedLyrics != "" {
			lines := parseSyncedLyrics(lrc.SyncedLyrics)
			if len(lines) > 0 {
				res := LyricsResult{
					Structured: []StructuredLyrics{
						{
							DisplayArtist: lrc.ArtistName,
							DisplayTitle:  lrc.TrackName,
							Synced:        true,
							Lines:         lines,
						},
					},
				}
				SetCachedLyrics(songID, res)
				return res, nil
			}
		}
		if lrc.PlainLyrics != "" {
			res := LyricsResult{Plain: lrc.PlainLyrics}
			SetCachedLyrics(songID, res)
			return res, nil
		}
	}

	plain, err := GetLyrics(artist, title)
	if err != nil {
		return LyricsResult{}, err
	}
	res := LyricsResult{Plain: plain}
	SetCachedLyrics(songID, res)
	return res, nil
}
