package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"reed-solomon-backup/internal/app"
)

func main() {
	if len(os.Args) > 1 {
		runCLI()
		return
	}

	wailsApp := createApp()
	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}

func runCLI() {
	var err error
	switch os.Args[1] {
	case "backup":
		err = runBackup(os.Args[2:])
	case "restore":
		err = runRestore(os.Args[2:])
	case "help", "-h", "--help":
		usage()
		return
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}

	if err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}
}

func runBackup(args []string) error {
	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	input := fs.String("input", "", "path to source file")
	shares := fs.Int("shares", 8, "total number of shares (3~128, default: 8)")
	threshold := fs.Int("threshold", 5, "minimum shares required to restore (1~shares, default: 5)")
	password := fs.String("password", "", "backup password (optional, prompt if empty)")
	outDir := fs.String("out-dir", "", "output directory (default: source file directory)")
	encrypt := fs.Bool("encrypt", false, "enable AES-256-GCM encryption")
	encryptFilename := fs.Bool("encrypt-filename", false, "encrypt the original filename (requires --encrypt)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *encryptFilename && !*encrypt {
		return errors.New("--encrypt-filename requires --encrypt")
	}

	return app.RunBackup(app.BackupOptions{
		InputPath:       *input,
		Shares:          *shares,
		Threshold:       *threshold,
		Password:        *password,
		OutputDir:       *outDir,
		Encrypt:         *encrypt,
		EncryptFilename: *encryptFilename,
	})
}

func runRestore(args []string) error {
	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	input := fs.String("input", "", "path to any share file or .rsmeta file")
	password := fs.String("password", "", "backup password (optional, prompt if empty)")
	outDir := fs.String("out-dir", "", "output directory (default: current directory)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	return app.RunRestore(app.RestoreOptions{
		InputPath: *input,
		Password:  *password,
		OutputDir: *outDir,
	})
}

func usage() {
	fmt.Println(`ReSo Backup - Reed-Solomon encrypted backup tool

Usage:
  rsbackup                          Launch GUI (no arguments)
  rsbackup backup  --input <file> [--shares 8] [--threshold 5] [--password <pwd>] [--out-dir <dir>] [--encrypt] [--encrypt-filename]
  rsbackup restore --input <any .rs.NNN or .rsmeta file> [--password <pwd>] [--out-dir <dir>]

Notes:
  - shares must be between 3 and 128
  - threshold must be between 1 and shares
  - risky share/threshold combinations require interactive confirmation
  - --encrypt enables AES-256-GCM encryption (password required)
  - --encrypt-filename encrypts the original filename (requires --encrypt)
  - password can be provided by flag or entered interactively`)
}
