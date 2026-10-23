//go:build windows

package ole

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

// These tests only use the test thread's own message queue and a message-only
// window owned by the test process. Nothing outside the process is touched.

var (
	modUser32Test        = windows.NewLazySystemDLL("user32.dll")
	procPeekMessageW     = modUser32Test.NewProc("PeekMessageW")
	procPostMessageW     = modUser32Test.NewProc("PostMessageW")
	procPostQuitMessage  = modUser32Test.NewProc("PostQuitMessage")
	procRegisterClassExW = modUser32Test.NewProc("RegisterClassExW")
	procCreateWindowExW  = modUser32Test.NewProc("CreateWindowExW")
	procDestroyWindow    = modUser32Test.NewProc("DestroyWindow")
	procDefWindowProcW   = modUser32Test.NewProc("DefWindowProcW")
)

const (
	wmUserTest  = 0x0400 + 7 // WM_USER + 7
	wmQuit      = 0x0012
	pmNoRemove  = 0x0000
	hwndMessage = ^uintptr(2) // HWND_MESSAGE (-3)
)

// forceMessageQueue makes sure the calling thread has a message queue.
func forceMessageQueue(t *testing.T) {
	t.Helper()
	var scratch [64]byte
	procPeekMessageW.Call(uintptr(unsafe.Pointer(&scratch[0])), 0, 0, 0, pmNoRemove)
}

func TestGetMessage_ReadsPostedThreadMessage(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	forceMessageQueue(t)

	r, _, errno := procPostMessageW.Call(0, wmUserTest, 0x1234, 0x5678)
	require.NotZero(t, r, "PostMessage: %v", errno)

	var msg Msg
	ret, err := GetMessage(&msg, 0, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int32(1), ret)
	assert.Equal(t, uint32(wmUserTest), msg.Message)
	assert.Equal(t, int32(0x1234), msg.Wparam)
	assert.Equal(t, int32(0x5678), msg.Lparam)
}

func TestGetMessage_QuitReturnsZeroWithoutError(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	forceMessageQueue(t)

	procPostQuitMessage.Call(0)

	var msg Msg
	ret, err := GetMessage(&msg, 0, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int32(0), ret)
	assert.Equal(t, uint32(wmQuit), msg.Message)
}

// GetMessage must not write past the exported Msg.
func TestGetMessage_DoesNotWritePastMsg(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	forceMessageQueue(t)

	r, _, errno := procPostMessageW.Call(0, wmUserTest, 1, 2)
	require.NotZero(t, r, "PostMessage: %v", errno)

	var guarded struct {
		msg   Msg
		guard [64]byte
	}
	for i := range guarded.guard {
		guarded.guard[i] = 0xAA
	}

	GetMessage(&guarded.msg, 0, 0, 0) // the error result is covered by the other tests
	for i, b := range guarded.guard {
		require.Equal(t, byte(0xAA), b, "guard byte %d overwritten", i)
	}
}

var dispatchedMessages []dispatchedMessage

type dispatchedMessage struct {
	msg    uint32
	wparam uintptr
	lparam uintptr
}

func testWndProc(hwnd, msg, wparam, lparam uintptr) uintptr {
	if msg == wmUserTest {
		dispatchedMessages = append(dispatchedMessages, dispatchedMessage{uint32(msg), wparam, lparam})
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, msg, wparam, lparam)
	return r
}

type wndClassExW struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

var (
	registerWndClassOnce sync.Once
	wndClassName         *uint16
	wndClassErr          error
)

// registerTestWndClass registers the window class and callback once per
// process, so repeated runs (-test.count) neither collide on the class name nor
// consume further callback slots.
func registerTestWndClass() error {
	registerWndClassOnce.Do(func() {
		wndClassName, wndClassErr = windows.UTF16PtrFromString("GoOleTestMessageOnly")
		if wndClassErr != nil {
			return
		}
		wc := wndClassExW{
			WndProc:   syscall.NewCallback(testWndProc),
			ClassName: wndClassName,
		}
		wc.Size = uint32(unsafe.Sizeof(wc))
		atom, _, errno := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
		if atom == 0 {
			wndClassErr = fmt.Errorf("RegisterClassEx: %w", errno)
		}
	})
	return wndClassErr
}

// DispatchMessage must hand the window procedure the message that GetMessage
// read, through a correctly laid out MSG: Wparam zero-extended, Lparam
// sign-extended.
func TestDispatchMessage_DeliversMessageToWindowProc(t *testing.T) {
	tests := []struct {
		name       string
		wparam     uintptr
		lparam     uintptr
		wantWparam uintptr
		wantLparam uintptr
	}{
		{"plain values", 0x1234, 0x5678, 0x1234, 0x5678},
		{"Lparam -1 is sign-extended, Wparam high bit is not", 0x80000000, ^uintptr(0), 0x80000000, ^uintptr(0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			dispatchedMessages = nil
			require.NoError(t, registerTestWndClass())

			hwnd, _, errno := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(wndClassName)), 0, 0, 0, 0, 0, 0, hwndMessage, 0, 0, 0)
			require.NotZero(t, hwnd, "CreateWindowEx: %v", errno)
			defer procDestroyWindow.Call(hwnd)

			// Post the 32-bit patterns; a 32-bit Msg cannot carry more.
			r, _, errno := procPostMessageW.Call(hwnd, wmUserTest, tt.wparam, uintptr(uint32(tt.lparam)))
			require.NotZero(t, r, "PostMessage: %v", errno)

			var msg Msg
			GetMessage(&msg, 0, 0, 0) // the error result is covered by the other tests
			require.Equal(t, uint32(wmUserTest), msg.Message)

			DispatchMessage(&msg)

			require.Len(t, dispatchedMessages, 1)
			assert.Equal(t, uint32(wmUserTest), dispatchedMessages[0].msg)
			assert.Equal(t, tt.wantWparam, dispatchedMessages[0].wparam)
			assert.Equal(t, tt.wantLparam, dispatchedMessages[0].lparam)
		})
	}
}

// hWnd -1 is the documented "thread messages only" filter. The uint32 hwnd
// argument carries it as 0xFFFFFFFF, which must be sign-extended to a native
// -1 on 64-bit Windows.
func TestGetMessage_ThreadMessagesFilter(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	forceMessageQueue(t)

	r, _, errno := procPostMessageW.Call(0, wmUserTest, 9, 10)
	require.NotZero(t, r, "PostMessage: %v", errno)

	var msg Msg
	ret, err := GetMessage(&msg, 0xFFFFFFFF, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int32(1), ret)
	assert.Equal(t, uint32(wmUserTest), msg.Message)
	assert.Equal(t, int32(9), msg.Wparam)
}

func TestVariantInit_ReturnsNilError(t *testing.T) {
	failures := 0
	for i := 0; i < 1000; i++ {
		v := VARIANT{VT: VT_I4, Val: 7}
		if err := VariantInit(&v); err != nil {
			failures++
		}
		require.Equal(t, VT_EMPTY, v.VT)
	}
	assert.Zero(t, failures, "VariantInit returned an error on %d of 1000 calls", failures)
}

func TestSysFreeString_ReturnsNilError(t *testing.T) {
	failures := 0
	for i := 0; i < 1000; i++ {
		if err := SysFreeString(SysAllocString("hello")); err != nil {
			failures++
		}
	}
	assert.Zero(t, failures, "SysFreeString returned an error on %d of 1000 calls", failures)
}
