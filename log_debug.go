//go:build debug

package main

import "fmt"

func debugPrintln(args ...any) {
	fmt.Println(args...)
}

func debugPrintf(format string, args ...any) {
	fmt.Printf(format, args...)
}
