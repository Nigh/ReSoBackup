//go:build !cli

package main

import "github.com/egoist/mygo/ui"

type palette struct {
	theme                            *ui.Theme
	secondary, accent, info, overlay ui.Color
}

func xianiiTheme(dark bool, textScale float32) palette {
	t := ui.DarkTheme()
	p := palette{theme: t, overlay: ui.Oklch(0, 0, 0).Alpha(0.55)}
	t.Dark = dark
	t.Background = ui.Oklch(0.26, 0, 0)
	t.Surface = ui.Oklch(0.20, 0, 0)
	t.Border = ui.Oklch(0.37, 0, 0)
	t.Text = ui.Oklch(0.96, 0.001, 286.375)
	t.Accent = ui.Oklch(0.81, 0.117, 11.638)
	t.AccentText = ui.Oklch(0.26, 0, 0)
	p.secondary = ui.Oklch(0.82, 0.119, 306.383)
	p.accent = ui.Oklch(0.80, 0.08, 185)
	p.info = ui.Oklch(0.75, 0.10, 250)
	t.Success = ui.Oklch(0.75, 0.10, 150)
	t.Warning = ui.Oklch(0.82, 0.14, 85)
	t.Danger = ui.Oklch(0.70, 0.18, 25)
	if !dark {
		t.Background = ui.Oklch(0.97, 0, 0)
		t.Surface = ui.Oklch(0.93, 0, 0)
		t.Border = ui.Oklch(0.87, 0, 0)
		t.Text = ui.Oklch(0.20, 0.001, 286.375)
		t.Accent = ui.Oklch(0.55, 0.14, 11.638)
		t.AccentText = ui.Oklch(0.98, 0, 0)
		p.secondary = ui.Oklch(0.55, 0.14, 306.383)
		p.accent = ui.Oklch(0.52, 0.09, 185)
		p.info = ui.Oklch(0.50, 0.10, 250)
		t.Success = ui.Oklch(0.50, 0.10, 150)
		t.Warning = ui.Oklch(0.55, 0.14, 85)
		t.Danger = ui.Oklch(0.50, 0.19, 25)
	}
	t.TextMuted = t.Text.Alpha(0.70)
	t.SurfaceHover = t.Surface.Mix(t.Text, 0.08)
	t.SurfacePressed = t.Surface.Mix(t.Accent, 0.16)
	t.AccentHover = t.Accent.Mix(t.Text, 0.08)
	t.AccentPressed = t.Accent.Mix(t.Text, 0.16)
	t.Focus = t.Accent
	t.Selection = t.Accent.Alpha(0.30)
	t.Inverse, t.InverseText = t.Surface, t.Text
	t.Radius, t.Spacing = 8, 4
	t.FontSize = 16 * max(1, textScale)
	t.Font = "sans-serif"
	return p
}
