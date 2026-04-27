//go:build windows

package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32            = windows.NewLazySystemDLL("user32.dll")
	procGetForeground = user32.NewProc("GetForegroundWindow")
	procGetWindowText = user32.NewProc("GetWindowTextW")
	procGetMessage    = user32.NewProc("GetMessageW")
	procDispatchMsg   = user32.NewProc("DispatchMessageW")
	winEventHook      = user32.NewProc("SetWinEventHook")
)

const (
	EVENT_SYSTEM_FOREGROUND = 0x0003
	WINEVENT_OUTOFCONTEXT   = 0x0000
	WINEVENT_INCONTEXT      = 0x0001
	windowChangeDebounce    = 2 * time.Second
)

type Msg struct {
	hwnd     uintptr
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	pt       point
	lPrivate uint32
}

type point struct {
	x int32
	y int32
}

func getWindowTitle(hwnd uintptr) string {
	buf := make([]uint16, 256)
	procGetWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

var foregroundCheck = make(chan uintptr, 1)
var hookCallback uintptr
var hookHandle uintptr

func onWindowChange(hwnd uintptr) {
	debugPrintf("onWindowChange triggered for hwnd=%d\n", hwnd)
	select {
	case foregroundCheck <- hwnd:
	default:
		select {
		case <-foregroundCheck:
		default:
		}
		foregroundCheck <- hwnd
	}
}

func installWindowEventHook() error {
	debugPrintln("Initializing window event hook...")

	hookCallback = windows.NewCallback(func(eventHook, event, hwnd, idObject, childWnd, threadId, timestamp uintptr) uintptr {
		debugPrintf("!!! CALLBACK FIRED: event=%d, hwnd=%d !!!\n", event, hwnd)
		if event == EVENT_SYSTEM_FOREGROUND && hwnd != 0 {
			debugPrintln("Foreground event detected!")
			onWindowChange(hwnd)
		}
		return 0
	})

	ret, _, err := winEventHook.Call(
		EVENT_SYSTEM_FOREGROUND,
		EVENT_SYSTEM_FOREGROUND,
		0,
		hookCallback,
		0,
		0,
		WINEVENT_OUTOFCONTEXT,
	)
	debugPrintf("SetWinEventHook result: %d, err: %v\n", ret, err)

	if ret == 0 {
		return fmt.Errorf("SetWinEventHook failed: %w", err)
	}

	hookHandle = ret
	debugPrintln("Hook installed successfully")
	return nil
}

func runMessagePump() {
	debugPrintln("Starting message pump...")
	for {
		var msg Msg
		ret, _, err := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		debugPrintf("GetMessage returned ret=%d message=%d hwnd=%d err=%v\n", ret, msg.message, msg.hwnd, err)
		if int32(ret) <= 0 {
			debugPrintln("Message pump exiting")
			break
		}
		procDispatchMsg.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func initWindowEventHook() {
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if err := installWindowEventHook(); err != nil {
			debugPrintln("Hook failed to install:", err)
			return
		}

		runMessagePump()
	}()
}

func resetDebounceTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(windowChangeDebounce)
	debugPrintf("Debounce timer reset to %s\n", windowChangeDebounce)
}

func startWindowMonitor() {
	var lastWindow uintptr
	var pendingWindow uintptr
	debounceTimer := time.NewTimer(time.Hour)
	if !debounceTimer.Stop() {
		<-debounceTimer.C
	}
	debounceCh := debounceTimer.C

	for {
		select {
		case hwnd := <-foregroundCheck:
			pendingWindow = hwnd
			debugPrintf("Event: foreground window changed to hwnd=%d\n", hwnd)

			if atomic.LoadInt32(&isEnabled) == 0 {
				continue
			}

			resetDebounceTimer(debounceTimer)

		case <-debounceCh:
			if atomic.LoadInt32(&isEnabled) == 0 {
				continue
			}

			hwnd := pendingWindow
			if hwnd != lastWindow && hwnd != 0 {
				lastWindow = hwnd
				title := getWindowTitle(hwnd)
				applyWindowTitle(title)
				updateStatus()
			}
		}
	}
}
