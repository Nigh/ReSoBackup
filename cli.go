package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
	"reed-solomon-backup/internal/app"
)

func confirmBackupWarnings(opts app.BackupOptions) error {
	warnings := app.BackupWarnings(opts)
	if len(warnings) == 0 {
		return nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("backup parameters require confirmation in interactive mode: %s", strings.Join(warnings, "; "))
	}

	fmt.Fprintln(os.Stderr, "Warnings:")
	for _, warning := range warnings {
		fmt.Fprintf(os.Stderr, "- %s\n", warning)
	}
	fmt.Fprint(os.Stderr, "Continue anyway? [y/N]: ")

	var input string
	if _, err := fmt.Fscanln(os.Stdin, &input); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("backup cancelled: confirmation declined")
		}
		return fmt.Errorf("read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(input))
	if answer != "y" && answer != "yes" {
		return errors.New("backup cancelled: confirmation declined")
	}
	return nil
}

func readPasswordIfNeeded(pwd, prompt string) (string, error) {
	if pwd != "" {
		return pwd, nil
	}
	fmt.Fprint(os.Stderr, prompt)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	if len(b) == 0 {
		return "", errors.New("password cannot be empty")
	}
	return string(b), nil
}
