package utils

import (
	"image"
	"image/color"
	"image/draw"
)

func DrawBorder(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	draw.Draw(newImg, bounds, img, bounds.Min, draw.Src)

	red := color.RGBA{R: 255, A: 255}
	width := bounds.Max.X
	height := bounds.Max.Y
	for x := 0; x < width; x++ {
		newImg.Set(x, 0, red)
	}
	for x := 0; x < width; x++ {
		newImg.Set(x, height-1, red)
	}
	for y := 0; y < height; y++ {
		newImg.Set(0, y, red)
	}
	for y := 0; y < height; y++ {
		newImg.Set(width-1, y, red)
	}
	return newImg
}
