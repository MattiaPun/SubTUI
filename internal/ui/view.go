package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
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

	return zone.Scan(base)
}

func (m model) BaseView() string {
	if m.viewMode == viewLogin {
		return loginView(m)
	}

	// SIZING
	headerHeight := 1
	footerHeight := 6

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
		Render(searchbarContent(m))

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
	if m.loading &&
		(m.displayMode == displaySongs && len(m.songs) == 0 ||
			m.displayMode == displayAlbums && len(m.albums) == 0 ||
			m.displayMode == displayArtist && len(m.artists) == 0) {
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
	footerView := m.buildFooterBorder().
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

// Generate the login view
func loginView(m model) string {
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	authStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)

	errorDisplay := ""
	if m.loginErr != "" {
		errorDisplay = errorStyle.Render(m.loginErr)
	}

	urlBar := m.loginInputs[0].View()
	var authMethodBar string
	var loginInputs string

	switch m.loginType {
	case loginPassword:
		authMethodBar = authStyle.Render("Auth Method: < Password >")
		loginInputs = lipgloss.JoinVertical(lipgloss.Left,
			m.loginInputs[1].View(), // Username
			m.loginInputs[2].View(), // Password
		)

	case loginPasswordHashed:
		authMethodBar = authStyle.Render("Auth Method: < Token + Salt >")
		loginInputs = lipgloss.JoinVertical(lipgloss.Left,
			m.loginInputs[1].View(), // Username
			m.loginInputs[2].View(), // Token
			m.loginInputs[3].View(), // Salt
		)

	case loginApi:
		authMethodBar = authStyle.Render("Auth Method: < API Key >")
		loginInputs = lipgloss.JoinVertical(lipgloss.Left,
			m.loginInputs[1].View(), // Username
			m.loginInputs[2].View(), // Api Key
		)
	}

	form := lipgloss.JoinVertical(lipgloss.Left,
		urlBar,
		"", // Spacer
		authMethodBar,
		"", // Spacer
		loginInputs,
	)

	content := lipgloss.JoinVertical(lipgloss.Center,
		loginHeaderStyle.Render("Welcome to SubTUI"),
		"", // Spacer
		form,
		"", // Spacer
		errorDisplay,
		loginHelpStyle.Render("[TAB] Next Field   [Ctrl+t] Change Auth   [ENTER] Login"),
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		loginBoxStyle.Render(content),
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.NoColor{}),
	)
}

// Generate the search bar
func searchbarContent(m model) string {

	leftContent := "Search: " + m.textInput.View()
	filterMode := ""

	switch m.filterMode {
	case filterSongs:
		filterMode = "Songs"
	case filterAlbums:
		filterMode = "Albums"
	case filterArtist:
		filterMode = "Artist"
	}

	rightContent := fmt.Sprintf("%s %s %s", zone.Mark("filter_prev", "<"), filterMode, zone.Mark("filter_next", ">"))

	innerWidth := m.width - 5
	gapWidth := innerWidth - lipgloss.Width(leftContent) - lipgloss.Width(rightContent)
	if gapWidth < 0 {
		gapWidth = 0
	}

	gap := strings.Repeat(" ", gapWidth)
	return leftContent + gap + rightContent
}

// Generate the side bar
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
			style = highlightStyle.Bold(true)
			cursor = "> "
		}

		line := cursor + truncate(name, sidebarWidth-4)

		id := fmt.Sprintf("sidebar_item_%d", i)
		content += zone.Mark(id, style.Render(line)) + "\n"
		currentLine++
	}

	return content
}

