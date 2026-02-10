package ui

import (
	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func InitialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Search songs..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	startMode := viewList
	if api.AppConfig.Server.Username == "" || api.AppConfig.Server.Password == "" || api.AppConfig.Server.URL == "" {
		startMode = viewLogin
	}

	return model{
		textInput:        ti,
		songs:            []api.Song{},
		focus:            focusSearch,
		cursorMain:       0,
		cursorSide:       0,
		cursorPopup:      0,
		viewMode:         startMode,
		filterMode:       filterSongs,
		displayMode:      displaySongs,
		starredMap:       make(map[string]bool),
		lastPlayedSongID: "",
		loginInputs:      initialLoginInputs(),
		lastKey:          "",
		showHelp:         false,
		showPlaylists:    false,
		helpModel:        NewHelpModel(),
		discordRPC:       api.AppConfig.App.DiscordRPC,
		notify:           api.AppConfig.App.Notifications,
	}
}

func (m model) Init() tea.Cmd {
	if m.viewMode == viewList {
		return tea.Batch(
			textinput.Blink,
			attemptLoginCmd(),
		)
	}

	return textinput.Blink
}

func initialLoginInputs() []textinput.Model {
	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "http(s)://music.example.com"
	inputs[0].Width = 30
	inputs[0].Focus()
	inputs[0].Prompt = "URL:      "

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "username"
	inputs[1].Width = 30
	inputs[1].Prompt = "Username: "

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "password"
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].Width = 30
	inputs[2].Prompt = "Password: "

	return inputs
}
