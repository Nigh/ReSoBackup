//go:build !cli

package main

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestNativeNavigationAndAppearance(t *testing.T) {
	s := newDesktop()
	tt := ui.NewTester(s.view, 900, 800)
	if !tt.HasText("Create Backup") {
		t.Fatal("backup form missing")
	}
	if box, ok := tt.Find("Select a file to backup..."); !ok || box.H < 28 {
		t.Fatalf("input text would be clipped: %v", box)
	}
	if err := tt.Click("Restore"); err != nil {
		t.Fatal(err)
	}
	if s.tab != 1 || !tt.HasText("Restore File") {
		t.Fatal("restore form missing")
	}
	if err := tt.Click("中文"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("恢复文件") {
		t.Fatal("Chinese translation missing")
	}
	if err := tt.Click("浅色主题"); err != nil {
		t.Fatal(err)
	}
	if s.dark {
		t.Fatal("theme did not change")
	}
	tt.SetSize(700, 500)
	if !tt.HasText("恢复文件") {
		t.Fatal("restore form lost after resize")
	}
}

func TestShareMappingAndRatio(t *testing.T) {
	for shares := 3.0; shares <= 128; shares++ {
		if got := valueFromPos(posFromValue(shares)); got != shares {
			t.Fatalf("%v mapped to %v", shares, got)
		}
	}
	s := newDesktop()
	s.setShares(16)
	if s.threshold != 10 {
		t.Fatalf("threshold %v, want 10", s.threshold)
	}
	s.setThreshold(8)
	s.setShares(32)
	if s.threshold != 16 {
		t.Fatal("chosen storage ratio not preserved")
	}
	s.setThreshold(1000)
	if s.threshold != s.shares {
		t.Fatal("threshold exceeds shares")
	}
	s.setShares(-1)
	if s.shares != 3 || s.threshold < 1 {
		t.Fatal("invalid minimum")
	}
	if math.IsInf(s.storedRatio, 0) {
		t.Fatal("invalid ratio")
	}
}

func TestNativePasswordAndBusyState(t *testing.T) {
	s := newDesktop()
	s.encrypt = true
	s.backup.password = "secret-value"
	tt := ui.NewTester(s.view, 900, 1100)
	if tt.HasText("secret-value") {
		t.Fatal("password visible")
	}
	if err := tt.Click("Show password"); err != nil {
		t.Fatal(err)
	}
	if !s.backup.showPassword {
		t.Fatal("password visibility did not change")
	}
	s.busy = true
	tt.Frame()
	if err := tt.Click("Restore"); err != nil {
		t.Fatal(err)
	}
	if s.tab != 0 {
		t.Fatal("tab changed during operation")
	}
	if err := tt.Click("Enable Encryption"); err != nil {
		t.Fatal(err)
	}
	if !s.encrypt {
		t.Fatal("form changed during operation")
	}
}

func TestXianiiTheme(t *testing.T) {
	for _, dark := range []bool{true, false} {
		p := xianiiTheme(dark, 1.5)
		if p.theme.Dark != dark || p.theme.FontSize != 24 || p.theme.Radius != 8 {
			t.Fatal("theme metrics differ")
		}
		if p.theme.Accent == p.theme.AccentText {
			t.Fatal("primary button has no contrast")
		}
	}
}

func TestRiskDialogCancel(t *testing.T) {
	s := newDesktop()
	s.lang = "zh"
	s.shares, s.threshold = 3, 3
	dir := t.TempDir()
	s.backup.inputPath = filepath.Join(dir, "input.txt")
	s.backup.outputDir = filepath.Join(dir, "backup")
	if err := os.WriteFile(s.backup.inputPath, []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	s.requestBackup()
	if !s.confirmBackup {
		t.Fatal("risk did not require confirmation")
	}
	tt := ui.NewTester(s.view, 900, 800)
	if !tt.HasText("是否继续？") {
		t.Fatal("confirmation missing")
	}
	tt.Key(0, ui.KeyEscape)
	if s.confirmBackup || s.pendingBackup.InputPath != "" {
		t.Fatal("escape did not clear confirmation")
	}
	if _, err := os.Stat(s.backup.outputDir); !os.IsNotExist(err) {
		t.Fatal("cancel wrote backup files")
	}
}

func TestMeasuredHeightFitsBackupAndModeSwitch(t *testing.T) {
	for _, lang := range []string{"en", "zh"} {
		for _, encrypted := range []bool{false, true} {
			s := newDesktop()
			s.lang, s.encrypt = lang, encrypted
			tt := ui.NewTester(s.view, 900, minimumWindowHeight)
			height := s.initialHeight
			if height <= minimumWindowHeight {
				t.Fatalf("content was not measured: %d", height)
			}
			tt.SetSize(900, height)
			submit, ok := tt.Find(s.t("createBackup"))
			footer, footerOK := tt.Find(s.t("footer"))
			if !ok || !footerOK || submit.Y+submit.H > footer.Y {
				t.Fatalf("%s backup action requires scrolling: %v, footer %v", lang, submit, footer)
			}
			if footer.Y-(submit.Y+submit.H) > 32 {
				t.Fatalf("excessive space below form: %v, footer %v", submit, footer)
			}
			backup, backupOK := tt.Find(s.t("backupTab"))
			restore, restoreOK := tt.Find(s.t("restoreTab"))
			if !backupOK || !restoreOK || math.Abs(float64(backup.W-restore.W)) > 1 || math.Abs(float64(backup.X+restore.X+restore.W-900)) > 1 {
				t.Fatalf("mode switches are not symmetric: %v, %v", backup, restore)
			}
			if backup.H < 44 {
				t.Fatalf("mode target too small: %v", backup)
			}
			if err := tt.Click(s.t("backupTab")); err != nil {
				t.Fatal(err)
			}
			tt.Key(0, ui.KeyRight)
			if s.tab != 1 {
				t.Fatal("right arrow did not switch mode")
			}
			tt.Key(0, ui.KeyLeft)
			if s.tab != 0 {
				t.Fatal("left arrow did not switch mode")
			}
			tt.SetSize(900, 600)
			if s.initialHeight != height {
				t.Fatal("initial size changed after manual resize")
			}
		}
	}
}

func TestMeasuredHeightDoesNotDependOnInitialWindowHeight(t *testing.T) {
	var measured int
	for _, height := range []int{500, 1200, 1600} {
		s := newDesktop()
		ui.NewTester(s.view, 900, height)
		if measured == 0 {
			measured = s.initialHeight
		}
		if s.initialHeight != measured {
			t.Fatalf("initial window %d produced content height %d; want %d", height, s.initialHeight, measured)
		}
	}
}
