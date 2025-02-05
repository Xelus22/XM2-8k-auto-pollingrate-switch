package main

import (
	_ "embed"
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

// Embed the icon into the binary
//
//go:embed icon.ico
var iconData []byte

// Track app state
var isEnabled = true

var isLeague = false

var fwVerString = ""

var (
	user32            = windows.NewLazySystemDLL("user32.dll")
	procGetForeground = user32.NewProc("GetForegroundWindow")
	procGetWindowText = user32.NewProc("GetWindowTextW")
)

// getForegroundWindow retrieves the handle of the currently focused window.
func getForegroundWindow() uintptr {
	hwnd, _, _ := procGetForeground.Call()
	return hwnd
}

// getWindowTitle retrieves the title of a window by its handle.
func getWindowTitle(hwnd uintptr) string {
	buf := make([]uint16, 256)
	procGetWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

// Monitors window changes and triggers asyncFunction when needed.
func startWindowMonitor() {
	var lastWindow uintptr

	for {
		if !isEnabled {
			time.Sleep(1 * time.Second) // Pause checking when disabled
			continue
		}

		hwnd := getForegroundWindow()
		if hwnd != lastWindow && hwnd != 0 {
			lastWindow = hwnd
			title := getWindowTitle(hwnd)
			fmt.Println("Window changed:", title)

			// Check for target window title
			// EDIT ME IF YOU WANT TO CHANGE WHAT WINDOW TO DETECT
			if strings.Contains(strings.ToLower(title), "league") {
				fmt.Println("WE IN LEAGUE")
				set1k()
			} else {
				fmt.Println("WE NOT IN LEAGUE")
				set8k()
			}
			setConfig()
		}
		time.Sleep(5000 * time.Millisecond) // Polling interval
	}
}

// onReady initializes the system tray
func onReady() {
	// Load the tray icon
	systray.SetIcon(iconData)
	systray.SetTitle("Window Watcher")
	systray.SetTooltip("Monitoring active windows")

	// Create menu items
	option := "fwVer: " + fwVerString
	mVer := systray.AddMenuItem(option, "Version")
	mStatus := systray.AddMenuItem("Status", "Monitoring active windows")
	mEnable := systray.AddMenuItem("Disable", "Pause window monitoring")
	mQuit := systray.AddMenuItem("Quit", "Exit application")

	mVer.Disable()
	mStatus.Disable()
	mStatus.SetTitle("Status: Enabled")

	// Run window monitor in background
	go startWindowMonitor()

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mEnable.ClickedCh:
				isEnabled = !isEnabled
				if isEnabled {
					mEnable.SetTitle("Disable")
					systray.SetTooltip("Monitoring active windows")
					mStatus.SetTitle("Status: Enabled")
				} else {
					mEnable.SetTitle("Enable")
					systray.SetTooltip("Paused")
					mStatus.SetTitle("Status: Disabled")
				}

			case <-mQuit.ClickedCh:
				systray.Quit()

			case <-mVer.ClickedCh:
				fmt.Println("Version clicked")
			}
		}
	}()
}

func onExit() {
	// Cleanup before exit
	fmt.Println("Exiting...")
	time.Sleep(1 * time.Second)
	fmt.Println("Goodbye!")
}

func main() {
	// get device info
	value, err := getDeviceInfo()
	if err != nil {
		fmt.Println("Error: ", err)
		syscall.Exit(1)
	}

	fwVerString = value

	getConfig()

	// Initialize the system tray
	systray.Run(onReady, onExit)
}
