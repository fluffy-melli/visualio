package windows

import (
	"image"
	"image/color"
	"image/draw"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Screen struct {
	animator *Animator
	hwnd     windows.HWND
	Do       func(image.Image) image.Image
	X, Y     int
}

func NewScreen() *Screen {
	return &Screen{}
}

func (s *Screen) GetModuleHandle() windows.Handle {
	ret, _, _ := procGetModuleHandle.Call(0)
	return windows.Handle(ret)
}

func (s *Screen) GetScreenSize() (int, int) {
	width, _, _ := procGetSystemMetrics.Call(0)
	height, _, _ := procGetSystemMetrics.Call(1)
	return int(width), int(height)
}

func (s *Screen) WndProc(hwnd windows.HWND, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		s.Paint(s.X, s.Y)
		return 0
	case WM_KEYDOWN:
		if wParam == VK_ESCAPE {
			procPostQuitMessage.Call(0)
		}
		return 0
	case WM_RBUTTONDOWN, WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := procDefWindowProc.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func (s *Screen) Paint(x, y int) {
	if s.animator == nil {
		return
	}

	currentImage := s.animator.GetCurrentImage(s.Do)

	if currentImage == nil {
		return
	}

	bounds := currentImage.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	draw.Draw(rgba, bounds, currentImage, bounds.Min, draw.Over)
	hdc, _, _ := procGetWindowDC.Call(uintptr(s.hwnd))
	defer procReleaseDC.Call(uintptr(s.hwnd), hdc)
	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	defer procDeleteDC.Call(memDC)

	bi := BITMAPINFO{
		bmiHeader: BITMAPINFOHEADER{
			biSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
			biWidth:       int32(bounds.Dx()),
			biHeight:      -int32(bounds.Dy()),
			biPlanes:      1,
			biBitCount:    32,
			biCompression: 0,
		},
	}

	var bits uintptr
	hBitmap, _, _ := procCreateDIBSection.Call(hdc, uintptr(unsafe.Pointer(&bi)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	defer procDeleteObject.Call(hBitmap)

	if bits != 0 {
		pixelData := (*[1 << 30]byte)(unsafe.Pointer(bits))[:len(rgba.Pix)]
		for i := 0; i < len(rgba.Pix); i += 4 {
			a := rgba.Pix[i+3]
			if a == 0 {
				pixelData[i] = 0
				pixelData[i+1] = 0
				pixelData[i+2] = 0
				pixelData[i+3] = 0
			} else {
				pixelData[i] = rgba.Pix[i+2]
				pixelData[i+1] = rgba.Pix[i+1]
				pixelData[i+2] = rgba.Pix[i]
				pixelData[i+3] = rgba.Pix[i+3]
			}
		}
	}

	oldBitmap, _, _ := procSelectObject.Call(memDC, hBitmap)
	defer procSelectObject.Call(memDC, oldBitmap)

	sx, sy := s.GetScreenSize()
	if x < 0 {
		x = sx + x
	}

	if y < 0 {
		y = sy + y
	}

	procBitBlt.Call(hdc, uintptr(x), uintptr(y), uintptr(bounds.Dx()), uintptr(bounds.Dy()), memDC, 0, 0, SRCCOPY)
}

func (s *Screen) Create(className, imagePath string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var err error
	s.animator, err = NewGifAnimator(s.hwnd, imagePath)
	if err != nil {
		return err
	}
	hInstance := s.GetModuleHandle()
	classNamePtr, _ := windows.UTF16PtrFromString(className)
	wc := WNDCLASS{
		lpfnWndProc:   syscall.NewCallback(s.WndProc),
		hInstance:     hInstance,
		lpszClassName: classNamePtr,
	}
	procRegisterClass.Call(uintptr(unsafe.Pointer(&wc)))
	screenWidth, screenHeight := s.GetScreenSize()
	ret, _, _ := procCreateWindowEx.Call(
		uintptr(WS_EX_TOPMOST|WS_EX_LAYERED|WS_EX_TRANSPARENT|WS_EX_NOACTIVATE),
		uintptr(unsafe.Pointer(classNamePtr)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Overlay"))),
		uintptr(WS_POPUP|WS_VISIBLE),
		0, 0,
		uintptr(screenWidth), uintptr(screenHeight),
		0, 0, uintptr(hInstance), 0,
	)
	s.hwnd = windows.HWND(ret)
	procSetWindowPos.Call(uintptr(s.hwnd), ^uintptr(0), 0, 0, 0, 0, 0x0001|0x0002|0x0010)
	procSetLayeredWindowAttributes.Call(uintptr(s.hwnd), TRANSPARENT_COLOR, 255, LWA_COLORKEY|LWA_ALPHA)
	procShowWindow.Call(uintptr(s.hwnd), SW_SHOW)
	procUpdateWindow.Call(uintptr(s.hwnd))
	s.animator.Start()
	var msg MSG
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 || ret == ^uintptr(0) {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
	s.animator.Stop()
	return nil
}
