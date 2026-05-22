package main

import (
	_ "embed"
	"fmt"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/getlantern/systray"
)

//go:embed icon.ico
var iconData []byte

var isEnabled int32 = 1
var fwVerString = ""
var mStatus *systray.MenuItem
var mVer *systray.MenuItem

var systrayMu sync.Mutex

func pollingRateToString(rate uint8) string {
	switch rate {
	case 8:
		return "8k"
	case 4:
		return "4k"
	case 2:
		return "2k"
	case 1:
		return "1k"
	default:
		return fmt.Sprintf("%dk", rate)
	}
}

func updateStatus() {
	systrayMu.Lock()
	defer systrayMu.Unlock()
	rateStr := pollingRateToString(GetCurrentPollingRate())
	if atomic.LoadInt32(&isEnabled) == 1 {
		mStatus.SetTitle(fmt.Sprintf("Status: Enabled (%s)", rateStr))
		systray.SetTooltip(fmt.Sprintf("Monitoring active windows - %s", rateStr))
	} else {
		mStatus.SetTitle(fmt.Sprintf("Status: Disabled (%s)", rateStr))
		systray.SetTooltip(fmt.Sprintf("Paused - %s", rateStr))
	}
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("Window Watcher")
	systray.SetTooltip("Monitoring active windows")

	option := "fwVer: " + fwVerString
	mVer = systray.AddMenuItem(option, "Version")
	mStatus = systray.AddMenuItem("Status", "Monitoring active windows")
	mEnable := systray.AddMenuItem("Disable", "Pause window monitoring")
	mQuit := systray.AddMenuItem("Quit", "Exit application")

	mVer.Disable()
	mStatus.Disable()
	updateStatus()

	initWindowEventHook()

	go func() {
		for {
			select {
			case <-mEnable.ClickedCh:
				systrayMu.Lock()
				if atomic.SwapInt32(&isEnabled, 1-atomic.LoadInt32(&isEnabled)) == 1 {
					mEnable.SetTitle("Disable")
				} else {
					mEnable.SetTitle("Enable")
				}
				systrayMu.Unlock()
				updateStatus()

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
