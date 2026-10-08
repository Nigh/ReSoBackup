//go:build !cli

package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"reed-solomon-backup/internal/app"
)

type formState struct {
	inputPath, outputDir, password, fileInfo string
	showPassword, valid, needsPassword       bool
	status, statusKind                       string
}

type desktop struct {
	busyLabel, lastAnnouncement               string
	win                                       *mygo.Window
	initialHeight, renderedHeight             int
	lang                                      string
	dark                                      bool
	tab                                       int
	busy, confirmBackup                       bool
	backup, restore                           formState
	shares, threshold, sharesPos, storedRatio float64
	encrypt, encryptFilename                  bool
	pendingBackup                             app.BackupOptions
}

const minimumWindowHeight = 500
const pagePadding = 20

func newDesktop() *desktop {
	return &desktop{lang: "en", dark: true, shares: 8, threshold: 5,
		sharesPos: posFromValue(8), storedRatio: 8.0 / 5, encryptFilename: true}
}

func createApp() *mygo.Application {
	s := newDesktop()
	mygo.App.SetName("ReSo Backup")
	mygo.App.WhenReady(func() {
		s.win = mygo.NewWindow(mygo.WindowOptions{
			Title: "ReSo Backup", Width: 900, Height: minimumWindowHeight, MinWidth: 700, MinHeight: minimumWindowHeight,
			UseContentSize: true, Hidden: true,
			Content: ui.View(s.view),
		})
		s.win.Show()
		s.win.OnClose(func(e *mygo.CloseEvent) {
			if s.busy {
				e.PreventDefault()
				s.currentForm().status, s.currentForm().statusKind = s.t("busyClose"), "info"
				s.win.Invalidate()
			}
		})
	})
	mygo.App.OnBeforeQuit(func(e *mygo.QuitEvent) {
		if s.busy {
			e.PreventDefault()
			s.currentForm().status, s.currentForm().statusKind = s.t("busyClose"), "info"
			s.win.Invalidate()
		}
	})
	return mygo.App
}

// Wayland can discard a resize during its initial configure handshake.
func (s *desktop) fitStartupWindow(attempts int) {
	time.AfterFunc(200*time.Millisecond, func() {
		s.win.Update(func() {
			if s.win.IsDestroyed() {
				return
			}
			bounds := s.win.Bounds()
			width, height := s.win.ContentSize()
			area := mygo.Screen.DisplayNearestPoint(mygo.Point{X: bounds.X + bounds.Width/2, Y: bounds.Y + bounds.Height/2}).WorkArea
			available := max(1, area.Height-max(0, bounds.Height-height))
			target := min(max(minimumWindowHeight, s.initialHeight), available)
			if s.renderedHeight == target {
				return
			}
			s.win.SetMinimumSize(700, min(minimumWindowHeight, available))
			s.win.SetContentSize(width, target)
			if attempts > 1 {
				s.fitStartupWindow(attempts - 1)
			}
		})
	})
}

func (s *desktop) currentForm() *formState {
	if s.tab == 0 {
		return &s.backup
	}
	return &s.restore
}

