package graphics

import (
	"fmt"
	"image"
	"image/draw"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/fluffy-melli/visualio/constants"
	"github.com/gonutz/d3d9"
	"golang.org/x/sys/windows"
)

type WNDCLASS struct {
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     windows.Handle
	hIcon         windows.Handle
	hCursor       windows.Handle
	hbrBackground windows.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
}

type MSG struct {
	hwnd    windows.HWND
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct{ x, y int32 }
}

type BITMAPINFOHEADER struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

type BITMAPINFO struct {
	bmiHeader BITMAPINFOHEADER
	bmiColors [1]uint32
}

type Render struct {
	animator    *Animator
	window      windows.HWND
	ProcessImg  func(*Render, image.Image) image.Image
	Routines    []func(*Render)
	AX, AY      int
	IsInside    bool
	IsClicked   bool
	d3d9Obj     *d3d9.Direct3D
	device      *d3d9.Device
	texture     *d3d9.Texture
	initialized bool
}

type Rect struct {
	Left, Top, Right, Bottom int32
}

func NewScreen() *Render {
	return &Render{}
}

func (s *Render) CurrentImage() image.Image {
	if s.animator == nil || s.ProcessImg == nil {
		return nil
	}
	return s.animator.GetCurrentImage(s, s.ProcessImg)
}

func (s *Render) ModuleHandle() windows.Handle {
	handle, _, _ := constants.ProcGetModuleHandle.Call(0)
	return windows.Handle(handle)
}

func (s *Render) ScreenSize() (int, int) {
	width, _, _ := constants.ProcGetSystemMetrics.Call(0)
	height, _, _ := constants.ProcGetSystemMetrics.Call(1)
	return int(width), int(height)
}

func (s *Render) initD3D9() error {
	var err error

	s.d3d9Obj, err = d3d9.Create(d3d9.SDK_VERSION)
	if err != nil {
		return err
	}

	width, height := s.ScreenSize()

	pp := d3d9.PRESENT_PARAMETERS{
		Windowed:               1,
		SwapEffect:             d3d9.SWAPEFFECT_DISCARD,
		BackBufferFormat:       d3d9.FMT_UNKNOWN,
		BackBufferWidth:        uint32(width),
		BackBufferHeight:       uint32(height),
		HDeviceWindow:          d3d9.HWND(s.window),
		EnableAutoDepthStencil: 0,
		Flags:                  d3d9.PRESENTFLAG_LOCKABLE_BACKBUFFER,
	}

	s.device, _, err = s.d3d9Obj.CreateDevice(
		d3d9.ADAPTER_DEFAULT,
		d3d9.DEVTYPE_HAL,
		d3d9.HWND(s.window),
		d3d9.CREATE_HARDWARE_VERTEXPROCESSING,
		pp,
	)
	if err != nil {
		s.device, _, err = s.d3d9Obj.CreateDevice(
			d3d9.ADAPTER_DEFAULT,
			d3d9.DEVTYPE_HAL,
			d3d9.HWND(s.window),
			d3d9.CREATE_SOFTWARE_VERTEXPROCESSING,
			pp,
		)
		if err != nil {
			return err
		}
	}

	s.initialized = true
	return nil
}

