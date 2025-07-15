package main

import (
	"context"
	"fmt"
	_ "image/jpeg"
	_ "image/png"

	"github.com/fluffy-melli/visualio/config"
	"github.com/fluffy-melli/visualio/cursor"
	"github.com/fluffy-melli/visualio/graphics"
	"github.com/fluffy-melli/visualio/log"
	"github.com/fluffy-melli/visualio/update"
)

var ConfigFile = "config.toml"
var ErrorLogs = "error.log"

func main() {
	logs := log.NewLogger(ErrorLogs)

	configs, err := config.Load(ConfigFile)

	if err != nil {
		logs.Panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if configs.App.UpdateCheck {
		githubs := update.NewClient("")
		release, err := githubs.GetLatestRelease(ctx, "fluffy-melli", "visualio")
		if err != nil {
			logs.Panic(fmt.Errorf("failed to check for updates: %w", err))
		}

		currentVersion := configs.App.Version
		latestVersion := release.Name

		fmt.Printf("Current version: %s\n", currentVersion)
		fmt.Printf("Latest version: %s\n", latestVersion)

		if latestVersion != currentVersion {
			logs.Panic(fmt.Errorf("UPDATE REQUIRED!\n"+
				"Current version: %s\n"+
				"Latest version: %s\n"+
				"Please update to the latest version from: https://github.com/fluffy-melli/visualio/releases/latest",
				currentVersion, latestVersion))
		}
	}

	screen := graphics.NewScreen()

	screen.AX = configs.ImagePosition.X
	screen.AY = configs.ImagePosition.Y

	screen.OnUpMButton = func(r *graphics.Render) {
		configs.ImagePosition.X = r.AX
		configs.ImagePosition.Y = r.AY
		config.Save(ConfigFile, configs)
	}

	screen.OnDownMButton = func(r *graphics.Render) {}

	position := make(chan cursor.Location)

	screen.Routines = make([]func(s *graphics.Render), 0)

	screen.Routines = append(screen.Routines, cursor.PositionReader(ctx, position))
	screen.Routines = append(screen.Routines, cursor.DeltaHandler(ctx, position))

	err = screen.CreateWindow("visualio", configs.Image.Source)

	if err != nil {
		logs.Panic(err)
	}
}