func (s *desktop) view(c *ui.Context) {
	p := xianiiTheme(s.dark, c.Preferences().TextScale)
	if c.Preferences().HighContrast {
		p.theme.Border, p.theme.TextMuted = p.theme.Text, p.theme.Text
	}
	c.SetTheme(p.theme)
	var pageBounds, contentBounds ui.Rect
	ui.Column(c).Fill().Background(p.theme.Background).DrawOver(func(_ *ui.Painter, r ui.Rect) {
		s.renderedHeight = int(math.Round(float64(r.H)))
		if s.initialHeight != 0 {
			return
		}
		s.initialHeight = int(math.Ceil(float64(r.H - pageBounds.H + contentBounds.H + 2*pagePadding)))
		if s.win != nil {
			s.fitStartupWindow(3)
		}
	}).Children(func() {
		ui.Row(c).Wrap().Padding(18, 24).Gap(12).AlignItems(ui.Center).Background(p.theme.Surface).Children(func() {
			ui.Text(c, s.t("appTitle")).FontSize(c.Theme().Rem(1.5)).Bold().Grow(1)
			language := "中文"
			if s.lang == "zh" {
				language = "EN"
			}
			if ui.Button(c.Key("language"), language).Clicked() {
				if s.lang == "en" {
					s.lang = "zh"
				} else {
					s.lang = "en"
				}
			}
			themeLabel := s.t("lightTheme")
			if !s.dark {
				themeLabel = s.t("darkTheme")
			}
			if ui.Button(c.Key("theme"), themeLabel).Clicked() {
				s.dark = !s.dark
			}
			ui.Button(c.Key("github"), "GitHub").OnClick(func() {
				if err := mygo.Shell.OpenExternal("https://github.com/Nigh/ReSoBackup"); err != nil {
					s.currentForm().status, s.currentForm().statusKind = err.Error(), "error"
				}
			})
		})
		ui.Row(c).Justify(ui.Center).Padding(16, 24, 0).Children(func() {
			tabs := ui.TabsBase(c.Key("tabs"), &s.tab, 2)
			tabs.List.Width(376).Padding(4).Gap(4).Radius(12).
				Background(p.theme.Surface).Border(1, p.theme.Border).Disabled(s.busy).Children(func() {
				for i, label := range []string{s.t("backupTab"), s.t("restoreTab")} {
					tab := tabs.Tab(i).Label(label).Grow(1).MinWidth(0).Height(52).Radius(8).
						Justify(ui.Center).AlignItems(ui.Center).TextColor(p.theme.TextMuted)
					if i == s.tab {
						color := p.theme.Accent
						if i == 1 {
							color = p.secondary
						}
						tab.Background(color).TextColor(p.theme.AccentText).FontWeight(600)
					} else if tab.Hovered() {
						tab.Background(p.theme.SurfaceHover).TextColor(p.theme.Text)
					}
					tab.Children(func() { ui.Text(c, label).FontSize(c.Theme().Rem(1.125)).SingleLine() })
				}
			})
		})
		ui.Scroll(c.Key(fmt.Sprintf("page-%d", s.tab))).Grow(1).Padding(pagePadding, 24).Gap(18).
			Draw(func(_ *ui.Painter, r ui.Rect) { pageBounds = r }).Children(func() {
			ui.Column(c).WidthPercent(100).MaxWidth(680).AlignSelf(ui.Center).Gap(18).
				Draw(func(_ *ui.Painter, r ui.Rect) { contentBounds = r }).Children(func() {
				if s.tab == 0 {
					s.backupView(c, p)
				} else {
					s.restoreView(c, p)
				}
			})
		})
		ui.Text(c, s.t("footer")).FontSize(c.Theme().Rem(0.75)).TextColor(p.theme.TextMuted).Padding(14, 24).Background(p.theme.Surface)
	})
	s.warningDialog(c, p)
}

func (s *desktop) card(c *ui.Context, p palette, title string, color ui.Color, body func()) {
	ui.Column(c).Padding(20).Gap(12).Radius(8).Background(p.theme.Surface).Children(func() {
		ui.Text(c, title).FontSize(c.Theme().Rem(1.25)).Bold().TextColor(color).Role(ui.RoleHeading).Level(2)
		body()
	})
}

func (s *desktop) fileField(c *ui.Context, f *formState, restore bool) {
	placeholder := s.t("selectFile")
	if restore {
		placeholder = s.t("selectRestore")
	}
	ui.Row(c).Gap(8).Children(func() {
		ui.TextInput(c.Key("input-path"), &f.inputPath).ReadOnly(true).Placeholder(placeholder).Label(placeholder).Grow(1)
		ui.Button(c.Key("input-browse"), s.t("browse")).OnClick(func() { s.pickFile(restore) })
	})
	if f.fileInfo != "" {
		ui.Text(c, f.fileInfo).FontSize(c.Theme().Rem(0.875)).TextColor(c.Theme().TextMuted)
	}
}

func (s *desktop) passwordField(c *ui.Context, f *formState, placeholder string) {
	ui.Field(c, s.t("password"), func() {
		ui.Row(c).Gap(8).Children(func() {
			input := ui.TextInput(c.Key("password"), &f.password).Placeholder(placeholder).Grow(1)
			if !f.showPassword {
				input.Password()
			}
			label := s.t("showPassword")
			if f.showPassword {
				label = s.t("hidePassword")
			}
			if ui.Button(c.Key("password-visibility"), label).Clicked() {
				f.showPassword = !f.showPassword
			}
		})
	})
}

