package ui

import "github.com/MattiaPun/SubTUI/v2/internal/api"

// preferredStructuredLyrics picks synced lyrics when available, otherwise the first entry.
func preferredStructuredLyrics(structured []api.StructuredLyrics) (api.StructuredLyrics, bool) {
	if len(structured) == 0 {
		return api.StructuredLyrics{}, false
	}

	chosen := structured[0]
	for _, s := range structured {
		if s.Synced {
			chosen = s
			break
		}
	}

	return chosen, true
}

// syncedAutoScrollOffset returns the same auto-scroll offset policy used for synced lyrics.
func syncedAutoScrollOffset(totalLines, currentLine, maxVisible int) int {
	if maxVisible < 1 {
		maxVisible = 1
	}

	linesAfter := totalLines - currentLine
	if linesAfter > maxVisible {
		offset := currentLine - (maxVisible / 2)
		if offset < 0 {
			offset = 0
		}
		return offset
	}

	offset := totalLines - maxVisible
	if offset < 0 {
		offset = 0
	}
	return offset
}
