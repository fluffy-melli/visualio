package constants

import "golang.org/x/sys/windows"

const (
	WM_DESTROY     = 0x0002
	WM_PAINT       = 0x000F
	WM_KEYDOWN     = 0x0100
	WM_RBUTTONDOWN = 0x0204

	WS_POPUP          = 0x80000000
	WS_VISIBLE        = 0x10000000
	WS_EX_TOPMOST     = 0x00000008
	WS_EX_LAYERED     = 0x00080000
	WS_EX_TRANSPARENT = 0x00000020
	WS_EX_NOACTIVATE  = 0x08000000
	WM_LBUTTONDOWN    = 0x0201
	WM_LBUTTONUP      = 0x0202
	WM_MBUTTONDOWN    = 0x0207
	WM_MBUTTONUP      = 0x0208

	SW_SHOW   = 5
	VK_ESCAPE = 0x1B

	LWA_ALPHA    = 0x00000002
	LWA_COLORKEY = 0x00000001

	SRCCOPY = 0x00CC0020

	TRANSPARENT_COLOR = 0x00000000
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	ProcRegisterClass              = user32.NewProc("RegisterClassW")
	ProcCreateWindowEx             = user32.NewProc("CreateWindowExW")
	ProcDefWindowProc              = user32.NewProc("DefWindowProcW")
	ProcGetMessage                 = user32.NewProc("GetMessageW")
	ProcTranslateMessage           = user32.NewProc("TranslateMessage")
	ProcDispatchMessage            = user32.NewProc("DispatchMessageW")
	ProcPostQuitMessage            = user32.NewProc("PostQuitMessage")
	ProcShowWindow                 = user32.NewProc("ShowWindow")
	ProcUpdateWindow               = user32.NewProc("UpdateWindow")
	ProcGetWindowDC                = user32.NewProc("GetWindowDC")
	ProcReleaseDC                  = user32.NewProc("ReleaseDC")
	ProcGetModuleHandle            = kernel32.NewProc("GetModuleHandleW")
	ProcSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	ProcSetWindowPos               = user32.NewProc("SetWindowPos")
	ProcGetSystemMetrics           = user32.NewProc("GetSystemMetrics")
	ProcInvalidateRect             = user32.NewProc("InvalidateRect")
	ProcCreateSolidBrush           = gdi32.NewProc("CreateSolidBrush")

	ProcCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	ProcCreateDIBSection   = gdi32.NewProc("CreateDIBSection")
	ProcSelectObject       = gdi32.NewProc("SelectObject")
	ProcBitBlt             = gdi32.NewProc("BitBlt")
	ProcDeleteDC           = gdi32.NewProc("DeleteDC")
	ProcDeleteObject       = gdi32.NewProc("DeleteObject")
	ProcGetCursorPos       = user32.NewProc("GetCursorPos")
	ProcFillRect           = user32.NewProc("FillRect")
)

var (
	ProcSetTimer  = user32.NewProc("SetTimer")
	ProcKillTimer = user32.NewProc("KillTimer")
)

const (
	WM_TIMER = 0x0113
)
