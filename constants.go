package main

import "sync"

const (
	POLLING_RATE_8K uint8 = 8
	POLLING_RATE_4K uint8 = 4
	POLLING_RATE_2K uint8 = 2
	POLLING_RATE_1K uint8 = 1
)

var currentPollingRate uint8 = POLLING_RATE_8K
var rateMu sync.Mutex

func GetCurrentPollingRate() uint8 {
	rateMu.Lock()
	defer rateMu.Unlock()
	return currentPollingRate
}

func SetCurrentPollingRate(rate uint8) {
	rateMu.Lock()
	defer rateMu.Unlock()
	currentPollingRate = rate
}
