package main

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/fluffy-melli/visualio/config"
	"github.com/fluffy-melli/visualio/cursor"
	"github.com/fluffy-melli/visualio/graphics"
	"github.com/fluffy-melli/visualio/images"
	"github.com/fluffy-melli/visualio/log"
	"github.com/fluffy-melli/visualio/strings"
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

	Xunit, found := strings.ExtractNumber(configs.ImageResize.Width)
	if !found {
		logs.Panic("X resize value not found")
	}

	Yunit, found := strings.ExtractNumber(configs.ImageResize.Height)
	if !found {
		logs.Panic("Y resize value not found")
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

	screen.OnImage = func(r *graphics.Render, i image.Image) image.Image {
		bounds := i.Bounds()

		var newWidth, newHeight int

		if Xunit.Unit == "%" {
			newWidth = int(float64(bounds.Dx()) * Xunit.Value / 100.0)
		} else {
			newWidth = int(Xunit.Value)
		}

		if Yunit.Unit == "%" {
			newHeight = int(float64(bounds.Dy()) * Yunit.Value / 100.0)
		} else {
			newHeight = int(Yunit.Value)
		}

		return images.Resize(i, newWidth, newHeight)
	}

	position := make(chan cursor.Location)

	screen.Routines = make([]func(s *graphics.Render), 0)

	screen.Routines = append(screen.Routines, cursor.PositionReader(ctx, position))
	screen.Routines = append(screen.Routines, cursor.DeltaHandler(ctx, position))

	err = screen.CreateWindow("visualio", configs.Image.Source)

	if err != nil {
		logs.Panic(err)
	}
}
