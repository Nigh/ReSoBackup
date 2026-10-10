//go:build linux && !cli

package main

import (
	"log"

	"github.com/ebitengine/purego"
	"github.com/egoist/mygo"
)

// GTK's size excludes decorations; measure the frame before fitting it.
func startupWindowFrame(win *mygo.Window) mygo.Rectangle {
	bounds := win.Bounds()
	lib, err := purego.Dlopen("libgtk-3.so.0", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		log.Printf("cannot measure window decorations: %v", err)
		return bounds
	}
	defer purego.Dlclose(lib)
	var getWindow func(uintptr) uintptr
	var getFrame func(uintptr, *struct{ X, Y, Width, Height int32 })
	purego.RegisterLibFunc(&getWindow, lib, "gtk_widget_get_window")
	purego.RegisterLibFunc(&getFrame, lib, "gdk_window_get_frame_extents")
	window := getWindow(win.NativeHandle())
	if window == 0 {
		return bounds
	}
	var frame struct{ X, Y, Width, Height int32 }
	getFrame(window, &frame)
	if frame.Width <= 0 || frame.Height <= 0 {
		return bounds
	}
	return mygo.Rectangle{X: int(frame.X), Y: int(frame.Y), Width: int(frame.Width), Height: int(frame.Height)}
}
