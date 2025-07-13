package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/fluffy-melli/visualio/config"
	"github.com/fluffy-melli/visualio/utils"
	"github.com/fluffy-melli/visualio/windows"
)

func main() {
	configs, err := config.LoadConfig("config.toml")

	if err != nil {
		panic(err)
	}

	screen := windows.NewScreen()

	screen.X = configs.ImagePosition.X
	screen.Y = configs.ImagePosition.Y

	screen.Do = func(i image.Image) image.Image {
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
		return i
	}

	err = screen.Create("visualio", configs.Image.Source)

	if err != nil {
		panic(err)
	}
}