func (s *desktop) outputField(c *ui.Context, f *formState, restore bool) {
	placeholder, hint := s.t("outputDirPlaceholder"), s.t("outputDirHint")
	if restore {
		placeholder, hint = s.t("restoreOutputPlaceholder"), s.t("restoreOutputHint")
	}
	ui.Column(c).Gap(8).Children(func() {
		ui.Row(c).Gap(8).Children(func() {
			ui.TextInput(c.Key("output-path"), &f.outputDir).Placeholder(placeholder).Label(s.t("outputDir")).Grow(1)
			ui.Button(c.Key("output-browse"), s.t("browse")).OnClick(func() { s.pickDirectory(restore) })
		})
	})
	ui.Text(c, hint).FontSize(c.Theme().Rem(0.875)).TextColor(c.Theme().TextMuted)
}

func (s *desktop) backupView(c *ui.Context, p palette) {
	ui.Column(c).Gap(12).Children(func() {
		s.card(c, p, s.t("inputFile"), p.theme.Accent, func() { s.fileField(c, &s.backup, false) })
		s.card(c, p, s.t("rsParams"), p.secondary, func() {
			ui.Field(c, s.t("totalShares"), func() {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					if ui.Slider(c.Key("shares-slider"), &s.sharesPos, 0, 1000).Grow(1).Label(s.t("totalShares")).Changed() {
						s.setShares(valueFromPos(s.sharesPos))
					}
					if ui.NumberInput(c.Key("shares-number"), &s.shares, 3, 128, 1).Width(150).Label(s.t("totalShares")).Changed() {
						s.setShares(s.shares)
					}
				})
			})
			ui.Field(c, s.t("threshold"), func() {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					if ui.StepSlider(c.Key("threshold-slider"), &s.threshold, 1, s.shares, 1).Grow(1).Label(s.t("threshold")).Changed() {
						s.setThreshold(s.threshold)
					}
					if ui.NumberInput(c.Key("threshold-number"), &s.threshold, 1, s.shares, 1).Width(150).Label(s.t("threshold")).Changed() {
						s.setThreshold(s.threshold)
					}
				})
			})
			ui.Text(c, fmt.Sprintf("%s: %.1f%%   ·   %s: %.2fx", s.t("redundancy"), (s.shares-s.threshold)/s.shares*100, s.t("storageOverhead"), s.shares/s.threshold))
			ui.Text(c, strings.ReplaceAll(s.t("canLose"), "{n}", fmt.Sprint(int(s.shares-s.threshold)))).TextColor(p.theme.TextMuted)
			for _, warning := range app.BackupWarnings(s.backupOptions()) {
				ui.Text(c, warning).TextColor(p.theme.Warning).FontSize(c.Theme().Rem(0.875))
			}
		})
		s.card(c, p, s.t("encryption"), p.accent, func() {
			ui.Checkbox(c.Key("encrypt"), &s.encrypt, s.t("enableEncryption"))
			if s.encrypt {
				ui.Checkbox(c.Key("encrypt-filename"), &s.encryptFilename, s.t("encryptFilename"))
				ui.Text(c, s.t("encryptFilenameDesc")).TextColor(p.theme.TextMuted).FontSize(c.Theme().Rem(0.875))
				s.passwordField(c, &s.backup, s.t("enterPassword"))
			}
		})
		s.card(c, p, s.t("outputDir"), p.info, func() { s.outputField(c, &s.backup, false) })
	}).Disabled(s.busy)
	s.statusView(c, &s.backup, p)
	ui.PrimaryButton(c.Key("backup-submit"), s.t("createBackup")).WidthPercent(100).
		Disabled(s.busy || s.backup.inputPath == "" || (s.encrypt && s.backup.password == "")).OnClick(s.requestBackup)
}