// Generate the main view for songs
func mainSongsContent(m model, width int, height int) string {
	var headerRow string
	var songRows string

	var target []api.Song
	var headerTitle string
	var emptyStatus string

	var availableWidth int
	var availableHeight int

	var songsDisplayStart int
	var songsDisplayStop int

	switch m.viewMode {
	case viewList:
		target = m.songs
		headerTitle = "TITLE"
		emptyStatus = "\n  Use the search bar to find songs..."
	case viewQueue:
		target = m.queue
		headerTitle = fmt.Sprintf("QUEUE (%d/%d)", m.queueIndex+1, len(m.queue))
		emptyStatus = "\n  Start playing some music to view your queue..."
	}

	// Return if no songs
	if len(target) == 0 {
		return emptyStatus
	}

	// Get available width and height
	availableWidth, availableHeight = calculateMainWidthAndHeight(width, height)

	activeColumns := getSongColumns(headerTitle)                         // Get active columns
	columnsWidth := calculateColumnsWidth(activeColumns, availableWidth) // Get active columns width
	headerRow = generateHeader(activeColumns, columnsWidth)              // Generate header

	songsDisplayStart = m.mainOffset
	songsDisplayStop = songsDisplayStart + availableHeight

	// Display songs
	for i := songsDisplayStart; i <= songsDisplayStop; i++ {
		var style lipgloss.Style
		var row string
		var zoneId string
		var song api.Song

		// Return if no more songs
		if i >= len(target) {
			break
		}

		// Get song
		song = target[i]

		// Cursor
		row, style = generateCursor(m, i)

		// Current playing song
		if len(m.queue) > 0 && song.ID == m.queue[m.queueIndex].ID {
			style = currentPlaySongStyle
		}

		// Favorite album
		row += generateStar(m, song.ID)

		// Filtered song
		if song.Filtered {
			style = filteredStyle
		}

		// Render song information
		for i, column := range activeColumns {
			row += LimitString(column.Value(song), columnsWidth[i])

			// Add padding
			if i < len(activeColumns)-1 {
				row += " "
			}
		}

		// Apply styling
		row = style.Render(row)

		// Add zone ID
		zoneId = fmt.Sprintf("mainview_item_%d", i)
		songRows += fmt.Sprintf("%s\n", zone.Mark(zoneId, row))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerRow,
		subtleStyle.Render("  "+strings.Repeat("-", availableWidth)),
		songRows,
	)
}

// Generate the main view for albums
func mainAlbumsContent(m model, width int, height int) string {
	var headerRow string
	var albumRows string

	var availableHeight int
	var availableWidth int

	var albumDisplayStart int
	var albumDisplayStop int

	// Return if no albums
	if len(m.albums) == 0 {
		return "\n  Use the search bar to find albums..."
	}

	// Get available width and height
	availableWidth, availableHeight = calculateMainWidthAndHeight(width, height)

	activeColumns := getAlbumColumns()                                   // Get active columns
	columnsWidth := calculateColumnsWidth(activeColumns, availableWidth) // Get active columns width
	headerRow = generateHeader(activeColumns, columnsWidth)              // Generate header

	albumDisplayStart = m.mainOffset
	albumDisplayStop = albumDisplayStart + availableHeight

	for i := albumDisplayStart; i <= albumDisplayStop; i++ {
		var style lipgloss.Style
		var row string
		var zoneId string
		var album api.Album

		// Return if no more albums
		if i >= len(m.albums) {
			break
		}

		album = m.albums[i]

		// Cursor
		row, style = generateCursor(m, i)

		// Favorite album
		row += generateStar(m, album.ID)

		// Render album information
		for i, column := range activeColumns {
			row += LimitString(column.Value(album), columnsWidth[i])

			// Add padding
			if i < len(activeColumns)-1 {
				row += " "
			}
		}

		// Apply styling
		row = style.Render(row)

		// Add zone ID
		zoneId = fmt.Sprintf("mainview_item_%d", i)
		albumRows += fmt.Sprintf("%s\n", zone.Mark(zoneId, row))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerRow,
		subtleStyle.Render("  "+strings.Repeat("-", availableWidth)),
		albumRows,
	)
}

