package config

import (
	"os"

	"github.com/pelletier/go-toml"
)

type App struct {
	Version     string `toml:"version"`
	UpdateCheck bool   `toml:"update-check"`
}

type Image struct {
	Source string `toml:"source"`
}

type ImagePosition struct {
	X int `toml:"x"`
	Y int `toml:"y"`
}

type Config struct {
	App           App           `toml:"app"`
	Image         Image         `toml:"image"`
	ImagePosition ImagePosition `toml:"image-position"`
}

func Load(configPath string) (*Config, error) {
	var config Config

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func Save(configPath string, config *Config) error {
	data, err := toml.Marshal(config)

	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