func (s *desktop) restoreView(c *ui.Context, p palette) {
	ui.Column(c).Gap(12).Children(func() {
		s.card(c, p, s.t("shareFile"), p.theme.Accent, func() {
			s.fileField(c, &s.restore, true)
			ui.Text(c, s.t("restoreHint")).TextColor(p.theme.TextMuted).FontSize(c.Theme().Rem(0.875))
		})
		if s.restore.needsPassword {
			s.card(c, p, s.t("decryptPwd"), p.accent, func() { s.passwordField(c, &s.restore, s.t("enterRestorePwd")) })
		}
		s.card(c, p, s.t("outputDir"), p.info, func() { s.outputField(c, &s.restore, true) })
	}).Disabled(s.busy)
	s.statusView(c, &s.restore, p)
	ui.PrimaryButton(c.Key("restore-submit"), s.t("restoreFile")).WidthPercent(100).
		Disabled(s.busy || !s.restore.valid || (s.restore.needsPassword && s.restore.password == "")).OnClick(s.startRestore)
}

func (s *desktop) statusView(c *ui.Context, f *formState, p palette) {
	if s.busy {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Spinner(c)
			ui.Text(c, s.busyLabel)
		})
	}
	if f.status != s.lastAnnouncement {
		s.lastAnnouncement = f.status
		if f.status != "" {
			c.Announce(f.status)
		}
	}
	if f.status == "" {
		return
	}
	color := p.info
	if f.statusKind == "error" {
		color = p.theme.Danger
	}
	if f.statusKind == "success" {
		color = p.theme.Success
	}
	ui.Text(c, f.status).Selectable().Padding(14).Radius(8).Background(color.Alpha(0.12)).Border(1, color)
}

func (s *desktop) warningDialog(c *ui.Context, p palette) {
	ui.DialogBase(c, &s.confirmBackup, func(backdrop, panel ui.Element) {
		backdrop.Background(p.overlay)
		panel.Width(520).MaxWidthPercent(90).Padding(24).Gap(16).Radius(8).Background(p.theme.Surface).Label(s.t("warnTitle"))
		ui.Text(c, s.t("warnTitle")).Bold().FontSize(c.Theme().Rem(1.25))
		for _, warning := range app.BackupWarnings(s.pendingBackup) {
			ui.Text(c, warning)
		}
		ui.Text(c, s.t("confirmContinue"))
		ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
			if ui.Button(c.Key("warning-cancel"), s.t("cancel")).AutoFocus().Clicked() {
				s.confirmBackup = false
				s.pendingBackup = app.BackupOptions{}
			}
			ui.PrimaryButton(c.Key("warning-continue"), s.t("continue")).OnClick(func() {
				opts := s.pendingBackup
				s.confirmBackup = false
				s.pendingBackup = app.BackupOptions{}
				s.startBackup(opts)
			})
		})
	})
	if !s.confirmBackup {
		s.pendingBackup = app.BackupOptions{}
	}
}

func (s *desktop) run(label string, work func() (string, error), done func(string, error)) {
	if s.busy {
		return
	}
	s.busy, s.busyLabel = true, label
	go func() {
		result, err := work()
		s.win.Update(func() { s.busy = false; done(result, err) })
	}()
}

func (s *desktop) pickFile(restore bool) {
	title := s.t("inputFile")
	if restore {
		title = s.t("shareFile")
	}
	var metaEncrypted bool
	var info string
	sizeLabel, invalidLabel := s.t("fileInfo"), s.t("invalidFile")
	s.run(title, func() (string, error) {
		opts := mygo.OpenDialogOptions{Parent: s.win, Title: title}
		if restore {
			opts.Filters = []mygo.FileFilter{{Name: "RS Meta", Extensions: []string{"rsmeta"}}, {Name: "All Files", Extensions: []string{"*"}}}
		}
		paths, err := mygo.Dialog.Open(opts)
		if err != nil || len(paths) == 0 {
			return "", err
		}
		path := paths[0]
		stat, err := os.Stat(path)
		if err != nil {
			return path, err
		}
		if stat.IsDir() {
			return path, errors.New("path is a directory")
		}
		info = fmt.Sprintf(sizeLabel, formatSize(stat.Size()))
		if restore {
			meta, err := app.ReadMetadata(path)
			if err != nil {
				return path, fmt.Errorf("%s: %w", invalidLabel, err)
			}
			metaEncrypted = meta.Encrypted || meta.Version <= 1
		}
		return path, nil
	}, func(path string, err error) {
		f := &s.backup
		if restore {
			f = &s.restore
		}
		if path != "" {
			*f = formState{inputPath: path, fileInfo: info, valid: err == nil, needsPassword: metaEncrypted, outputDir: f.outputDir}
			if !restore {
				f.outputDir = ""
			}
		}
		if err != nil {
			f.status, f.statusKind = err.Error(), "error"
		} else if restore && path != "" && !metaEncrypted {
			f.status, f.statusKind = s.t("noPasswordNeeded"), "info"
		}
	})
}

