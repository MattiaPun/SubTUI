package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MattiaPun/SubTUI/internal/api"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	overlay "github.com/rmhubbert/bubbletea-overlay"
)

func (m model) View() string {
	if m.width < 50 || m.height < 25 {
		return viewToSmallContent(m)
	}

	base := m.BaseView()

	if m.showPlaylists {
		content := addToPlaylistContent(m)

		styledContent := popupStyle.Render(
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Bold(true).Render("Select Playlist"),
				"",
				content,
			),
		)

		fg := ContentModel{Content: styledContent}
		bg := BackgroundWrapper{RenderedView: base}

		return overlay.New(fg, bg, overlay.Center, overlay.Center, 0, 0).View()
	}

	if m.showRating {
		content := addRatingContent(m)

		styledContent := popupStyle.Render(
			lipgloss.JoinVertical(lipgloss.Center,
				lipgloss.NewStyle().Bold(true).Render("Select Rating"),
				"",
				content,
			),
		)

		fg := ContentModel{Content: styledContent}
		bg := BackgroundWrapper{RenderedView: base}

		return overlay.New(fg, bg, overlay.Center, overlay.Center, 0, 0).View()

	}

	if m.showHelp {
		bg := BackgroundWrapper{RenderedView: base}
		return overlay.New(m.helpModel, bg, overlay.Center, overlay.Center, 0, 0).View()
	}

	return base
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

func formatTime(v int64) string {
	minutes := int(v) / 60
	seconds := int(v) % 60

	return fmt.Sprintf("%d:%02d", minutes, seconds)
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
	content := ""
	currentLine := 0

	totalItems := len(albumTypes) + len(m.playlists)

	for i := m.sideOffset; i < totalItems; i++ {
		// Stop if run out of space - 1
		if currentLine >= mainHeight-1 {
			break
		}

		// Handle Headers
		if i == 0 {
			header := lipgloss.NewStyle().Bold(true).Render("  ALBUMS")
			if currentLine+2 <= mainHeight-1 {
				content += header + "\n\n"
				currentLine += 2
			} else {
				// Not enough space for header + spacing
				break
			}
		} else if i == len(albumTypes) {
			header := lipgloss.NewStyle().Bold(true).Render("  PLAYLISTS")

			// If at top of view, use less padding above
			if i == m.sideOffset {
				if currentLine+2 <= mainHeight-1 {
					content += header + "\n\n"
					currentLine += 2
				} else {
					break
				}
			} else {
				// If not top, use full padding
				if currentLine+3 <= mainHeight-1 {
					content += "\n" + header + "\n\n"
					currentLine += 3
				} else {
					break
				}
			}
		}

		// Double check space for item before rendering
		if currentLine >= mainHeight-1 {
			break
		}

		// Item Logic
		var name string
		if i < len(albumTypes) {
			name = albumTypes[i]
		} else {
			name = m.playlists[i-len(albumTypes)].Name
		}

		cursor := "  "
		style := lipgloss.NewStyle()
		if m.cursorSide == i && m.focus == focusSidebar {
			style = style.Foreground(Theme.Highlight).Bold(true)
			cursor = "> "
		}

		line := cursor + truncate(name, sidebarWidth-4)
		content += style.Render(line) + "\n"
		currentLine++
	}

	return content
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

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(Theme.Subtle)
	header := fmt.Sprintf("  %s %s %s %s",
		LimitString(mainTableHeader, colTitle),
		LimitString("ARTIST", colArtist),
		LimitString("ALBUM", colAlbum),
		"TIME",
	)

	mainContent = headerStyle.Render(header) + "\n"
	mainContent += lipgloss.NewStyle().Foreground(Theme.Subtle).Render("  "+strings.Repeat("-", mainWidth-4)) + "\n"

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
				style = style.Foreground(Theme.Highlight).Bold(true)
			} else {
				style = style.Foreground(Theme.Subtle)
			}
		}

		if m.viewMode == viewQueue && i == m.queueIndex {
			style = style.Foreground(Theme.Special)
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
	colAlbum := int(float64(availableWidth) * 0.45)
	colArtist := int(float64(availableWidth) * 0.45)
	colDuration := int(float64(availableWidth) * 0.1)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(Theme.Subtle)
	header := fmt.Sprintf("  %s %s %s",
		LimitString("ALBUM", colAlbum),
		LimitString("ARTIST", colArtist),
		LimitString("DURATION", colDuration),
	)

	mainContent := headerStyle.Render(header) + "\n"
	mainContent += lipgloss.NewStyle().Foreground(Theme.Subtle).Render("  "+strings.Repeat("-", mainWidth-4)) + "\n"

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
				style = style.Foreground(Theme.Highlight).Bold(true)
			} else {
				style = style.Foreground(Theme.Subtle)
			}
		}

		starIcon := " "
		if m.starredMap[album.ID] {
			starIcon = "♥"
		}

		row := fmt.Sprintf("%s %s %s %s",
			starIcon, // 1 char
			LimitString(album.Name, colAlbum-2),
			LimitString(album.Artist, colArtist),
			LimitString(formatTime(album.Duration), colDuration),
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
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(Theme.Subtle)
	header := fmt.Sprintf("  %s", LimitString("ARTIST", colArtist))

	mainContent := headerStyle.Render(header) + "\n"
	mainContent += lipgloss.NewStyle().Foreground(Theme.Subtle).Render("  "+strings.Repeat("-", mainWidth-4)) + "\n"

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
				style = style.Foreground(Theme.Highlight).Bold(true)
			} else {
				style = style.Foreground(Theme.Subtle)
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

	topRow := lipgloss.NewStyle().Bold(true).Foreground(Theme.Highlight).Render("   " + LimitString(title, m.width-4))
	bottomRow := lipgloss.NewStyle().Foreground(Theme.Subtle).Render("   " + LimitString(bottomRowText, m.width-4))

	rawProgress := fmt.Sprintf("%s %s %s",
		currStr,
		lipgloss.NewStyle().Foreground(Theme.Special).Render("["+barStr+"]"),
		durStr,
	)

	rowProgress := lipgloss.NewStyle().
		Width(m.width - 2).
		Align(lipgloss.Center).
		Render(rawProgress)

	return fmt.Sprintf("%s\n%s\n\n%s", topRow, bottomRow, rowProgress)
}

