package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/wailsapp/wails/v3/pkg/application"

	"reed-solomon-backup/internal/app"
)

type BackupService struct {
	wailsApp *application.App
}

func NewBackupService() *BackupService {
	return &BackupService{}
}

func (s *BackupService) SelectInputFile() (string, error) {
	dialog := s.wailsApp.Dialog.OpenFile()
	dialog.CanChooseFiles(true)
	dialog.CanChooseDirectories(false)
	dialog.SetTitle("Select file to backup")
	dialog.AddFilter("All Files", "*.*")
	return dialog.PromptForSingleSelection()
}

func (s *BackupService) SelectRestoreFile() (string, error) {
	dialog := s.wailsApp.Dialog.OpenFile()
	dialog.CanChooseFiles(true)
	dialog.CanChooseDirectories(false)
	dialog.SetTitle("Select share file or .rsmeta")
	dialog.AddFilter("RS Meta", "*.rsmeta")
	dialog.AddFilter("RS Share", "*.rs.*")
	dialog.AddFilter("All Files", "*.*")
	return dialog.PromptForSingleSelection()
}

func (s *BackupService) SelectOutputDirectory() (string, error) {
	dialog := s.wailsApp.Dialog.OpenFile()
	dialog.CanChooseFiles(false)
	dialog.CanChooseDirectories(true)
	dialog.SetTitle("Select output directory")
	return dialog.PromptForSingleSelection()
}

func (s *BackupService) GetBackupWarnings(shares, threshold int) []string {
	return collectWarnings(shares, threshold)
}

func (s *BackupService) RunBackup(inputPath, password, outputDir string, shares, threshold int) error {
	if inputPath == "" {
		return errors.New("input file is required")
	}
	if password == "" {
		return errors.New("password is required")
	}
	return app.RunBackup(app.BackupOptions{
		InputPath: inputPath,
		Shares:    shares,
		Threshold: threshold,
		Password:  password,
		OutputDir: outputDir,
	})
}

func (s *BackupService) RunRestore(inputPath, password, outputDir string) error {
	if inputPath == "" {
		return errors.New("input file is required")
	}
	if password == "" {
		return errors.New("password is required")
	}
	return app.RunRestore(app.RestoreOptions{
		InputPath: inputPath,
		Password:  password,
		OutputDir: outputDir,
	})
}

func (s *BackupService) GetFileInfo(inputPath string) (string, error) {
	info, err := os.Stat(inputPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", errors.New("path is a directory")
	}
	return fmt.Sprintf("Size: %s", formatSize(info.Size())), nil
}

func collectWarnings(shares, threshold int) []string {
	parity := shares - threshold
	var warnings []string

	if parity < 2 || parity*100 < shares*15 {
		warnings = append(warnings,
			fmt.Sprintf("Low redundancy: only %d redundant share(s), losing more than %d share(s) makes recovery impossible", parity, parity))
	}

	expansion := float64(shares) / float64(threshold)
	if expansion > 2.0 {
		warnings = append(warnings,
			"High redundancy: stored data will be ~"+strconv.FormatFloat(expansion, 'f', 2, 64)+"x the encrypted payload")
	}

	return warnings
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func createApp() *application.App {
	svc := NewBackupService()

	wailsApp := application.New(application.Options{
		Name:        "RSBackup",
		Description: "Reed-Solomon Encrypted Backup Tool",
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	svc.wailsApp = wailsApp

	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Reed-Solomon Backup",
		Width:            900,
		Height:           700,
		MinWidth:         700,
		MinHeight:        500,
		BackgroundColour: application.NewRGB(26, 26, 26),
		URL:              "/",
	})

	return wailsApp
}
