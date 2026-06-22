package main

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/wailsapp/wails/v3/pkg/application"

	"reed-solomon-backup/internal/app"
)

//go:embed all:frontend/dist
var assets embed.FS

type BackupService struct {
	wailsApp      *application.App
	lastPrefix    string
	lastEncrypted bool
	lastEncFN     bool
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

func (s *BackupService) RunBackup(inputPath, password, outputDir string, shares, threshold int, encrypt, encryptFilename bool) error {
	if inputPath == "" {
		return errors.New("input file is required")
	}
	if encrypt && password == "" {
		return errors.New("password is required when encryption is enabled")
	}
	err := app.RunBackup(app.BackupOptions{
		InputPath:       inputPath,
		Shares:          shares,
		Threshold:       threshold,
		Password:        password,
		OutputDir:       outputDir,
		Encrypt:         encrypt,
		EncryptFilename: encryptFilename,
	})
	if err == nil {
		s.lastEncrypted = encrypt
		s.lastEncFN = encrypt && encryptFilename
	}
	return err
}

func (s *BackupService) GetLastBackupPrefix() string {
	return s.lastPrefix
}

func (s *BackupService) ValidateRestoreFile(inputPath string) (bool, error) {
	if inputPath == "" {
		return false, errors.New("input file is required")
	}
	meta, err := app.ReadMetadata(inputPath)
	if err != nil {
		return false, fmt.Errorf("invalid backup file: %w", err)
	}
	isEncrypted := meta.Encrypted || meta.Version <= 1
	return isEncrypted, nil
}

func (s *BackupService) RunRestore(inputPath, password, outputDir string) error {
	if inputPath == "" {
		return errors.New("input file is required")
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
		Name:        "ReSo Backup",
		Description: "ReSo Backup - Reed-Solomon Encrypted Backup Tool",
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
		Title:            "ReSo Backup",
		Width:            900,
		Height:           700,
		MinWidth:         700,
		MinHeight:        500,
		BackgroundColour: application.NewRGB(26, 26, 26),
		URL:              "/",
	})

	return wailsApp
}
