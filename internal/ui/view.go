package ui

import (
	"fmt"
	"strings"

	"git.punjwani.pm/Mattia/DepthTUI/internal/api"
	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	// SIZING
	searchHeight := int(float64(m.height) * 0.035)
	mainHeight := int(float64(m.height) * 0.75)
	footerHeight := int(float64(m.height) * 0.104)

	sidebarWidth := int(float64(m.width) * 0.25)
	mainWidth := m.width - sidebarWidth - 4

	// SEARCH BAR
	searchBorder := borderStyle
	if m.focus == focusSearch {
		searchBorder = activeBorderStyle
	}

	searchView := searchBorder.Width(m.width - 2).Height(searchHeight).Render("Search: " + m.textInput.View())

	// SIZE BAR
	sideBorder := borderStyle
	if m.focus == focusSidebar {
		sideBorder = activeBorderStyle
	}

	sidebarContent := lipgloss.NewStyle().Bold(true).Render("  PLAYLISTS") + "\n\n"

	for i, item := range m.playlists {

		if i >= mainHeight-3 {
			break
		}

		cursor := "  "
		if m.cursorSide == i && m.focus == focusSidebar {
			cursor = "> "
		}

		style := lipgloss.NewStyle()
		if m.cursorSide == i && m.focus == focusSidebar {
			style = style.Foreground(highlight).Bold(true)
		}

		line := cursor + item.Name
		sidebarContent += style.Render(line) + "\n"
	}

	leftPane := sideBorder.
		Width(sidebarWidth).
		Height(mainHeight).
		Render(sidebarContent)

	// MAIN VIEW
	mainBorder := borderStyle
	if m.focus == focusMain {
		mainBorder = activeBorderStyle
	}

	mainContent := ""

	var targetList []api.Song
	headerTitle := "TITLE"

	if m.viewMode == viewQueue {
		targetList = m.queue
		headerTitle = fmt.Sprintf("QUEUE (%d/%d)", m.queueIndex+1, len(m.queue))
	} else {
		targetList = m.songs
	}

	if m.loading {
		mainContent = "\n  Searching your library..."
	} else if len(targetList) == 0 {
		if m.viewMode == viewList {
			mainContent = "\n  Use the search bar to find music."

		} else if m.viewMode == viewQueue {
			mainContent = "\n  Queue is empty."
		}

	} else {
		availableWidth := mainWidth - 4
		colTitle := int(float64(availableWidth) * 0.40)
		colArtist := int(float64(availableWidth) * 0.15)
		colAlbum := int(float64(availableWidth) * 0.25)
		// Time takes whatever is left

		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(subtle)
		header := fmt.Sprintf("  %-*s %-*s %-*s %s",
			colTitle, headerTitle,
			colArtist, "ARTIST",
			colAlbum, "ALBUM",
			"TIME")

		mainContent += headerStyle.Render(header) + "\n"
		mainContent += lipgloss.NewStyle().Foreground(subtle).Render("  "+strings.Repeat("-", mainWidth-4)) + "\n"

		headerHeight := 4
		visibleRows := mainHeight - headerHeight
		if visibleRows < 1 {
			visibleRows = 1
		}

		start := m.mainOffset

		end := start + visibleRows
		if end >= len(targetList) {
			end = len(targetList)
		}

		for i := start; i <= end; i++ {
			if i >= len(targetList) {
				break
			}

			song := targetList[i]

			cursor := "  "
			style := lipgloss.NewStyle()

			if m.cursorMain == i {
				cursor = "> "
				if m.focus == focusMain {
					style = style.Foreground(highlight).Bold(true)
				} else {
					style = style.Foreground(subtle)
				}
			}

			if m.viewMode == viewQueue && i == m.queueIndex {
				style = style.Foreground(special)
				if m.cursorMain == i {
					cursor = "> "
				} else {
					cursor = "  "
				}
			}

			trunc := func(s string, w int) string {
				if w <= 1 {
					return ""
				}
				if len(s) > w {
					return s[:w-1] + "…"
				}
				return s
			}

			row := fmt.Sprintf("%-*s %-*s %-*s %s",
				colTitle, trunc(song.Title, colTitle),
				colArtist, trunc(song.Artist, colArtist),
				colAlbum, trunc(song.Album, colAlbum),
				formatDuration(song.Duration),
			)

			mainContent += fmt.Sprintf("%s%s\n", cursor, style.Render(row))
		}
	}

	rightPane := mainBorder.
		Width(mainWidth).
		Height(mainHeight).
		Render(mainContent)

	// Join sidebar and main view
	centerView := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	songBorder := borderStyle
	if m.focus == focusSong {
		songBorder = activeBorderStyle
	}

	// FOOTER
	title := ""
	artist := ""

	if m.playerStatus.Title == "" {
		title = "Not Playing"
	} else {
		title = m.playerStatus.Title
		artist = m.playerStatus.Artist + " - " + m.playerStatus.Album
	}

	barWidth := m.width - 20
	if barWidth < 10 {
		barWidth = 10
	}

	percent := 0.0
	if m.playerStatus.Duration > 0 {
		percent = m.playerStatus.Current / m.playerStatus.Duration
	}
	filledChars := int(percent * float64(barWidth))
	if filledChars > barWidth {
		filledChars = barWidth
	}

	barStr := ""
	if filledChars > 0 {
		barStr = strings.Repeat("=", filledChars-1) + ">"
	}
	emptyChars := barWidth - filledChars
	if emptyChars > 0 {
		barStr += strings.Repeat("-", emptyChars)
	}

	currStr := formatDuration(int(m.playerStatus.Current))
	durStr := formatDuration(int(m.playerStatus.Duration))

	rowTitle := lipgloss.NewStyle().Bold(true).Foreground(highlight).Render("  " + title)
	rowArtist := lipgloss.NewStyle().Foreground(subtle).Render("  " + artist)
	rowProgress := fmt.Sprintf("  %s %s %s",
		currStr,
		lipgloss.NewStyle().Foreground(special).Render("["+barStr+"]"),
		durStr,
	)

	footerContent := fmt.Sprintf("%s\n%s\n\n%s", rowTitle, rowArtist, rowProgress)

	footerView := songBorder.
		Width(m.width - 2).
		Height(footerHeight).
		Render(footerContent)

	// COMBINE ALL VERTICALLY
	return lipgloss.JoinVertical(lipgloss.Left,
		searchView,
		centerView,
		footerView,
	)
}
