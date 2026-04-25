//go:build !debug

package main

func debugPrintln(args ...any) {}

func debugPrintf(format string, args ...any) {}