func (s *Render) RenderGPU(x, y int) {
	if !s.initialized || s.animator == nil || s.device == nil {
		return
	}

	img := s.CurrentImage()
	if img == nil {
		return
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Over)

	if s.needsTextureUpdate(bounds) {
		s.updateTexture(rgba)
		if s.texture == nil {
			return
		}
	} else {
		s.updateTextureData(rgba)
	}

	deviceStatusErr := s.device.TestCooperativeLevel()

	if deviceStatusErr != nil {
		if deviceStatusErr.Code() == d3d9.ERR_DEVICENOTRESET {
			width, height := s.ScreenSize()
			pp := d3d9.PRESENT_PARAMETERS{
				Windowed:               1,
				SwapEffect:             d3d9.SWAPEFFECT_DISCARD,
				BackBufferFormat:       d3d9.FMT_UNKNOWN,
				BackBufferWidth:        uint32(width),
				BackBufferHeight:       uint32(height),
				HDeviceWindow:          d3d9.HWND(s.window),
				EnableAutoDepthStencil: 0,
				Flags:                  d3d9.PRESENTFLAG_LOCKABLE_BACKBUFFER,
			}
			if _, err := s.device.Reset(pp); err != nil {
				fmt.Println("Device reset failed:", err)
				s.initialized = false
				return
			}
			fmt.Println("Device reset successfully.")
		} else {
			fmt.Println("Other D3D9 device error:", deviceStatusErr)
			return
		}
	}
	s.device.Clear(nil, d3d9.CLEAR_TARGET, d3d9.ColorRGBA(0, 0, 0, 0), 1.0, 0)
	if err := s.device.BeginScene(); err != nil {
		fmt.Println("Failed to begin scene:", err)
		return
	}

	s.renderTexturedQuad(x, y, bounds.Dx(), bounds.Dy())

	if err := s.device.EndScene(); err != nil {
		fmt.Println("Failed to end scene:", err)
		return
	}

	s.device.Present(nil, nil, 0, nil)
}

type CUSTOM_VERTEX struct {
	X, Y, Z float32
	Rhw     float32
	Color   uint32
	U, V    float32
}

const (
	D3DFVF_XYZRHW  = 0x004
	D3DFVF_DIFFUSE = 0x040
	D3DFVF_TEX1    = 0x100
	CUSTOM_FVF     = D3DFVF_XYZRHW | D3DFVF_DIFFUSE | D3DFVF_TEX1
)

func (s *Render) renderTexturedQuad(x, y, width, height int) {
	if s.texture == nil || s.device == nil {
		return
	}

	s.device.SetRenderState(d3d9.RS_CULLMODE, d3d9.CULL_NONE)
	s.device.SetRenderState(d3d9.RS_LIGHTING, 0)
	s.device.SetRenderState(d3d9.RS_ZENABLE, 0)
	s.device.SetRenderState(d3d9.RS_ALPHABLENDENABLE, 1)
	s.device.SetRenderState(d3d9.RS_SRCBLEND, d3d9.BLEND_SRCALPHA)
	s.device.SetRenderState(d3d9.RS_DESTBLEND, d3d9.BLEND_INVSRCALPHA)

	s.device.SetTextureStageState(0, d3d9.TSS_COLOROP, d3d9.TOP_SELECTARG1)
	s.device.SetTextureStageState(0, d3d9.TSS_COLORARG1, d3d9.TA_TEXTURE)
	s.device.SetTextureStageState(0, d3d9.TSS_ALPHAOP, d3d9.TOP_MODULATE)
	s.device.SetTextureStageState(0, d3d9.TSS_ALPHAARG1, d3d9.TA_TEXTURE)
	s.device.SetTextureStageState(0, d3d9.TSS_ALPHAARG2, d3d9.TA_DIFFUSE)
	s.device.SetSamplerState(0, d3d9.SAMP_MINFILTER, d3d9.TEXF_LINEAR)
	s.device.SetSamplerState(0, d3d9.SAMP_MAGFILTER, d3d9.TEXF_LINEAR)

	s.device.SetTexture(0, s.texture)

	vertices := []CUSTOM_VERTEX{
		{X: float32(x), Y: float32(y), Z: 0.0, Rhw: 1.0, Color: 0xFFFFFFFF, U: 0.0, V: 0.0},
		{X: float32(x + width), Y: float32(y), Z: 0.0, Rhw: 1.0, Color: 0xFFFFFFFF, U: 1.0, V: 0.0},
		{X: float32(x), Y: float32(y + height), Z: 0.0, Rhw: 1.0, Color: 0xFFFFFFFF, U: 0.0, V: 1.0},
		{X: float32(x + width), Y: float32(y + height), Z: 0.0, Rhw: 1.0, Color: 0xFFFFFFFF, U: 1.0, V: 1.0},
	}

	s.device.SetFVF(CUSTOM_FVF)

	s.device.DrawPrimitiveUP(
		d3d9.PT_TRIANGLESTRIP,
		2,
		uintptr(unsafe.Pointer(&vertices[0])),
		uint(unsafe.Sizeof(CUSTOM_VERTEX{})),
	)
}

