package ui

import (
	"fmt"
	"strings"

	"github.com/MattiaPun/SubTUI/internal/api"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	overlay "github.com/rmhubbert/bubbletea-overlay"
)

func (m model) View() string {
	base := m.BaseView()

	if !m.showHelp {
		return base
	}

	bgModel := BackgroundWrapper{RenderedView: base}
	overlayView := overlay.New(m.helpModel, bgModel, overlay.Center, overlay.Center, 0, 0)

	return overlayView.View()
}

func (m model) BaseView() string {
	if m.viewMode == viewLogin {
		return loginView(m)
	}

	// SIZING
	headerHeight := 1

	footerHeight := int(float64(m.height) * 0.10)
	if footerHeight < 5 {
		footerHeight = 5
	}

	mainHeight := m.height - headerHeight - footerHeight - (3 * 2) // 3 sections with each 2 borders (top and bottom)
	if mainHeight < 0 {
		mainHeight = 0
	}

	sidebarWidth := int(float64(m.width) * 0.25)
	mainWidth := m.width - sidebarWidth - 4

	// HEADER
	headerBorder := borderStyle
	if m.focus == focusSearch {
		headerBorder = activeBorderStyle
	}

	topView := headerBorder.
		Width(m.width - 2).
		Height(headerHeight).
		Render(headerContent(m))

	// SIDEBAR
	sideBorder := borderStyle
	if m.focus == focusSidebar {
		sideBorder = activeBorderStyle
	}

	leftPane := sideBorder.
		Width(sidebarWidth).
		Height(mainHeight).
		Render(sidebarContent(m, mainHeight, sidebarWidth))

	// MAIN VIEW
	mainBorder := borderStyle
	if m.focus == focusMain {
		mainBorder = activeBorderStyle
	}

	mainContent := ""
	if m.loading {
		mainContent = "\n  Searching your library..."
	} else if m.displayMode == displaySongs {
		mainContent = mainSongsContent(m, mainWidth, mainHeight)
	} else if m.displayMode == displayAlbums {
		mainContent = mainAlbumsContent(m, mainWidth, mainHeight)
	} else if m.displayMode == displayArtist {
		mainContent = mainArtistContent(m, mainWidth, mainHeight)
	}

	rightPane := mainBorder.
		Width(mainWidth).
		Height(mainHeight).
		Render(mainContent)

	// Join sidebar and main view
	centerView := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// FOOTER
	footerBorder := borderStyle
	if m.focus == focusSong {
		footerBorder = activeBorderStyle
	}

	footerView := footerBorder.
		Width(m.width - 2).
		Height(footerHeight).
		Render(footerContent(m))

	// COMBINE ALL VERTICALLY
	return lipgloss.JoinVertical(lipgloss.Left,
		topView,
		centerView,
		footerView,
	)
}

func truncate(s string, w int) string {
	if w <= 1 {
		return ""
	}
	if len(s) > w {
		return s[:w-1] + "…"
	}
	return s
}

func LimitString(s string, limit int) string {
	if limit <= 0 {
		return ""
	}

	width := runewidth.StringWidth(s)

	if width <= limit {
		padding := strings.Repeat(" ", limit-width)
		return s + padding
	}

	curWidth := 0
	res := ""

	for _, r := range s {
		w := runewidth.RuneWidth(r)

		if curWidth+w > limit {
			break
		}

		res += string(r)
		curWidth += w
	}

	return res + strings.Repeat(" ", limit-curWidth)
}

func loginView(m model) string {
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	errorDisplay := ""
	if m.loginErr != "" {
		errorDisplay = errorStyle.Render(m.loginErr)
	} else {
		errorDisplay = ""
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		loginHeaderStyle.Render("Welcome to SubTUI"),
		"", // Spacer
		m.loginInputs[0].View(),
		m.loginInputs[1].View(),
		m.loginInputs[2].View(),
		"", // Spacer
		errorDisplay,
		loginHelpStyle.Render("[ Press Enter to Login ]"),
	)

	box := loginBoxStyle.Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.NoColor{}),
	)
}

