//go:build linux

package main

import "time"

func startWindowMonitor() {
	for {
		if !isEnabled {
			time.Sleep(1 * time.Second)
			continue
		}

		time.Sleep(5 * time.Second)

		debugPrintln("Window changed:", "Linux stub - no window detection")

		debugPrintln("WE NOT IN LEAGUE")
		set8k()
		setConfig()
	}
}

func initWindowEventHook() {
}