func (s *Render) needsTextureUpdate(bounds image.Rectangle) bool {
	if s.texture == nil {
		return true
	}
	desc, err := s.texture.GetLevelDesc(0)
	if err != nil {
		return true
	}
	return desc.Width != uint32(bounds.Dx()) || desc.Height != uint32(bounds.Dy())
}

func (s *Render) updateTexture(rgba *image.RGBA) {
	if s.texture != nil {
		s.texture.Release()
	}

	bounds := rgba.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	var err error
	s.texture, err = s.device.CreateTexture(
		uint(width),
		uint(height),
		1,
		d3d9.USAGE_DYNAMIC,
		d3d9.FMT_A8R8G8B8,
		d3d9.POOL_DEFAULT,
		0,
	)
	if err != nil {
		return
	}

	s.updateTextureData(rgba)
}

func (s *Render) updateTextureData(rgba *image.RGBA) {
	if s.texture == nil {
		return
	}

	bounds := rgba.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	lockedRect, err := s.texture.LockRect(0, nil, d3d9.LOCK_DISCARD)
	if err != nil {
		return
	}

	pitch := int(lockedRect.Pitch)
	pixels := (*[1 << 30]byte)(unsafe.Pointer(lockedRect.PBits))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcIdx := (y*width + x) * 4
			dstIdx := y*pitch + x*4

			if srcIdx+3 < len(rgba.Pix) && dstIdx+3 < len(pixels) {
				a := rgba.Pix[srcIdx+3]
				r := rgba.Pix[srcIdx+0]
				g := rgba.Pix[srcIdx+1]
				b := rgba.Pix[srcIdx+2]

				if a == 0 {
					pixels[dstIdx+0] = 0
					pixels[dstIdx+1] = 0
					pixels[dstIdx+2] = 0
					pixels[dstIdx+3] = 0
				} else {
					pixels[dstIdx+0] = b
					pixels[dstIdx+1] = g
					pixels[dstIdx+2] = r
					pixels[dstIdx+3] = a
				}
			}
		}
	}

	s.texture.UnlockRect(0)
}

func (s *Render) WindowProc(hwnd windows.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case constants.WM_PAINT:
		if s.initialized {
			s.RenderGPU(s.AX, s.AY)
		} else {
			s.Render(s.AX, s.AY)
		}
		return 0
	case constants.WM_KEYDOWN:
		if wParam == constants.VK_ESCAPE {
			constants.ProcPostQuitMessage.Call(0)
		}
		return 0
	case constants.WM_DESTROY:
		s.cleanup()
		constants.ProcPostQuitMessage.Call(0)
		return 0
	case constants.WM_RBUTTONDOWN:
		if s.IsClicked {
			constants.ProcPostQuitMessage.Call(0)
		}
		return 0
	case constants.WM_MBUTTONDOWN:
		s.IsClicked = true
		return 0
	case constants.WM_MBUTTONUP:
		s.IsClicked = false
		return 0
	}
	ret, _, _ := constants.ProcDefWindowProc.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func (s *Render) ClearWindow() {
	if s.initialized && s.device != nil {
		s.device.Clear(nil, d3d9.CLEAR_TARGET, d3d9.ColorRGBA(0, 0, 0, 0), 1.0, 0)
		s.device.Present(nil, nil, 0, nil)
	} else {
		s.clearWindowGDI()
	}
}

