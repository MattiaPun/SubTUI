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
	if api.AppServerConfig.Server.URL == "" ||
		(api.AppServerConfig.Server.AuthMethod == "plaintext" && (api.AppServerConfig.Server.Username == "" || api.AppServerConfig.Server.Password == "")) ||
		(api.AppServerConfig.Server.AuthMethod == "hashed" && (api.AppServerConfig.Server.Username == "" || api.AppServerConfig.Server.PasswordToken == "" || api.AppServerConfig.Server.PasswordSalt == "")) ||
		(api.AppServerConfig.Server.AuthMethod == "api_key" && (api.AppServerConfig.Server.Username == "" || api.AppServerConfig.Server.ApiKey == "")) {
		startMode = viewLogin
	}

	return model{
		textInput:          ti,
		songs:              []api.Song{},
		focus:              focusSearch,
		cursorMain:         0,
		cursorSide:         0,
		cursorPopup:        0,
		viewMode:           startMode,
		filterMode:         filterSongs,
		displayMode:        displaySongs,
		starredMap:         make(map[string]bool),
		lastPlayedSongPath: "",
		loginInputs:        initialLoginInputs(),
		lastKey:            "",
		showHelp:           false,
		showPlaylists:      false,
		helpModel:          NewHelpModel(),
		discordRPC:         api.AppConfig.App.DiscordRPC,
		notify:             api.AppConfig.App.Notifications,
	}
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, textinput.Blink)

	if m.viewMode == viewList {
		cmds = append(cmds, attemptLoginCmd())
	}

	if api.AppConfig.App.MouseSupport {
		cmds = append(cmds, tea.EnableMouseCellMotion)
	}

	return tea.Batch(cmds...)
}

func initialLoginInputs() []textinput.Model {
	inputs := make([]textinput.Model, 4)

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

	inputs[3] = textinput.New()
	inputs[3].Width = 30

	return inputs
}
