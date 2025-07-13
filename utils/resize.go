package utils

import (
	"image"
	"image/color"
)

func Resize(i image.Image, newWidth, newHeight int) image.Image {
	bounds := i.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := float64(x) * float64(width) / float64(newWidth)
			srcY := float64(y) * float64(height) / float64(newHeight)

			x1 := int(srcX)
			y1 := int(srcY)
			x2 := x1 + 1
			y2 := y1 + 1

			if x2 >= width {
				x2 = width - 1
			}
			if y2 >= height {
				y2 = height - 1
			}

			c1 := color.RGBAModel.Convert(i.At(x1, y1)).(color.RGBA)
			c2 := color.RGBAModel.Convert(i.At(x2, y1)).(color.RGBA)
			c3 := color.RGBAModel.Convert(i.At(x1, y2)).(color.RGBA)
			c4 := color.RGBAModel.Convert(i.At(x2, y2)).(color.RGBA)

			wx := srcX - float64(x1)
			wy := srcY - float64(y1)

			r := uint8((1-wx)*(1-wy)*float64(c1.R) + wx*(1-wy)*float64(c2.R) +
				(1-wx)*wy*float64(c3.R) + wx*wy*float64(c4.R))
			g := uint8((1-wx)*(1-wy)*float64(c1.G) + wx*(1-wy)*float64(c2.G) +
				(1-wx)*wy*float64(c3.G) + wx*wy*float64(c4.G))
			b := uint8((1-wx)*(1-wy)*float64(c1.B) + wx*(1-wy)*float64(c2.B) +
				(1-wx)*wy*float64(c3.B) + wx*wy*float64(c4.B))
			a := uint8((1-wx)*(1-wy)*float64(c1.A) + wx*(1-wy)*float64(c2.A) +
				(1-wx)*wy*float64(c3.A) + wx*wy*float64(c4.A))

			newImg.Set(x, y, color.RGBA{r, g, b, a})
		}
	}

	return newImg
}
