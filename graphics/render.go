package graphics

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
	"time"
	"unsafe"

	"github.com/fluffy-melli/visualio/constants"
	"github.com/gonutz/d3d9"
	"golang.org/x/sys/windows"
)

type Animator struct {
	hwnd              windows.HWND
	device            *d3d9.Device
	textures          []*d3d9.Texture
	frames            []image.Image
	delays            []int
	currentFrame      int
	done              chan bool
	isAnimated        bool
	staticImage       image.Image
	staticTexture     *d3d9.Texture
	isPreprocessed    bool
	bounds            image.Rectangle
	processedBounds   image.Rectangle
	processFunc       func(*Render, image.Image) image.Image
	render            *Render
	needsUpdate       bool
	processedTextures []*d3d9.Texture
	processedStatic   *d3d9.Texture
}

func NewGPUAnimator(device *d3d9.Device, imagePath string) (*Animator, error) {
	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, err
	}

	if len(imageBytes) > 3 && string(imageBytes[:3]) == "GIF" {
		return loadGPUGifAnimation(device, imageBytes)
	}
	return loadGPUStaticImage(device, imageBytes)
}

func loadGPUGifAnimation(device *d3d9.Device, imageBytes []byte) (*Animator, error) {
	gifImg, err := gif.DecodeAll(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}

	animator := &Animator{
		device:            device,
		frames:            make([]image.Image, len(gifImg.Image)),
		textures:          make([]*d3d9.Texture, len(gifImg.Image)),
		processedTextures: make([]*d3d9.Texture, len(gifImg.Image)),
		delays:            make([]int, len(gifImg.Image)),
		currentFrame:      0,
		done:              make(chan bool),
		isAnimated:        true,
		isPreprocessed:    false,
		bounds:            gifImg.Image[0].Bounds(),
		needsUpdate:       true,
	}

	bounds := gifImg.Image[0].Bounds()
	accumulated := image.NewRGBA(bounds)

	for i, frame := range gifImg.Image {
		delay := max(gifImg.Delay[i]*10, 20)
		animator.delays[i] = delay

		disposal := gif.DisposalNone
		if i < len(gifImg.Disposal) {
			disposal = int(gifImg.Disposal[i])
		}

		if i > 0 && disposal == gif.DisposalBackground {
			draw.Draw(accumulated, bounds, &image.Uniform{color.RGBA{0, 0, 0, 0}}, image.Point{}, draw.Src)
		}

		draw.Draw(accumulated, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
		frameImg := image.NewRGBA(bounds)
		draw.Draw(frameImg, bounds, accumulated, bounds.Min, draw.Src)
		animator.frames[i] = frameImg
	}

	if device != nil {
		for i, frame := range animator.frames {
			texture, err := animator.createTextureFromImage(frame)
			if err != nil {
				for j := 0; j < i; j++ {
					if animator.textures[j] != nil {
						animator.textures[j].Release()
					}
				}
				return nil, err
			}
			animator.textures[i] = texture
		}
	}

	return animator, nil
}

func loadGPUStaticImage(device *d3d9.Device, imageBytes []byte) (*Animator, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}

	animator := &Animator{
		device:         device,
		staticImage:    img,
		isAnimated:     false,
		done:           make(chan bool),
		isPreprocessed: false,
		bounds:         img.Bounds(),
		needsUpdate:    true,
	}

	if device != nil {
		texture, err := animator.createTextureFromImage(img)
		if err != nil {
			return nil, err
		}
		animator.staticTexture = texture
	}

	return animator, nil
}

func (a *Animator) SetProcessor(processFunc func(*Render, image.Image) image.Image, render *Render) {
	a.processFunc = processFunc
	a.render = render
	a.needsUpdate = true
}

func (a *Animator) SetDevice(device *d3d9.Device, do func(*Render, image.Image) image.Image, r *Render) {
	a.device = device
	a.processFunc = do
	a.render = r
	a.needsUpdate = true

	a.cleanupTextures()
	a.cleanupProcessedTextures()

	if !a.isAnimated {
		if a.staticImage != nil {
			texture, err := a.createTextureFromImage(a.staticImage)
			if err == nil {
				a.staticTexture = texture
			}
		}
	} else {
		a.textures = make([]*d3d9.Texture, len(a.frames))
		a.processedTextures = make([]*d3d9.Texture, len(a.frames))
		for i, frame := range a.frames {
			texture, err := a.createTextureFromImage(frame)
			if err == nil {
				a.textures[i] = texture
			}
		}
	}
}

func (a *Animator) createTextureFromImage(img image.Image) (*d3d9.Texture, error) {
	if a.device == nil {
		return nil, nil
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	texture, err := a.device.CreateTexture(
		uint(width),
		uint(height),
		1,
		0,
		d3d9.FMT_A8R8G8B8,
		d3d9.POOL_MANAGED,
		0,
	)
	if err != nil {
		return nil, err
	}

	lockedRect, err := texture.LockRect(0, nil, 0)
	if err != nil {
		texture.Release()
		return nil, err
	}

	pitch := lockedRect.Pitch
	pBits := lockedRect.PBits

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcOffset := (y*rgba.Stride + x*4)
			dstOffset := y*int(pitch) + x*4

			if srcOffset+3 < len(rgba.Pix) && dstOffset+3 < int(lockedRect.Pitch)*height {
				r := rgba.Pix[srcOffset+0]
				g := rgba.Pix[srcOffset+1]
				b := rgba.Pix[srcOffset+2]
				a := rgba.Pix[srcOffset+3]

				size := width * height * 4
				pixels := unsafe.Slice((*byte)(unsafe.Pointer(pBits)), size)
				pixels[dstOffset+0] = b
				pixels[dstOffset+1] = g
				pixels[dstOffset+2] = r
				pixels[dstOffset+3] = a
			}
		}
	}

	texture.UnlockRect(0)
	return texture, nil
}

