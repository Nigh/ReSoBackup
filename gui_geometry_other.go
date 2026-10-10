//go:build !linux && !cli

package main

import "github.com/egoist/mygo"

func startupWindowFrame(win *mygo.Window) mygo.Rectangle {
	return win.Bounds()
}
