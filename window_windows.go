//go:build windows

package main

import (
	"fmt"
	"runtime"
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
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type point struct {
	x int32
	y int32
}

func getForegroundWindow() uintptr {
	hwnd, _, _ := procGetForeground.Call()
	return hwnd
}

func getWindowTitle(hwnd uintptr) string {
	buf := make([]uint16, 256)
	procGetWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

var foregroundCheck = make(chan struct{}, 1)
var hookCallback uintptr
var hookHandle uintptr

func onWindowChange() {
	debugPrintln("onWindowChange triggered")
	select {
	case foregroundCheck <- struct{}{}:
	default:
	}
}

func installWindowEventHook() error {
	debugPrintln("Initializing window event hook...")

	hookCallback = windows.NewCallback(func(eventHook, event, hwnd, idObject, childWnd, threadId, timestamp uintptr) uintptr {
		debugPrintf("!!! CALLBACK FIRED: event=%d, hwnd=%d !!!\n", event, hwnd)
		if event == EVENT_SYSTEM_FOREGROUND && hwnd != 0 {
			debugPrintln("Foreground event detected!")
			onWindowChange()
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
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
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

func startWindowMonitor() {
	var lastWindow uintptr
	var debounceTimer *time.Timer
	var debounceCh <-chan time.Time

	for {
		select {
		case <-foregroundCheck:
			debugPrintln("Event: foreground window changed")

			if !isEnabled {
				continue
			}

			if debounceTimer == nil {
				debounceTimer = time.NewTimer(windowChangeDebounce)
			} else {
				if !debounceTimer.Stop() {
					select {
					case <-debounceTimer.C:
					default:
					}
				}
				debounceTimer.Reset(windowChangeDebounce)
			}

			debounceCh = debounceTimer.C

		case <-debounceCh:
			debounceCh = nil

			if !isEnabled {
				continue
			}

			hwnd := getForegroundWindow()
			if hwnd != lastWindow && hwnd != 0 {
				lastWindow = hwnd
				title := getWindowTitle(hwnd)
				debugPrintln("Window changed:", title)

				if matched, rate := matchWindow(title); matched {
					debugPrintf("Matched target window, setting %dk\n", rate)
					setPollingRate(rate)
				} else {
					debugPrintln("Defaulting to 8k")
					set8k()
				}
				setConfig()
			}
		}
	}
}