func headerContent(m model) string {

	leftContent := "Search: " + m.textInput.View()
	rightContent := ""

	switch m.filterMode {
	case filterSongs:
		rightContent = "< Songs >"
	case filterAlbums:
		rightContent = "< Albums >"
	case filterArtist:
		rightContent = "< Artist >"
	}

	innerWidth := m.width - 5
	gapWidth := innerWidth - lipgloss.Width(leftContent) - lipgloss.Width(rightContent)
	if gapWidth < 0 {
		gapWidth = 0
	}

	gap := strings.Repeat(" ", gapWidth)
	return leftContent + gap + rightContent
}

func sidebarContent(m model, mainHeight int, sidebarWidth int) string {
	sidebarContent := lipgloss.NewStyle().Bold(true).Render("  ALBUMS") + "\n\n"
	for i, item := range albumTypes {
		if i >= mainHeight-3 {
			break
		}

		cursor := "  "
		style := lipgloss.NewStyle()
		if m.cursorSide == i && m.focus == focusSidebar {
			style = style.Foreground(highlight).Bold(true)
			cursor = "> "
		}

		line := cursor + truncate(item, sidebarWidth-4)
		sidebarContent += style.Render(line) + "\n"
	}

	albumOffset := len(albumTypes)
	if mainHeight-2-albumOffset > 5 { // -2 (album and \n) - albumoffset
		sidebarContent += lipgloss.NewStyle().Bold(true).Render("\n\n  PLAYLISTS") + "\n\n"
		for i, item := range m.playlists {

			playlistMaxHeight := mainHeight - 2 - albumOffset - 4 - 2 // Adding 2 as a margin
			if playlistMaxHeight < i {
				break
			}

			cursor := "  "
			style := lipgloss.NewStyle()
			if m.cursorSide == i+albumOffset && m.focus == focusSidebar {
				style = style.Foreground(highlight).Bold(true)
				cursor = "> "
			}

			line := cursor + truncate(item.Name, sidebarWidth-4)
			sidebarContent += style.Render(line) + "\n"
		}
	}
	return sidebarContent
}

func mainSongsContent(m model, mainWidth int, mainHeight int) string {
	mainContent := ""
	mainTableHeader := ""
	var targetList []api.Song

	if m.viewMode == viewList {
		mainTableHeader = "TITLE"
		targetList = m.songs
		mainContent = "\n  Use the search bar to find Songs."
	} else {
		mainTableHeader = fmt.Sprintf("QUEUE (%d/%d)", m.queueIndex+1, len(m.queue))
		targetList = m.queue
		mainContent = "\n  Queue is empty."
	}

	if len(targetList) == 0 {
		return mainContent
	}

	availableWidth := mainWidth - 4
	colTitle := int(float64(availableWidth) * 0.40)
	colArtist := int(float64(availableWidth) * 0.15)
	colAlbum := int(float64(availableWidth) * 0.25)
	// Time takes whatever is left

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(subtle)
	header := fmt.Sprintf("  %s %s %s %s",
		LimitString(mainTableHeader, colTitle),
		LimitString("ARTIST", colArtist),
		LimitString("ALBUM", colAlbum),
		"TIME",
	)

	mainContent = headerStyle.Render(header) + "\n"
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

		starIcon := " "
		if m.starredMap[song.ID] {
			starIcon = "♥"
		}

		row := fmt.Sprintf("%s %s %s %s %s",
			starIcon,
			LimitString(song.Title, colTitle-2),
			LimitString(song.Artist, colArtist),
			LimitString(song.Album, colAlbum),
			formatDuration(song.Duration),
		)

		mainContent += fmt.Sprintf("%s%s\n", cursor, style.Render(row))
	}

	return mainContent
}

