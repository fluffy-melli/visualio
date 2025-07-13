package windows

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

type Animator struct {
	hwnd         windows.HWND
	frames       []image.Image
	delays       []int
	currentFrame int
	done         chan bool
	isAnimated   bool
	staticImage  image.Image
}

func NewGifAnimator(hwnd windows.HWND, imagePath string) (*Animator, error) {
	imageBytes, err := os.ReadFile(imagePath)

	if err != nil {
		return nil, err
	}

	if len(imageBytes) > 3 && string(imageBytes[:3]) == "GIF" {
		return loadGifAnimation(hwnd, imageBytes)
	}

	return loadStaticImage(hwnd, imageBytes)
}

func loadGifAnimation(hwnd windows.HWND, imageBytes []byte) (*Animator, error) {
	gifImg, err := gif.DecodeAll(bytes.NewReader(imageBytes))

	if err != nil {
		return nil, err
	}

	animator := &Animator{
		hwnd:         hwnd,
		frames:       make([]image.Image, len(gifImg.Image)),
		delays:       make([]int, len(gifImg.Image)),
		currentFrame: 0,
		done:         make(chan bool),
		isAnimated:   true,
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
	return animator, nil
}

func loadStaticImage(hwnd windows.HWND, imageBytes []byte) (*Animator, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))

	if err != nil {
		return nil, err
	}

	return &Animator{
		hwnd:        hwnd,
		staticImage: img,
		isAnimated:  false,
		done:        make(chan bool),
	}, nil
}

func (a *Animator) GetCurrentImage(do func(image.Image) image.Image) image.Image {
	if !a.isAnimated {
		return do(a.staticImage)
	}

	if len(a.frames) == 0 {
		return nil
	}

	return do(a.frames[a.currentFrame])
}

func (a *Animator) NextFrame() {
	if a.isAnimated && len(a.frames) > 1 {
		a.currentFrame = (a.currentFrame + 1) % len(a.frames)
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
					procInvalidateRect.Call(uintptr(a.hwnd), 0, 1)
				}
			}
		}
	}()
}

func (a *Animator) Stop() {
	close(a.done)
}

func (a *Animator) IsAnimated() bool {
	return a.isAnimated
}
