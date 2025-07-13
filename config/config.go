package config

import (
	"github.com/BurntSushi/toml"
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

func LoadConfig(configPath string) (*Config, error) {
	var config Config

	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