func helpViewContent() string {
	keyStyle := lipgloss.NewStyle().Foreground(Theme.Special).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(Theme.Subtle)
	titleStyle := lipgloss.NewStyle().Foreground(Theme.Highlight).Bold(true).MarginBottom(1)
	colStyle := lipgloss.NewStyle().MarginRight(4)

	// Helper to format lines
	line := func(key, desc string) string {
		return fmt.Sprintf("%-15s %s", keyStyle.Render(key), descStyle.Render(desc))
	}

	// Helper to render a titled section
	section := func(title string, lines ...string) string {
		content := lipgloss.JoinVertical(lipgloss.Left, lines...)
		return lipgloss.JoinVertical(lipgloss.Left, titleStyle.Render(title), content)
	}

	// Helper to format key lists
	keys := func(k []string) string {
		return strings.Join(k, " / ")
	}

	globalKeybinds := section("GLOBAL",
		line(keys(api.AppConfig.Keybinds.Global.CycleFocusNext), "Cycle focus"),
		line(keys(api.AppConfig.Keybinds.Global.CycleFocusPrev), "Cycle focus"),
		line(keys(api.AppConfig.Keybinds.Global.Back), "Go back"),
		line(keys(api.AppConfig.Keybinds.Global.Help), "Shortcut menu"),
		line(keys(api.AppConfig.Keybinds.Global.Quit), "Quit"),
		line(keys(api.AppConfig.Keybinds.Global.HardQuit), "Quit"),
	)

	navigationKeybinds := section("NAVIGATION",
		line(keys(api.AppConfig.Keybinds.Navigation.Up), "Go up"),
		line(keys(api.AppConfig.Keybinds.Navigation.Down), "Go down"),
		line(keys(api.AppConfig.Keybinds.Navigation.Top), "Go to top"),
		line(keys(api.AppConfig.Keybinds.Navigation.Bottom), "Go to bottom"),
		line(keys(api.AppConfig.Keybinds.Navigation.Select), "Select"),
	)

	searchKeybinds := section("SEARCH",
		line(keys(api.AppConfig.Keybinds.Search.FocusSearch), "Focus search bar"),
		line(keys(api.AppConfig.Keybinds.Search.FilterNext), "Filter next"),
		line(keys(api.AppConfig.Keybinds.Search.FilterPrev), "Filter prev"),
	)

	libraryKeybinds := section("LIBRARY",
		line(keys(api.AppConfig.Keybinds.Library.AddToPlaylist), "Add to playlist"),
		line(keys(api.AppConfig.Keybinds.Library.AddRating), "Add rating"),
		line(keys(api.AppConfig.Keybinds.Library.GoToAlbum), "Go to album"),
		line(keys(api.AppConfig.Keybinds.Library.GoToArtist), "Go to artist"),
	)

	mediaKeybinds := section("MEDIA",
		line(keys(api.AppConfig.Keybinds.Library.GoToArtist), "Play/Pause"),
		line(keys(api.AppConfig.Keybinds.Media.Next), "Next song"),
		line(keys(api.AppConfig.Keybinds.Media.Prev), "Prev song"),
		line(keys(api.AppConfig.Keybinds.Media.Shuffle), "Shuffle"),
		line(keys(api.AppConfig.Keybinds.Media.Loop), "Loop mode"),
		line(keys(api.AppConfig.Keybinds.Media.Restart), "Restart song"),
		line(keys(api.AppConfig.Keybinds.Media.Rewind), "Rewind 10s"),
		line(keys(api.AppConfig.Keybinds.Media.Forward), "Forward 10s"),
	)

	queueKeybinds := section("QUEUE",
		line(keys(api.AppConfig.Keybinds.Queue.ToggleQueueView), "Toggle queue view"),
		line(keys(api.AppConfig.Keybinds.Queue.QueueNext), "Add next"),
		line(keys(api.AppConfig.Keybinds.Queue.QueueLast), "Queue last"),
		line(keys(api.AppConfig.Keybinds.Queue.RemoveFromQueue), "Remove from queue"),
		line(keys(api.AppConfig.Keybinds.Queue.ClearQueue), "Clear queue"),
		line(keys(api.AppConfig.Keybinds.Queue.MoveUp), "Queue up"),
		line(keys(api.AppConfig.Keybinds.Queue.MoveDown), "Queue down"),
	)

	starredKeybinds := section("FAVORITES",
		line(keys(api.AppConfig.Keybinds.Favorites.ToggleFavorite), "Toggle fav"),
		line(keys(api.AppConfig.Keybinds.Favorites.ViewFavorites), "View fav"),
	)

	otherKeybinds := section("OTHERS",
		line(keys(api.AppConfig.Keybinds.Other.ToggleNotifications), "Toggle notifications"),
		line(keys(api.AppConfig.Keybinds.Other.CreateShareLink), "Create share link"),
	)

	columnLeft := lipgloss.JoinVertical(lipgloss.Left,
		globalKeybinds,
		"", // spacer
		libraryKeybinds,
		"", // spacer
		otherKeybinds,
	)

	columnMiddle := lipgloss.JoinVertical(lipgloss.Left,
		mediaKeybinds,
		"", // spacer
		navigationKeybinds,
	)

	columnRight := lipgloss.JoinVertical(lipgloss.Left,
		queueKeybinds,
		"", // spacer
		starredKeybinds,
		"", // spacer
		searchKeybinds,
	)

	content := lipgloss.JoinHorizontal(lipgloss.Top,
		colStyle.Render(columnLeft),
		colStyle.Render(columnMiddle),
		columnRight,
	)

	return activeBorderStyle.Padding(1, 3).Render(content)

}

func addToPlaylistContent(m model) string {
	playlistContent := ""
	for i := 0; i < len(m.playlists); i++ {
		cursor := ""
		style := lipgloss.NewStyle()

		if m.cursorPopup == i {
			style = style.Foreground(Theme.Highlight).Bold(true)
			cursor = "> "
		}

		playlistContent += fmt.Sprintf("%s%s\n", cursor, style.Render(m.playlists[i].Name))

	}

	return playlistContent
}

func addRatingContent(m model) string {
	ratingContent := ""
	for i := 0; i <= 5; i++ {
		cursor := ""
		style := lipgloss.NewStyle()

		if m.cursorPopup == i {
			style = style.Foreground(Theme.Highlight).Bold(true)
			cursor = "> "
		} else {
			cursor = "  "
		}

		stars := strings.Repeat("★", i)

		ratingContent += fmt.Sprintf("%s%s %s\n", cursor, style.Render(strconv.Itoa(i)), stars)
	}

	return lipgloss.NewStyle().Align(lipgloss.Left).Render(ratingContent)
}

func viewToSmallContent(m model) string {
	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Align(lipgloss.Center).
		Render("Viewport too small")

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