func (a *Animator) Preprocess(s *Render) {
	if a.isPreprocessed {
		return
	}
	a.cleanupTextures()

	if !a.isAnimated {
		if a.staticImage != nil {
			if a.device != nil {
				texture, err := a.createTextureFromImage(a.staticImage)
				if err == nil {
					a.staticTexture = texture
				}
			}
		}
	} else {
		a.textures = make([]*d3d9.Texture, len(a.frames))
		for i, frame := range a.frames {
			a.frames[i] = frame

			if a.device != nil {
				texture, err := a.createTextureFromImage(frame)
				if err == nil {
					a.textures[i] = texture
				}
			}
		}
	}

	a.isPreprocessed = true
}

func (a *Animator) GetCurrentImage(s *Render) image.Image {
	if !a.isPreprocessed {
		a.Preprocess(s)
	}

	var currentImg image.Image
	if !a.isAnimated {
		currentImg = a.staticImage
	} else {
		if len(a.frames) == 0 {
			return nil
		}
		currentImg = a.frames[a.currentFrame]
	}

	if a.processFunc != nil {
		renderToUse := s
		if renderToUse == nil {
			renderToUse = a.render
		}
		if renderToUse != nil {
			return a.processFunc(renderToUse, currentImg)
		}
	}

	return currentImg
}

func (a *Animator) GetCurrentTexture() *d3d9.Texture {
	if a.processFunc != nil && a.render != nil {
		return a.getProcessedTexture()
	}

	if !a.isAnimated {
		return a.staticTexture
	}

	if len(a.textures) == 0 || a.currentFrame >= len(a.textures) {
		return nil
	}

	return a.textures[a.currentFrame]
}

func (a *Animator) getProcessedTexture() *d3d9.Texture {
	if a.device == nil || a.processFunc == nil || a.render == nil {
		return nil
	}

	var originalImg image.Image
	if !a.isAnimated {
		originalImg = a.staticImage
	} else {
		if len(a.frames) == 0 || a.currentFrame >= len(a.frames) {
			return nil
		}
		originalImg = a.frames[a.currentFrame]
	}

	if originalImg == nil {
		return nil
	}

	processedImg := a.processFunc(a.render, originalImg)
	if processedImg == nil {
		return nil
	}

	texture, err := a.createTextureFromImage(processedImg)
	if err != nil {
		return nil
	}

	if !a.isAnimated {
		if a.processedStatic != nil {
			a.processedStatic.Release()
		}
		a.processedStatic = texture
	} else {
		if a.processedTextures[a.currentFrame] != nil {
			a.processedTextures[a.currentFrame].Release()
		}
		a.processedTextures[a.currentFrame] = texture
	}

	a.processedBounds = processedImg.Bounds()

	return texture
}

func (a *Animator) GetCurrentBounds() image.Rectangle {
	if a.processFunc != nil && !a.processedBounds.Empty() {
		return a.processedBounds
	}
	return a.bounds
}

func (a *Animator) GetCurrentImageRaw() image.Image {
	if !a.isAnimated {
		return a.staticImage
	}

	if len(a.frames) == 0 {
		return nil
	}

	return a.frames[a.currentFrame]
}

func (a *Animator) NextFrame() {
	if a.isAnimated && len(a.frames) > 1 {
		a.currentFrame = (a.currentFrame + 1) % len(a.frames)
		a.needsUpdate = true
	}
}

func (a *Animator) Start() {
	if !a.isAnimated || len(a.frames) <= 1 {
		return
	}

	go func() {
		for {
			select {
			case <-a.done:
				return
			default:
				delay := time.Duration(a.delays[a.currentFrame]) * time.Millisecond
				time.Sleep(delay)
				a.NextFrame()
				if a.hwnd != 0 {
					constants.ProcInvalidateRect.Call(uintptr(a.hwnd), 0, 1)
				}
			}
		}
	}()
}

func (a *Animator) Stop() {
	if a.done != nil {
		close(a.done)
	}
}

func (a *Animator) cleanupTextures() {
	if a.staticTexture != nil {
		a.staticTexture.Release()
		a.staticTexture = nil
	}

	for i, texture := range a.textures {
		if texture != nil {
			texture.Release()
			a.textures[i] = nil
		}
	}
}

func (a *Animator) cleanupProcessedTextures() {
	if a.processedStatic != nil {
		a.processedStatic.Release()
		a.processedStatic = nil
	}

	for i, texture := range a.processedTextures {
		if texture != nil {
			texture.Release()
			a.processedTextures[i] = nil
		}
	}
}

func (a *Animator) Cleanup() {
	a.Stop()
	a.cleanupTextures()
	a.cleanupProcessedTextures()
}

func (a *Animator) IsAnimated() bool {
	return a.isAnimated
}

func (a *Animator) IsPreprocessed() bool {
	return a.isPreprocessed
}

func (a *Animator) ResetPreprocessing() {
	a.isPreprocessed = false
	a.needsUpdate = true
}

func (a *Animator) HasProcessor() bool {
	return a.processFunc != nil && a.render != nil
}

func (a *Animator) RemoveProcessor() {
	a.processFunc = nil
	a.render = nil
	a.cleanupProcessedTextures()
}
