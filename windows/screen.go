package windows

import (
	"image"
	"image/draw"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type Screen struct {
	animator   *Animator
	window     windows.HWND
	ProcessImg func(*Screen, image.Image) image.Image
	Routines   []func(*Screen)
	AX, AY     int
	IsInside   bool
	IsClicked  bool
}

type Rect struct {
	Left, Top, Right, Bottom int32
}

func NewScreen() *Screen {
	return &Screen{}
}

func (s *Screen) CurrentImage() image.Image {
	if s.animator == nil || s.ProcessImg == nil {
		return nil
	}
	return s.animator.GetCurrentImage(s, s.ProcessImg)
}

func (s *Screen) ModuleHandle() windows.Handle {
	handle, _, _ := procGetModuleHandle.Call(0)
	return windows.Handle(handle)
}

func (s *Screen) ScreenSize() (int, int) {
	width, _, _ := procGetSystemMetrics.Call(0)
	height, _, _ := procGetSystemMetrics.Call(1)
	return int(width), int(height)
}

func (s *Screen) WindowProc(hwnd windows.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		s.Render(s.AX, s.AY)
		return 0
	case WM_KEYDOWN:
		if wParam == VK_ESCAPE {
			procPostQuitMessage.Call(0)
		}
		return 0
	case WM_RBUTTONDOWN, WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	case WM_MBUTTONDOWN:
		s.IsClicked = true
		return 0
	case WM_MBUTTONUP:
		s.IsClicked = false
		return 0
	}
	ret, _, _ := procDefWindowProc.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func (s *Screen) ClearWindow() {
	if s.window == 0 {
		return
	}
	hdc, _, _ := procGetWindowDC.Call(uintptr(s.window))
	if hdc == 0 {
		return
	}
	defer procReleaseDC.Call(uintptr(s.window), hdc)

	width, height := s.ScreenSize()
	brush, _, _ := procCreateSolidBrush.Call(uintptr(TRANSPARENT_COLOR))
	if brush == 0 {
		return
	}
	defer procDeleteObject.Call(brush)

	rect := Rect{0, 0, int32(width), int32(height)}
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rect)), brush)
	procInvalidateRect.Call(uintptr(s.window), 0, 1)
}

func (s *Screen) Render(x, y int) {
	if s.animator == nil {
		return
	}
	img := s.CurrentImage()
	if img == nil {
		return
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Over)

	hdc, _, _ := procGetWindowDC.Call(uintptr(s.window))
	if hdc == 0 {
		return
	}
	defer procReleaseDC.Call(uintptr(s.window), hdc)

	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	if memDC == 0 {
		return
	}
	defer procDeleteDC.Call(memDC)

	bitmapInfo := BITMAPINFO{
		bmiHeader: BITMAPINFOHEADER{
			biSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
			biWidth:       int32(bounds.Dx()),
			biHeight:      -int32(bounds.Dy()),
			biPlanes:      1,
			biBitCount:    32,
			biCompression: 0,
		},
	}

	var pixelPtr uintptr
	hBitmap, _, _ := procCreateDIBSection.Call(hdc, uintptr(unsafe.Pointer(&bitmapInfo)), 0, uintptr(unsafe.Pointer(&pixelPtr)), 0, 0)
	if hBitmap == 0 || pixelPtr == 0 {
		return
	}
	defer procDeleteObject.Call(hBitmap)

	pixels := (*[1 << 30]byte)(unsafe.Pointer(pixelPtr))[:len(rgba.Pix)]
	for i := 0; i < len(rgba.Pix); i += 4 {
		a := rgba.Pix[i+3]
		if a == 0 {
			pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = 0, 0, 0, 0
		} else {
			pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] =
				rgba.Pix[i+2], rgba.Pix[i+1], rgba.Pix[i], rgba.Pix[i+3]
		}
	}

	if s.IsClicked {
		s.ClearWindow()
	}

	oldBitmap, _, _ := procSelectObject.Call(memDC, hBitmap)
	procBitBlt.Call(hdc, uintptr(x), uintptr(y), uintptr(bounds.Dx()), uintptr(bounds.Dy()), memDC, 0, 0, SRCCOPY)
	procSelectObject.Call(memDC, oldBitmap)
}

func (s *Screen) CreateWindow(className, imagePath string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var err error
	s.animator, err = NewGifAnimator(s.window, imagePath)
	if err != nil {
		return err
	}

	hInstance := s.ModuleHandle()
	classNamePtr, _ := windows.UTF16PtrFromString(className)
	wndClass := WNDCLASS{
		lpfnWndProc:   syscall.NewCallback(s.WindowProc),
		hInstance:     hInstance,
		lpszClassName: classNamePtr,
	}
	procRegisterClass.Call(uintptr(unsafe.Pointer(&wndClass)))

	width, height := s.ScreenSize()
	ret, _, _ := procCreateWindowEx.Call(
		uintptr(WS_EX_TOPMOST|WS_EX_LAYERED|WS_EX_NOACTIVATE),
		uintptr(unsafe.Pointer(classNamePtr)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Overlay"))),
		uintptr(WS_POPUP|WS_VISIBLE),
		0, 0,
		uintptr(width), uintptr(height),
		0, 0, uintptr(hInstance), 0,
	)
	s.window = windows.HWND(ret)

	procSetWindowPos.Call(uintptr(s.window), ^uintptr(0), 0, 0, 0, 0, 0x0001|0x0002|0x0010)
	procSetLayeredWindowAttributes.Call(uintptr(s.window), TRANSPARENT_COLOR, 255, LWA_COLORKEY|LWA_ALPHA)
	procShowWindow.Call(uintptr(s.window), SW_SHOW)
	procUpdateWindow.Call(uintptr(s.window))

	s.animator.Start()
	s.RunRoutines()

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

func (s *Screen) RunRoutines() {
	for _, routine := range s.Routines {
		go routine(s)
	}
}