func (s *desktop) pickDirectory(restore bool) {
	title := s.t("outputDir")
	s.run(title, func() (string, error) {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: s.win, Title: title, Directory: true, CreateDirectories: true})
		if err != nil || len(paths) == 0 {
			return "", err
		}
		return paths[0], nil
	}, func(path string, err error) {
		f := &s.backup
		if restore {
			f = &s.restore
		}
		if err != nil {
			f.status, f.statusKind = err.Error(), "error"
		} else if path != "" {
			f.outputDir = path
		}
	})
}

func (s *desktop) backupOptions() app.BackupOptions {
	return app.BackupOptions{InputPath: s.backup.inputPath, OutputDir: s.backup.outputDir,
		Password: s.backup.password, Shares: int(s.shares), Threshold: int(s.threshold),
		Encrypt: s.encrypt, EncryptFilename: s.encrypt && s.encryptFilename}
}

func (s *desktop) requestBackup() {
	if s.busy {
		return
	}
	opts := s.backupOptions()
	if err := app.ValidateBackupOptions(opts); err != nil {
		s.backup.status, s.backup.statusKind = err.Error(), "error"
		return
	}
	if opts.Encrypt && opts.Password == "" {
		s.backup.status, s.backup.statusKind = s.t("errEnterPwd"), "error"
		return
	}
	if len(app.BackupWarnings(opts)) > 0 {
		s.pendingBackup, s.confirmBackup = opts, true
		return
	}
	s.startBackup(opts)
}

func (s *desktop) startBackup(opts app.BackupOptions) {
	s.backup.status = ""
	s.run(s.t("creatingBackup"), func() (string, error) { return app.RunBackup(opts) }, func(prefix string, err error) {
		if err != nil {
			s.backup.status, s.backup.statusKind = err.Error(), "error"
			return
		}
		s.backup.status, s.backup.statusKind = s.t("backupSuccess"), "success"
		if opts.EncryptFilename {
			s.backup.status = fmt.Sprintf(s.t("backupSuccessEncFN"), prefix)
		}
	})
}

func (s *desktop) startRestore() {
	if s.busy || !s.restore.valid {
		return
	}
	if s.restore.needsPassword && s.restore.password == "" {
		s.restore.status, s.restore.statusKind = s.t("errEnterRestorePwd"), "error"
		return
	}
	opts := app.RestoreOptions{InputPath: s.restore.inputPath, OutputDir: s.restore.outputDir, Password: s.restore.password}
	s.restore.status = ""
	s.run(s.t("restoring"), func() (string, error) { return "", app.RunRestore(opts) }, func(_ string, err error) {
		if err != nil {
			s.restore.status, s.restore.statusKind = err.Error(), "error"
		} else {
			s.restore.status, s.restore.statusKind = s.t("restoreSuccess"), "success"
		}
	})
}

func valueFromPos(pos float64) float64 {
	pos = max(0, min(1000, pos))
	if pos <= 500 {
		return math.Round(3 + 13*pos/500)
	}
	return math.Round(16 + 112*(pos-500)/500)
}

func posFromValue(value float64) float64 {
	value = max(3, min(128, value))
	if value <= 16 {
		return 500 * (value - 3) / 13
	}
	return 500 + 500*(value-16)/112
}

func (s *desktop) setShares(value float64) {
	s.shares = math.Round(max(3, min(128, value)))
	s.sharesPos = posFromValue(s.shares)
	s.threshold = max(1, min(s.shares, math.Round(s.shares/s.storedRatio)))
}

func (s *desktop) setThreshold(value float64) {
	s.threshold = math.Round(max(1, min(s.shares, value)))
	s.storedRatio = s.shares / s.threshold
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.2f GB", float64(bytes)/(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.2f KB", float64(bytes)/(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
