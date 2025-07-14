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
	"github.com/fluffy-melli/visualio/update"
	"github.com/fluffy-melli/visualio/utils"
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

	sx, sy := screen.ScreenSize()
	if configs.ImagePosition.X < 0 {
		screen.AX = sx + configs.ImagePosition.X
	} else {
		screen.AX = configs.ImagePosition.X
	}
	if configs.ImagePosition.Y < 0 {
		screen.AY = sy + configs.ImagePosition.Y
	} else {
		screen.AY = configs.ImagePosition.Y
	}

	screen.ProcessImg = func(s *graphics.Render, i image.Image) image.Image {
		bounds := i.Bounds()

		Xunit, found := utils.ExtractNumber(configs.ImageResize.Width)
		if !found {
			logs.Panic("X resize value not found")
		}

		Yunit, found := utils.ExtractNumber(configs.ImageResize.Height)
		if !found {
			logs.Panic("Y resize value not found")
		}

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

		i = images.Resize(i, newWidth, newHeight)

		if s.IsClicked && s.IsInside {
			i = images.DrawBorder(i)
			configs.ImagePosition.X = s.AX
			configs.ImagePosition.Y = s.AY
			config.Save(ConfigFile, configs)
		}

		return i
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