func mainAlbumsContent(m model, mainWidth int, mainHeight int) string {
	if len(m.albums) == 0 {
		return "\n  Use the search bar to find Albums."
	}

	availableWidth := mainWidth - 4
	colAlbum := int(float64(availableWidth) * 0.5)
	colArtist := int(float64(availableWidth) * 0.5)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(subtle)
	header := fmt.Sprintf("  %s %s",
		LimitString("ALBUM", colAlbum),
		LimitString("ARTIST", colArtist),
	)

	mainContent := headerStyle.Render(header) + "\n"
	mainContent += lipgloss.NewStyle().Foreground(subtle).Render("  "+strings.Repeat("-", mainWidth-4)) + "\n"

	headerHeight := 4
	visibleRows := mainHeight - headerHeight
	if visibleRows < 1 {
		visibleRows = 1
	}

	start := m.mainOffset
	end := start + visibleRows
	if end >= len(m.albums) {
		end = len(m.albums)
	}

	for i := start; i <= end; i++ {
		if i >= len(m.albums) {
			break
		}

		album := m.albums[i]

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

		starIcon := " "
		if m.starredMap[album.ID] {
			starIcon = "♥"
		}

		row := fmt.Sprintf("%s %s %s",
			starIcon, // 1 char
			LimitString(album.Name, colAlbum-2),
			LimitString(album.Artist, colArtist),
		)

		mainContent += fmt.Sprintf("%s%s\n", cursor, style.Render(row))
	}

	return mainContent
}

func mainArtistContent(m model, mainWidth int, mainHeight int) string {
	if len(m.artists) == 0 {
		return "\n  Use the search bar to find Artists."
	}

	colArtist := mainWidth - 4
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(subtle)
	header := fmt.Sprintf("  %s", LimitString("ARTIST", colArtist))

	mainContent := headerStyle.Render(header) + "\n"
	mainContent += lipgloss.NewStyle().Foreground(subtle).Render("  "+strings.Repeat("-", mainWidth-4)) + "\n"

	headerHeight := 4
	visibleRows := mainHeight - headerHeight
	if visibleRows < 1 {
		visibleRows = 1
	}

	start := m.mainOffset
	end := start + visibleRows
	if end >= len(m.artists) {
		end = len(m.artists)
	}

	for i := start; i <= end; i++ {
		if i >= len(m.artists) {
			break
		}

		artist := m.artists[i]

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

		starIcon := " "
		if m.starredMap[artist.ID] {
			starIcon = lipgloss.NewStyle().Render("♥︎")
		}

		row := fmt.Sprintf("%s %s",
			starIcon,
			LimitString(artist.Name, colArtist-2),
		)

		mainContent += fmt.Sprintf("%s%s\n", cursor, style.Render(row))
	}

	return mainContent
}

