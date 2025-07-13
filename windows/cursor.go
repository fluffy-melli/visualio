package windows

import (
	"context"
	"time"
	"unsafe"
)

type Point struct {
	X int32
	Y int32
}

func CursorPosition() (Point, error) {
	var pt Point
	ret, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ret == 0 {
		return pt, err
	}
	return pt, nil
}

func CursorPositionReader(ctx context.Context, out chan<- Point) func(s *Screen) {
	return func(s *Screen) {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pos, err := CursorPosition()
				if err != nil {
					panic(err)
				}
				out <- pos
			}
		}
	}
}

func CursorDeltaHandler(ctx context.Context, in <-chan Point) func(s *Screen) {
	return func(s *Screen) {
		var last Point
		for {
			select {
			case <-ctx.Done():
				return
			case pos := <-in:
				if pos != last {
					img := s.CurrentImage()
					if img != nil {
						bounds := img.Bounds()
						minX, minY := s.AX, s.AY
						maxX, maxY := minX+bounds.Dx(), minY+bounds.Dy()
						s.IsInside = int(pos.X) >= minX && int(pos.X) < maxX &&
							int(pos.Y) >= minY && int(pos.Y) < maxY
					}
					if s.IsClicked {
						dx, dy := int(pos.X-last.X), int(pos.Y-last.Y)
						s.AX += dx
						s.AY += dy
					}
					last = pos
				}
			}
		}
	}
}
