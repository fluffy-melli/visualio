package main

import (
	"context"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/fluffy-melli/visualio/config"
	"github.com/fluffy-melli/visualio/utils"
	"github.com/fluffy-melli/visualio/windows"
)

var ConfigFile = "config.toml"

func main() {
	configs, err := config.Load(ConfigFile)

	if err != nil {
		panic(err)
	}

	screen := windows.NewScreen()

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

	screen.ProcessImg = func(s *windows.Screen, i image.Image) image.Image {
		bounds := i.Bounds()

		Xunit, found := utils.ExtractNumber(configs.ImageResize.Width)
		if !found {
			panic("X resize value not found")
		}

		Yunit, found := utils.ExtractNumber(configs.ImageResize.Height)
		if !found {
			panic("Y resize value not found")
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

		i = utils.Resize(i, newWidth, newHeight)

		if s.IsClicked && s.IsInside {
			i = utils.DrawBorder(i)
			configs.ImagePosition.X = s.AX
			configs.ImagePosition.Y = s.AY
			config.Save(ConfigFile, configs)
		}

		return i
	}

	position := make(chan windows.Point)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	screen.Routines = make([]func(s *windows.Screen), 0)

	screen.Routines = append(screen.Routines, windows.CursorPositionReader(ctx, position))
	screen.Routines = append(screen.Routines, windows.CursorDeltaHandler(ctx, position))

	err = screen.CreateWindow("visualio", configs.Image.Source)

	if err != nil {
		panic(err)
	}
}
