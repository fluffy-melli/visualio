package cursor

import (
	"context"
	"time"
	"unsafe"

	"github.com/fluffy-melli/visualio/constants"
	"github.com/fluffy-melli/visualio/graphics"
)

type Location struct {
	X int32
	Y int32
}

func Position() (Location, error) {
	var pt Location
	ret, _, err := constants.ProcGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ret == 0 {
		return pt, err
	}
	return pt, nil
}

func PositionReader(ctx context.Context, out chan<- Location) func(s *graphics.Render) {
	return func(s *graphics.Render) {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pos, err := Position()
				if err != nil {
					panic(err)
				}
				out <- pos
			}
		}
	}
}

func DeltaHandler(ctx context.Context, in <-chan Location) func(s *graphics.Render) {
	return func(s *graphics.Render) {
		var last Location
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
