//go:build linux

package main

import (
	"sync/atomic"
	"time"
)

func startWindowMonitor() {
	for {
		if atomic.LoadInt32(&isEnabled) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		time.Sleep(5 * time.Second)

		applyWindowTitle("Linux stub - no window detection")
		updateStatus()
	}
}

func initWindowEventHook() {
}