// Generate the main view for artist
func mainArtistContent(m model, width int, height int) string {
	var headerRow string
	var artistRows string

	var availableHeight int
	var availableWidth int

	var artistDisplayStart int
	var artistDisplayStop int

	// Return if no artists
	if len(m.artists) == 0 {
		return "\n  Use the search bar to find artists..."
	}

	// Get available width and height
	availableWidth, availableHeight = calculateMainWidthAndHeight(width, height)

	activeColumns := getArtistColumns()                                  // Get active columns
	columnsWidth := calculateColumnsWidth(activeColumns, availableWidth) // Get active columns width
	headerRow = generateHeader(activeColumns, columnsWidth)              // Generate header

	artistDisplayStart = m.mainOffset
	artistDisplayStop = artistDisplayStart + availableHeight

	for i := artistDisplayStart; i <= artistDisplayStop; i++ {
		var style lipgloss.Style
		var row string
		var zoneId string
		var artist api.Artist

		// Return if no more artists
		if i >= len(m.artists) {
			break
		}

		artist = m.artists[i]

		// Cursor
		row, style = generateCursor(m, i)

		// Favorite artist
		row += generateStar(m, artist.ID)

		// Render album information
		for i, column := range activeColumns {
			row += LimitString(column.Value(artist), columnsWidth[i])

			// Add padding
			if i < len(activeColumns)-1 {
				row += " "
			}
		}

		// Apply styling
		row = style.Render(row)

		// Add zone ID
		zoneId = fmt.Sprintf("mainview_item_%d", i)
		artistRows += fmt.Sprintf("%s\n", zone.Mark(zoneId, row))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerRow,
		subtleStyle.Render("  "+strings.Repeat("-", availableWidth)),
		artistRows,
	)
}

// Generate the footer
func footerContent(m model) string {
	var content string

	if api.AppConfig.Theme.DisplayAlbumArt && m.coverArt != nil {
		albumArt := m.coverMosaic.Render(m.coverArt)
		infoText := footerInformation(m, m.width-16)

		content = lipgloss.JoinHorizontal(lipgloss.Left, "  ", albumArt, "  ", infoText)
	} else {
		infoText := footerInformation(m, m.width-6)

		content = lipgloss.JoinHorizontal(lipgloss.Left, "  ", infoText, "  ")
	}

	return "\n" + content
}

func (m model) buildFooterBorder() lipgloss.Style {
	base := borderStyle
	if m.focus == focusSong {
		base = activeBorderStyle
	}

	if len(m.queue) > 0 {
		queueStatus := fmt.Sprintf(" QUEUE (%d/%d) ", m.queueIndex+1, len(m.queue))

		b := lipgloss.RoundedBorder()
		topWidth := max(m.width-2-len(queueStatus)-2, 0)
		b.Top = strings.Repeat("─", topWidth) + queueStatus + "──"

		return base.Border(b)
	}

	return base
}

// Generate the footer information
func footerInformation(m model, width int) string {
	var topRow string
	var middleRow string
	var bottomRow string

	// Top row
	var songTitle string
	var notifcationStatus string

	if m.playerStatus.Title == "<nil>" {
		songTitle = "Nothing playing"
	} else if strings.Contains(m.playerStatus.Title, "stream?c=SubTUI") {
		songTitle = "Loading..."
	} else {
		songTitle = api.SanitizeDisplayString(m.playerStatus.Title)
	}

	if !m.notify {
		notifcationStatus = "[Silent]"
	}

	topRowGap := width - runewidth.StringWidth(songTitle) - runewidth.StringWidth(notifcationStatus)
	if topRowGap < 0 {
		topRowGap = 0
	}

	truncatedTitle := truncate(songTitle, width-runewidth.StringWidth(notifcationStatus))
	topRow = lipgloss.JoinHorizontal(
		lipgloss.Center,
		highlightStyle.Render(truncatedTitle),
		strings.Repeat(" ", topRowGap),
		notifcationStatus,
	)

	// Middle row
	var songAlbumArtistInfo string
	var loopStatus string
	var volumeStatus string

	if m.playerStatus.Title == "<nil>" {
		songAlbumArtistInfo = ""
	} else if strings.Contains(m.playerStatus.Title, "stream?c=SubTUI") {
		songAlbumArtistInfo = ""
	} else {
		songAlbumArtistInfo = api.SanitizeDisplayString(m.playerStatus.Artist + " - " + m.playerStatus.Album)
	}

	switch m.loopMode {
	case LoopNone:
		loopStatus = ""
	case LoopAll:
		loopStatus = "[Loop all]"
	case LoopOne:
		loopStatus = "[Loop one]"
	}

	if m.playerStatus.Volume != 100 {
		volumeStatus = fmt.Sprintf("[%v%%]", m.playerStatus.Volume)
	}

	middleRowGap := width - runewidth.StringWidth(songAlbumArtistInfo) - runewidth.StringWidth(loopStatus) - 1 - runewidth.StringWidth(volumeStatus)
	if middleRowGap < 0 {
		middleRowGap = 0
	}

	truncatedSongAlbumArtistInfo := truncate(songAlbumArtistInfo, width-runewidth.StringWidth(loopStatus)-1-runewidth.StringWidth(volumeStatus))
	middleRow = lipgloss.JoinHorizontal(
		lipgloss.Center,
		truncatedSongAlbumArtistInfo,
		strings.Repeat(" ", middleRowGap),
		loopStatus,
		" ",
		volumeStatus,
	)

	// Bottom row
	var currentTime string
	var progressBar string
	var totalTime string

	currentTime = formatDuration(int(m.playerStatus.Current))
	totalTime = formatDuration(int(m.playerStatus.Duration))

	percent := 0.0
	if m.playerStatus.Duration > 0 {
		percent = m.playerStatus.Current / m.playerStatus.Duration
	}
	infoLen := len(currentTime) + 4 + len(totalTime) // 2x padding
	progressLen := int(percent * float64(width-infoLen))
	progressBar += " [" + strings.Repeat("=", progressLen) + ">"
	progressBar += strings.Repeat("-", width-infoLen-progressLen-1) + "] " // >-char

	bottomRow = lipgloss.JoinHorizontal(
		lipgloss.Center,
		currentTime,
		specialStyle.Render(progressBar),
		totalTime,
	)

	return lipgloss.JoinVertical(lipgloss.Center, topRow, middleRow, "", bottomRow)
}

