package main

import "strings"

type TargetWindow struct {
	Name string
	Rate uint8
}

var targetWindows = []TargetWindow{
	{Name: "league", Rate: POLLING_RATE_1K},
	{Name: "getting over it", Rate: POLLING_RATE_1K},
}

func matchWindow(title string) (bool, uint8) {
	lower := strings.ToLower(title)
	for _, tw := range targetWindows {
		if strings.Contains(lower, tw.Name) {
			return true, tw.Rate
		}
	}
	return false, 0
}
