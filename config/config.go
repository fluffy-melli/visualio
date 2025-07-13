package config

import (
	"os"

	"github.com/pelletier/go-toml"
)

type Image struct {
	Source string `toml:"source"`
}

type ImagePosition struct {
	X int `toml:"x"`
	Y int `toml:"y"`
}

type ImageResize struct {
	Width  string `toml:"width"`
	Height string `toml:"height"`
}

type Config struct {
	Image         Image         `toml:"image"`
	ImagePosition ImagePosition `toml:"image-position"`
	ImageResize   ImageResize   `toml:"image-resize"`
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
