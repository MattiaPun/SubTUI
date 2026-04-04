package api

import "sync"

var lyricsCache sync.Map

// GetCachedLyrics returns the cached LyricsResult for the song if it exists.
func GetCachedLyrics(songID string) (LyricsResult, bool) {
	val, ok := lyricsCache.Load(songID)
	if ok {
		return val.(LyricsResult), true
	}
	return LyricsResult{}, false
}

// SetCachedLyrics stores the fetched LyricsResult into the cache.
func SetCachedLyrics(songID string, res LyricsResult) {
	lyricsCache.Store(songID, res)
}
