package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Dirs  []string `yaml:"dirs"`
	URL   string   `yaml:"url"`
	Org   string   `yaml:"org"`
	Token string   `yaml:"token"`
	OCR   string   `yaml:"ocr"`
}

func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user home dir: %w", err)
	}

	appConfigDir := filepath.Join(homeDir, ".config", "paprawatch")
	if _, err := os.Stat(appConfigDir); os.IsNotExist(err) {
		if err := os.MkdirAll(appConfigDir, 0755); err != nil {
			return "", fmt.Errorf("could not create config directory: %w", err)
		}
	}

	return filepath.Join(appConfigDir, "config.yaml"), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if file exists, if not return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Dirs: []string{"."},
			URL:  "https://api.papra.app",
		}, nil
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	return nil
}

func (c *Config) RunSetup() error {
	var dirsInput string
	if len(c.Dirs) > 0 {
		dirsInput = strings.Join(c.Dirs, ", ")
	} else {
		dirsInput = "."
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Welcome to Paprawatch!").
				Description("Let's set up your watcher in a few easy steps."),

			huh.NewInput().
				Title("Organization ID").
				Description("Your Papra Organization ID (found in settings)").
				Value(&c.Org).
				Placeholder("Required").
				Validate(func(str string) error {
					if str == "" {
						return fmt.Errorf("organization ID is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("API Token").
				Description("Your Papra API Token").
				Value(&c.Token).
				EchoMode(huh.EchoModePassword).
				Placeholder("Required").
				Validate(func(str string) error {
					if str == "" {
						return fmt.Errorf("token is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Watch Directories").
				Description("Comma-separated list of directories to watch").
				Value(&dirsInput).
				Placeholder("."),

			huh.NewInput().
				Title("API URL").
				Description("The Papra instance URL").
				Value(&c.URL).
				Placeholder("https://api.papra.app"),

			huh.NewInput().
				Title("OCR Languages").
				Description("Optional: Comma-separated languages (e.g. 'eng,fra')").
				Value(&c.OCR),
		),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	// Process dirsInput
	dirs := strings.Split(dirsInput, ",")
	c.Dirs = []string{}
	for _, d := range dirs {
		trimmed := strings.TrimSpace(d)
		if trimmed != "" {
			c.Dirs = append(c.Dirs, trimmed)
		}
	}
	if len(c.Dirs) == 0 {
		c.Dirs = []string{"."}
	}

	if c.URL == "" {
		c.URL = "https://api.papra.app"
	}

	return c.Save()
}
