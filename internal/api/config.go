package api

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Domain   string `yaml:"domain"`
}

var AppConfig Config

func InitConfig() error {
	home, _ := os.UserConfigDir()
	appDir := filepath.Join(home, "DepthTUI")
	configPath := filepath.Join(appDir, "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("--- DepthTUI First Run Setup ---")

		if err := os.MkdirAll(appDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter Subsonic Domain (e.g., music.example.com): ")
		domain, _ := reader.ReadString('\n')

		fmt.Print("Enter Username: ")
		username, _ := reader.ReadString('\n')

		fmt.Print("Enter Password: ")
		password, _ := reader.ReadString('\n')

		AppConfig = Config{
			Domain:   strings.TrimSpace(domain),
			Username: strings.TrimSpace(username),
			Password: strings.TrimSpace(password),
		}

		data, err := yaml.Marshal(AppConfig)
		if err != nil {
			return err
		}
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return err
		}
		fmt.Println("Config saved successfully!")
	} else {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(data, &AppConfig); err != nil {
			return err
		}
	}
	return nil
}
