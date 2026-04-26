//go:build darwin

package main

import "time"

func startWindowMonitor() {
	for {
		if !isEnabled {
			time.Sleep(1 * time.Second)
			continue
		}

		time.Sleep(5 * time.Second)

		applyWindowTitle("macOS stub - no window detection")
	}
}

func initWindowEventHook() {
}