func (s *Render) clearWindowGDI() {
	if s.window == 0 {
		return
	}

	hdc, _, _ := constants.ProcGetWindowDC.Call(uintptr(s.window))
	if hdc == 0 {
		return
	}
	defer constants.ProcReleaseDC.Call(uintptr(s.window), hdc)

	width, height := s.ScreenSize()
	brush, _, _ := constants.ProcCreateSolidBrush.Call(uintptr(constants.TRANSPARENT_COLOR))
	if brush == 0 {
		return
	}
	defer constants.ProcDeleteObject.Call(brush)

	rect := Rect{0, 0, int32(width), int32(height)}
	constants.ProcFillRect.Call(hdc, uintptr(unsafe.Pointer(&rect)), brush)
	constants.ProcInvalidateRect.Call(uintptr(s.window), 0, 1)
}

func (s *Render) Render(x, y int) {
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

	hdc, _, _ := constants.ProcGetWindowDC.Call(uintptr(s.window))
	if hdc == 0 {
		return
	}
	defer constants.ProcReleaseDC.Call(uintptr(s.window), hdc)

	memDC, _, _ := constants.ProcCreateCompatibleDC.Call(hdc)
	if memDC == 0 {
		return
	}
	defer constants.ProcDeleteDC.Call(memDC)

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
	hBitmap, _, _ := constants.ProcCreateDIBSection.Call(hdc, uintptr(unsafe.Pointer(&bitmapInfo)), 0, uintptr(unsafe.Pointer(&pixelPtr)), 0, 0)
	if hBitmap == 0 || pixelPtr == 0 {
		return
	}
	defer constants.ProcDeleteObject.Call(hBitmap)

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

	oldBitmap, _, _ := constants.ProcSelectObject.Call(memDC, hBitmap)
	constants.ProcBitBlt.Call(hdc, uintptr(x), uintptr(y), uintptr(bounds.Dx()), uintptr(bounds.Dy()), memDC, 0, 0, constants.SRCCOPY)
	constants.ProcSelectObject.Call(memDC, oldBitmap)
}

func (s *Render) CreateWindow(className, imagePath string) error {
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

	constants.ProcRegisterClass.Call(uintptr(unsafe.Pointer(&wndClass)))

	width, height := s.ScreenSize()
	ret, _, _ := constants.ProcCreateWindowEx.Call(
		uintptr(constants.WS_EX_TOPMOST|constants.WS_EX_LAYERED|constants.WS_EX_NOACTIVATE),
		uintptr(unsafe.Pointer(classNamePtr)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Overlay"))),
		uintptr(constants.WS_POPUP|constants.WS_VISIBLE),
		0, 0,
		uintptr(width), uintptr(height),
		0, 0, uintptr(hInstance), 0,
	)

	s.window = windows.HWND(ret)

	if err := s.initD3D9(); err != nil {
		s.initialized = false
	}

	constants.ProcSetWindowPos.Call(uintptr(s.window), ^uintptr(0), 0, 0, 0, 0, 0x0001|0x0002|0x0010)
	constants.ProcSetLayeredWindowAttributes.Call(uintptr(s.window), constants.TRANSPARENT_COLOR, 255, constants.LWA_COLORKEY|constants.LWA_ALPHA)
	constants.ProcShowWindow.Call(uintptr(s.window), constants.SW_SHOW)
	constants.ProcUpdateWindow.Call(uintptr(s.window))

	s.animator.Start()
	s.RunRoutines()

	var msg MSG
	for {
		ret, _, _ := constants.ProcGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 || ret == ^uintptr(0) {
			break
		}
		constants.ProcTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		constants.ProcDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}

	s.animator.Stop()
	s.cleanup()
	return nil
}

func (s *Render) RunRoutines() {
	for _, routine := range s.Routines {
		go routine(s)
	}
}

func (s *Render) cleanup() {
	if s.texture != nil {
		s.texture.Release()
		s.texture = nil
	}
	if s.device != nil {
		s.device.Release()
		s.device = nil
	}
	if s.d3d9Obj != nil {
		s.d3d9Obj.Release()
		s.d3d9Obj = nil
	}
	s.initialized = false
}
