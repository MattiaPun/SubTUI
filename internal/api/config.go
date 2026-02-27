package api

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/pelletier/go-toml/v2"
)

//go:embed config.toml
var defaultConfig []byte
var AppConfig Config

//go:embed credentials.toml
var defaultServerConfig []byte
var AppServerConfig ServerConfig

type Config struct {
	App      App      `toml:"app"`
	Theme    Theme    `toml:"theme"`
	Filters  Filters  `toml:"filters"`
	Keybinds Keybinds `toml:"keybinds"`
}

type ServerConfig struct {
	Server   Server   `toml:"server"`
	Security Security `toml:"security"`
}

type Server struct {
	URL      string `toml:"url"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type Security struct {
	RedactCredentialsInLogs bool `toml:"redact_credentials_in_logs"`
}

type App struct {
	ReplayGain    string `toml:"replaygain"`
	Notifications bool   `toml:"desktop_notifications"`
	DiscordRPC    bool   `toml:"discord_rich_presence"`
	MouseSupport  bool   `toml:"mouse_support"`
}

type Theme struct {
	Subtle    []string `toml:"subtle"`
	Highlight []string `toml:"highlight"`
	Special   []string `toml:"special"`
	Filtered  []string `toml:"filtered"`
}

type Filters struct {
	Titles           []string `toml:"titles"`
	Artists          []string `toml:"artists"`
	AlbumArtists     []string `toml:"album_artists"`
	MinDuration      int      `toml:"min_duration"`
	Genres           []string `toml:"genres"`
	Notes            []string `toml:"notes"`
	Paths            []string `toml:"paths"`
	MaxPlayCount     int      `toml:"max_play_count"`
	ExcludeFavorites bool     `toml:"exclude_favorites"`
	MaxRating        int      `toml:"max_rating"`
}

type Keybinds struct {
	Global     GlobalKeybinds     `toml:"global"`
	Navigation NavigationKeybinds `toml:"navigation"`
	Search     SearchKeybinds     `toml:"search"`
	Library    LibraryKeybinds    `toml:"library"`
	Media      MediaKeybinds      `toml:"media"`
	Queue      QueueKeybinds      `toml:"queue"`
	Favorites  FavoriteKeybinds   `toml:"favorites"`
	Other      OtherKeybinds      `toml:"other"`
}

type GlobalKeybinds struct {
	CycleFocusNext []string `toml:"cycle_focus_next"`
	CycleFocusPrev []string `toml:"cycle_focus_prev"`
	Back           []string `toml:"back"`
	Help           []string `toml:"help"`
	Quit           []string `toml:"quit"`
	HardQuit       []string `toml:"hard_quit"`
}

type NavigationKeybinds struct {
	Up           []string `toml:"up"`
	Down         []string `toml:"down"`
	Top          []string `toml:"top"`
	Bottom       []string `toml:"bottom"`
	Select       []string `toml:"select"`
	PlayShuffled []string `toml:"play_shuffeled"`
}

type SearchKeybinds struct {
	FocusSearch []string `toml:"focus_search"`
	FilterNext  []string `toml:"filter_next"`
	FilterPrev  []string `toml:"filter_prev"`
}

type LibraryKeybinds struct {
	AddToPlaylist []string `toml:"add_to_playlist"`
	AddRating     []string `toml:"add_rating"`
	GoToAlbum     []string `toml:"go_to_album"`
	GoToArtist    []string `toml:"go_to_artist"`
}

type MediaKeybinds struct {
	PlayPause  []string `toml:"play_pause"`
	Next       []string `toml:"next"`
	Prev       []string `toml:"prev"`
	Shuffle    []string `toml:"shuffle"`
	Loop       []string `toml:"loop"`
	Restart    []string `toml:"restart"`
	Rewind     []string `toml:"rewind"`
	Forward    []string `toml:"forward"`
	VolumeUp   []string `toml:"volume_up"`
	VolumeDown []string `toml:"volume_down"`
}

type QueueKeybinds struct {
	ToggleQueueView []string `toml:"toggle_queue_view"`
	QueueNext       []string `toml:"queue_next"`
	QueueLast       []string `toml:"queue_last"`
	RemoveFromQueue []string `toml:"remove_from_queue"`
	ClearQueue      []string `toml:"clear_queue"`
	MoveUp          []string `toml:"move_up"`
	MoveDown        []string `toml:"move_down"`
}

type FavoriteKeybinds struct {
	ToggleFavorite []string `toml:"toggle_favorite"`
	ViewFavorites  []string `toml:"view_favorites"`
}

type OtherKeybinds struct {
	ToggleNotifications []string `toml:"toggle_notifications"`
	CreateShareLink     []string `toml:"create_share_link"`
}

func getConfigPath(configName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".config", "subtui", configName)
}

func createDefaultConfig(path string, content []byte, label string, permissions os.FileMode) error {
	// Create config dir
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// Write Default config
	if err := os.WriteFile(path, content, permissions); err != nil {
		return err
	}

	log.Printf("[CONFIG] Created default %s config file at %s", label, path)
	return nil
}

func LoadConfig() error {
	// Get config paths
	configPath := getConfigPath("config.toml")
	if configPath == "" {
		return fmt.Errorf("could not determine config path")
	}
	serverConfigPath := getConfigPath("credentials.toml")
	if serverConfigPath == "" {
		return fmt.Errorf("could not determine server config path")
	}

	// Create config files if missing
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath, defaultConfig, "app", 0644); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	}
	if _, err := os.Stat(serverConfigPath); os.IsNotExist(err) {
		if err := createDefaultConfig(serverConfigPath, defaultServerConfig, "server", 0600); err != nil {
			return fmt.Errorf("failed to create default credential config: %w", err)
		}
	}

	// Read config files
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("could not open config file: %v", err)
	}
	serverConfigFile, err := os.ReadFile(serverConfigPath)
	if err != nil {
		return fmt.Errorf("could not open config file: %v", err)
	}

	// Load default config values
	if err := toml.Unmarshal(defaultConfig, &AppConfig); err != nil {
		return fmt.Errorf("could not decode default config: %v", err)
	}
	if err := toml.Unmarshal(defaultServerConfig, &AppServerConfig); err != nil {
		return fmt.Errorf("could not decode default server config: %v", err)
	}

	// Load configs into variables
	var userConfig Config
	var userServerConfig ServerConfig
	if err := toml.Unmarshal(configFile, &userConfig); err != nil {
		return fmt.Errorf("could not decode user config: %v", err)
	}
	if err := toml.Unmarshal(serverConfigFile, &userServerConfig); err != nil {
		return fmt.Errorf("could not decode user server config: %v", err)
	}

	// Overwrite config values with user custom values
	if err := toml.Unmarshal(configFile, &AppConfig); err != nil {
		return fmt.Errorf("could not decode user config: %v", err)
	}
	if err := toml.Unmarshal(serverConfigFile, &AppServerConfig); err != nil {
		return fmt.Errorf("could not decode user server config: %v", err)
	}

	configChanged := !reflect.DeepEqual(userConfig, AppConfig)
	serverConfigChanged := !reflect.DeepEqual(userServerConfig, AppServerConfig)

	// Save if keys were actually added/changed
	if configChanged || serverConfigChanged {
		if err := SaveConfig(); err != nil {
			log.Printf("Warning: failed to migrate config files with new defaults: %v", err)
		}
	}

	return nil
}

func SaveConfig() error {
	// Get config paths
	configPath := getConfigPath("config.toml")
	if configPath == "" {
		return fmt.Errorf("could not determine config path")
	}
	serverConfigPath := getConfigPath("credentials.toml")
	if serverConfigPath == "" {
		return fmt.Errorf("could not determine server config path")
	}

	// Create config dir
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	// Process configs
	configData, err := toml.Marshal(AppConfig)
	if err != nil {
		return err
	}
	serverConfigData, err := toml.Marshal(AppServerConfig)
	if err != nil {
		return err
	}

	// Write configs
	return errors.Join(
		os.WriteFile(configPath, configData, 0644),
		os.WriteFile(serverConfigPath, serverConfigData, 0600),
	)
}