func footerContent(m model) string {
	title := ""
	artistAlbumText := ""

	if m.playerStatus.Title == "<nil>" {
		title = "Nothing playing"
		artistAlbumText = ""
	} else if strings.Contains(m.playerStatus.Title, "stream?c=SubTUI") {
		title = "Loading..."
		artistAlbumText = ""
	} else {
		title = m.playerStatus.Title
		artistAlbumText = m.playerStatus.Artist + " - " + m.playerStatus.Album
	}

	notifyText := ""
	if !m.notify {
		notifyText = "[Silent]"
	}

	topRowGap := m.width - 2 - 3 - 3 - len(notifyText) - len(title)

	if topRowGap > 0 {
		title += strings.Repeat(" ", topRowGap) + notifyText
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

	loopText := ""
	switch m.loopMode {
	case LoopNone:
		loopText = ""
	case LoopAll:
		loopText = "[Loop all]"
	case LoopOne:
		loopText = "[Loop one]"
	}

	bottomRowGap := 0
	bottomRowSpaceTaken := 2 + 3 + 3 + len(artistAlbumText) + len(loopText) // 2: border, 3: spacing, 3: spacing
	if artistAlbumText != "" && m.width != 0 && m.width-bottomRowSpaceTaken > 0 {
		bottomRowGap = m.width - bottomRowSpaceTaken
	} else if m.width != 0 {
		bottomRowGap = m.width - 2 - 3 - 3 - len(loopText)
	}

	bottomRowText := artistAlbumText + strings.Repeat(" ", bottomRowGap) + loopText

	topRow := lipgloss.NewStyle().Bold(true).Foreground(highlight).Render("   " + LimitString(title, m.width-4))
	bottomRow := lipgloss.NewStyle().Foreground(subtle).Render("   " + LimitString(bottomRowText, m.width-4))

	rawProgress := fmt.Sprintf("%s %s %s",
		currStr,
		lipgloss.NewStyle().Foreground(special).Render("["+barStr+"]"),
		durStr,
	)

	rowProgress := lipgloss.NewStyle().
		Width(m.width - 2).
		Align(lipgloss.Center).
		Render(rawProgress)

	return fmt.Sprintf("%s\n%s\n\n%s", topRow, bottomRow, rowProgress)
}

func helpViewContent() string {
	keyStyle := lipgloss.NewStyle().Foreground(special).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(subtle)
	titleStyle := lipgloss.NewStyle().Foreground(highlight).Bold(true).MarginBottom(1)
	colStyle := lipgloss.NewStyle().MarginRight(4)
	sectionSpacer := lipgloss.NewStyle().MarginBottom(2)

	// Helper to format lines
	line := func(key, desc string) string {
		return fmt.Sprintf("%-15s %s", keyStyle.Render(key), descStyle.Render(desc))
	}

	// Helper to render a titled section
	section := func(title string, lines ...string) string {
		content := lipgloss.JoinVertical(lipgloss.Left, lines...)
		return lipgloss.JoinVertical(lipgloss.Left, titleStyle.Render(title), content)
	}

	globalKeybinds := section("NAVIGATION",
		line("Tab", "Cycle focus"),
		line("Shift+Tab", "Cycle focus"),
		line("k / Up", "Move Up"),
		line("j / Down", "Move Down"),
		line("Enter", "Select"),
		line("/", "Search bar"),
		line("q / Ctrl+C", "Quit"),
	)

	searchKeybinds := section("SEARCH",
		line("Ctrl+n", "Filter Next"),
		line("Ctrl+b", "Filter Prev"),
	)

	libraryKeybinds := section("LIBRARY",
		line("gg", "Scroll Top"),
		line("G", "Scroll Bottom"),
		line("ga", "Go to Album"),
		line("gr", "Go to Artist"),
	)

	mediaKeybinds := section("MEDIA",
		line("p", "Play/Pause"),
		line("n", "Next Song"),
		line("b", "Prev Song"),
		line("S", "Shuffle"),
		line("L", "Loop Mode"),
		line("w", "Restart Song"),
		line(",", "Rewind 10s"),
		line(";", "Forward 10s"),
	)

	starredKeybinds := section("FAVORITES",
		line("f", "Like/Unlike"),
		line("F", "View Liked"),
	)

	queueKeybinds := section("QUEUE",
		line("Q", "Toggle View"),
		line("N", "Add Next"),
		line("a", "Add Last"),
		line("d", "Remove"),
		line("D", "Clear All"),
		line("K", "Move Up"),
		line("J", "Move Down"),
	)

	columnLeft := lipgloss.JoinVertical(lipgloss.Left,
		sectionSpacer.Render(globalKeybinds),
		" ", // spacer
		libraryKeybinds,
	)

	columnMiddle := lipgloss.JoinVertical(lipgloss.Left,
		sectionSpacer.Render(mediaKeybinds),
		starredKeybinds,
	)

	columnRight := lipgloss.JoinVertical(lipgloss.Left,
		sectionSpacer.Render(queueKeybinds),
		" ", // spacer
		searchKeybinds,
	)

	content := lipgloss.JoinHorizontal(lipgloss.Top,
		colStyle.Render(columnLeft),
		colStyle.Render(columnMiddle),
		columnRight,
	)

	return activeBorderStyle.Padding(1, 3).Render(content)

}
