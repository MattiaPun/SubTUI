package ui

import (
	"strings"

	"github.com/MattiaPun/SubTUI/v2/internal/api"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func loginTypeFromAuthMethod(authMethod string) int {
	switch strings.ToLower(authMethod) {
	case "hashed":
		return loginPasswordHashed
	case "api_key":
		return loginApi
	default:
		return loginPassword
	}
}

func configureLoginInputsForType(inputs []textinput.Model, loginType int) {
	inputs[1].Prompt = "Username: "
	inputs[1].Placeholder = "username"
	inputs[1].EchoMode = textinput.EchoNormal

	switch loginType {
	case loginPassword:
		inputs[2].Prompt = "Password: "
		inputs[2].Placeholder = "password"
		inputs[2].EchoMode = textinput.EchoPassword
		inputs[3].Prompt = ""
		inputs[3].Placeholder = ""

	case loginPasswordHashed:
		inputs[2].Prompt = "Token: "
		inputs[2].Placeholder = "md5 hash"
		inputs[2].EchoMode = textinput.EchoNormal
		inputs[3].Prompt = "Salt: "
		inputs[3].Placeholder = "random string"
		inputs[3].EchoMode = textinput.EchoNormal

	case loginApi:
		inputs[2].Prompt = "API Key: "
		inputs[2].Placeholder = "api key"
		inputs[2].EchoMode = textinput.EchoPassword
		inputs[3].Prompt = ""
		inputs[3].Placeholder = ""
	}
}

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

	loginType := loginTypeFromAuthMethod(api.AppServerConfig.Server.AuthMethod)
	loginInputs := initialLoginInputs()
	configureLoginInputsForType(loginInputs, loginType)
	loginInputs[0].SetValue(api.AppServerConfig.Server.URL)
	loginInputs[1].SetValue(api.AppServerConfig.Server.Username)
	api.AppConfig.Lyrics.SourceMode = api.NormalizeLyricsSourceMode(api.AppConfig.Lyrics.SourceMode)
	if loginType == loginPassword {
		loginInputs[2].SetValue(api.AppServerConfig.Server.Password)
	} else if loginType == loginPasswordHashed {
		loginInputs[2].SetValue(api.AppServerConfig.Server.PasswordToken)
		loginInputs[3].SetValue(api.AppServerConfig.Server.PasswordSalt)
	} else {
		loginInputs[2].SetValue(api.AppServerConfig.Server.ApiKey)
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
		loginInputs:        loginInputs,
		loginType:          loginType,
		lastKey:            "",
		showMediaPlayer:    false,
		showHelp:           false,
		showLoginExtras:    false,
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