// Generete the help overlay
func helpViewContent() string {
	keyStyle := specialStyle.Bold(true)
	descStyle := subtleStyle
	titleStyle := highlightStyle.Bold(true).MarginBottom(1)
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
		line(keys(api.AppConfig.Keybinds.Navigation.PlayShuffled), "Start shuffled"),
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
		line(keys(api.AppConfig.Keybinds.Media.PlayPause), "Play/Pause"),
		line(keys(api.AppConfig.Keybinds.Media.Next), "Next song"),
		line(keys(api.AppConfig.Keybinds.Media.Prev), "Prev song"),
		line(keys(api.AppConfig.Keybinds.Media.Shuffle), "Shuffle"),
		line(keys(api.AppConfig.Keybinds.Media.Loop), "Loop mode"),
		line(keys(api.AppConfig.Keybinds.Media.Restart), "Restart song"),
		line(keys(api.AppConfig.Keybinds.Media.Rewind), "Rewind 10s"),
		line(keys(api.AppConfig.Keybinds.Media.Forward), "Forward 10s"),
		line(keys(api.AppConfig.Keybinds.Media.VolumeUp), "Volume up"),
		line(keys(api.AppConfig.Keybinds.Media.VolumeDown), "Volume down"),
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

// Generete the add-to-playlist overlay
func addToPlaylistContent(m model) string {
	playlistContent := ""
	for i := 0; i < len(m.playlists); i++ {
		cursor := ""
		style := lipgloss.NewStyle()

		if m.cursorPopup == i {
			style = highlightStyle.Bold(true)
			cursor = "> "
		}

		playlistContent += fmt.Sprintf("%s%s\n", cursor, style.Render(m.playlists[i].Name))

	}

	return playlistContent
}

// Generete the add-rating overlay
func addRatingContent(m model) string {
	ratingContent := ""
	for i := 0; i <= 5; i++ {
		cursor := ""
		style := lipgloss.NewStyle()

		if m.cursorPopup == i {
			style = highlightStyle.Bold(true)
			cursor = "> "
		} else {
			cursor = "  "
		}

		stars := strings.Repeat("★", i)

		ratingContent += fmt.Sprintf("%s%s %s\n", cursor, style.Render(strconv.Itoa(i)), stars)
	}

	return lipgloss.NewStyle().Align(lipgloss.Left).Render(ratingContent)
}

// Generete the view for a too small viewport
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

// Helper: Cut of strings at a specified width
func truncate(s string, width int) string {
	if width <= 1 {
		return ""
	}

	if len(s) > width {
		return s[:width-1] + "…"
	}

	return s
}

// Helper: Cut of string OR add padding to get specified width
func LimitString(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	stringWidth := runewidth.StringWidth(s)

	if stringWidth <= maxWidth {
		padding := strings.Repeat(" ", maxWidth-stringWidth)
		return s + padding
	}

	currentWidth := 0
	result := ""

	for _, r := range s {
		w := runewidth.RuneWidth(r)

		if currentWidth+w > maxWidth {
			break
		}

		result += string(r)
		currentWidth += w
	}

	return result + strings.Repeat(" ", maxWidth-currentWidth)
}

// Helper: Get all active song columns
func getSongColumns(headerTitle string) []headerColumn[api.Song] {
	var cols []headerColumn[api.Song]

	if api.AppConfig.Columns.Songs.ShowTrackNumber {
		cols = append(cols, headerColumn[api.Song]{
			Title:      "#",
			FixedWidth: 4,
			Value: func(s api.Song) string {
				if s.DiscNumber > 0 {
					return fmt.Sprintf("%d-%d", s.DiscNumber, s.TrackNumber)
				} else if s.TrackNumber > 0 {
					return fmt.Sprintf("%d", s.TrackNumber)
				}
				return "-"
			},
		})
	}

	if api.AppConfig.Columns.Songs.ShowTitle {
		cols = append(cols, headerColumn[api.Song]{
			Title:  headerTitle,
			Weight: 0.40,
			Value:  func(s api.Song) string { return s.Title },
		})
	}

	if api.AppConfig.Columns.Songs.ShowArtist {
		cols = append(cols, headerColumn[api.Song]{
			Title:  "ARTIST",
			Weight: 0.30,
			Value:  func(s api.Song) string { return s.Artist },
		})
	}

	if api.AppConfig.Columns.Songs.ShowAlbum {
		cols = append(cols, headerColumn[api.Song]{
			Title:  "ALBUM",
			Weight: 0.30,
			Value:  func(s api.Song) string { return s.Album },
		})
	}

	if api.AppConfig.Columns.Songs.ShowYear {
		cols = append(cols, headerColumn[api.Song]{
			Title:      "YEAR",
			FixedWidth: 4,
			Value:      func(s api.Song) string { return fmt.Sprintf("%d", s.Year) },
		})
	}

	if api.AppConfig.Columns.Songs.ShowGenre {
		cols = append(cols, headerColumn[api.Song]{
			Title:  "GENRE",
			Weight: 0.25,
			Value:  func(s api.Song) string { return s.Genre },
		})
	}

	if api.AppConfig.Columns.Songs.ShowRating {
		cols = append(cols, headerColumn[api.Song]{
			Title:      "RATE",
			FixedWidth: 4,
			Value:      func(s api.Song) string { return fmt.Sprintf("%d", s.Rating) },
		})
	}

	if api.AppConfig.Columns.Songs.ShowPlayCount {
		cols = append(cols, headerColumn[api.Song]{
			Title:      "PLAYS",
			FixedWidth: 5,
			Value:      func(s api.Song) string { return fmt.Sprintf("%d", s.PlayCount) },
		})
	}

	if api.AppConfig.Columns.Songs.ShowDuration {
		cols = append(cols, headerColumn[api.Song]{
			Title:      "TIME",
			FixedWidth: 6,
			Value:      func(s api.Song) string { return formatDuration(s.Duration) },
		})
	}

	return cols
}

// Helper: Get all active album columns
func getAlbumColumns() []headerColumn[api.Album] {
	var cols []headerColumn[api.Album]

	if api.AppConfig.Columns.Albums.ShowName {
		cols = append(cols, headerColumn[api.Album]{
			Title:  "ALBUM",
			Weight: 0.60,
			Value:  func(a api.Album) string { return a.Name },
		})
	}

	if api.AppConfig.Columns.Albums.ShowArtists {
		cols = append(cols, headerColumn[api.Album]{
			Title:  "ARTIST",
			Weight: 0.50,
			Value:  func(a api.Album) string { return a.Artist },
		})
	}

	if api.AppConfig.Columns.Albums.ShowSongcount {
		cols = append(cols, headerColumn[api.Album]{
			Title:      "SONGS",
			FixedWidth: 5,
			Value:      func(a api.Album) string { return fmt.Sprintf("%d", a.SongCount) },
		})
	}

	if api.AppConfig.Columns.Albums.ShowYear {
		cols = append(cols, headerColumn[api.Album]{
			Title:      "YEAR",
			FixedWidth: 4,
			Value:      func(a api.Album) string { return fmt.Sprintf("%d", a.Year) },
		})
	}

	if api.AppConfig.Columns.Albums.ShowGenre {
		cols = append(cols, headerColumn[api.Album]{
			Title:  "GENRE",
			Weight: 0.15,
			Value:  func(a api.Album) string { return a.Genre },
		})
	}

	if api.AppConfig.Columns.Albums.ShowRating {
		cols = append(cols, headerColumn[api.Album]{
			Title:      "RATE",
			FixedWidth: 4,
			Value:      func(a api.Album) string { return fmt.Sprintf("%d", a.Rating) },
		})
	}

	if api.AppConfig.Columns.Albums.ShowDuration {
		cols = append(cols, headerColumn[api.Album]{
			Title:      "TIME",
			FixedWidth: 6,
			Value:      func(a api.Album) string { return formatDuration(a.Duration) },
		})
	}

	return cols
}

// Helper: Get all active artist columns
func getArtistColumns() []headerColumn[api.Artist] {
	var cols []headerColumn[api.Artist]

	if api.AppConfig.Columns.Albums.ShowName {
		cols = append(cols, headerColumn[api.Artist]{
			Title:  "ARTIST",
			Weight: 0.40,
			Value:  func(a api.Artist) string { return a.Name },
		})
	}

	if api.AppConfig.Columns.Albums.ShowSongcount {
		cols = append(cols, headerColumn[api.Artist]{
			Title:      "ALBUMS",
			FixedWidth: 6,
			Value:      func(a api.Artist) string { return fmt.Sprintf("%d", a.AlbumCount) },
		})
	}

	if api.AppConfig.Columns.Albums.ShowRating {
		cols = append(cols, headerColumn[api.Artist]{
			Title:      "RATE",
			FixedWidth: 4,
			Value:      func(a api.Artist) string { return fmt.Sprintf("%d", a.Rating) },
		})
	}

	return cols
}

// Helper: Calculate the availble width and height for main view
func calculateMainWidthAndHeight(width int, height int) (int, int) {
	availableWidth := width - 2 - 2       // 2: borders | 2: padding right
	availableHeight := height - 2 - 1 - 1 // 2: borders | 1: header | 1: seperator

	if availableHeight < 1 {
		availableHeight = 1
	}

	return availableWidth, availableHeight
}

// Helper: Calculate the column values
func calculateColumnsWidth[T any](cols []headerColumn[T], totalWidth int) []int {
	availableWidth := totalWidth - 2 - 2 - 2 - 2 // 1: Border | 1: padding | 1: cursor | 1: star

	var weightTotal float64
	var fixedTotal int

	for _, column := range cols {
		if column.FixedWidth > 0 {
			fixedTotal += column.FixedWidth
		} else {
			weightTotal += column.Weight
		}
	}

	dynamicSpace := availableWidth - fixedTotal
	if dynamicSpace < 10 {
		dynamicSpace = 10
	}

	widths := make([]int, len(cols))
	for i, column := range cols {
		if column.FixedWidth > 0 {
			widths[i] = column.FixedWidth
		} else {
			widths[i] = int((column.Weight / weightTotal) * float64(dynamicSpace))
		}
	}

	return widths
}

// Helper: Generate header
func generateHeader[T any](cols []headerColumn[T], widths []int) string {
	headerText := "  "

	for i, column := range cols {
		headerText += LimitString(column.Title, widths[i])

		// Add padding on all but last column
		if i < len(cols)-1 {
			headerText += " "
		}
	}

	headerStyle := subtleStyle.Bold(true)
	return headerStyle.Render("  " + strings.TrimRight(headerText, " "))
}

// Helper: Generate cursor
func generateCursor(m model, index int) (string, lipgloss.Style) {
	var cursor string
	var style lipgloss.Style

	if m.cursorMain == index {
		cursor += "> "

		// Highlight when focussed
		if m.focus == focusMain {
			style = cursorFocusedStyle
		} else {
			style = cursorStyle
		}

	} else {
		cursor += "  "
	}

	return cursor, style
}

// Helper: Generate star
func generateStar(m model, ID string) string {
	if m.starredMap[ID] {
		return "♥ "
	}

	return "  "
}
