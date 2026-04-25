package main

import (
	_ "embed"
	"syscall"
	"time"

	"github.com/getlantern/systray"
)

// Embed the icon into the binary
//
//go:embed icon.ico
var iconData []byte

var isEnabled = true

var fwVerString = ""

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("Window Watcher")
	systray.SetTooltip("Monitoring active windows")

	option := "fwVer: " + fwVerString
	mVer := systray.AddMenuItem(option, "Version")
	mStatus := systray.AddMenuItem("Status", "Monitoring active windows")
	mEnable := systray.AddMenuItem("Disable", "Pause window monitoring")
	mQuit := systray.AddMenuItem("Quit", "Exit application")

	mVer.Disable()
	mStatus.Disable()
	mStatus.SetTitle("Status: Enabled")

	initWindowEventHook()
	go startWindowMonitor()

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
				debugPrintln("Version clicked")
			}
		}
	}()
}

func onExit() {
	debugPrintln("Exiting...")
	time.Sleep(1 * time.Second)
	debugPrintln("Goodbye!")
}

func main() {
	value, err := getDeviceInfo()
	if err != nil {
		debugPrintln("Error: ", err)
		syscall.Exit(1)
	}

	fwVerString = value

	getConfig()

	systray.Run(onReady, onExit)
}
