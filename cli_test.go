package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIRiskRequiresInteractiveConfirmation(t *testing.T) {
	stdin := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	writer.Close()
	os.Stdin = reader
	t.Cleanup(func() { os.Stdin = stdin; reader.Close() })
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(input, []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "backup")
	err = runBackup([]string{"--input", input, "--shares", "3", "--threshold", "3", "--out-dir", output})
	if err == nil || !strings.Contains(err.Error(), "confirmation") {
		t.Fatalf("expected confirmation error, got %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("CLI wrote files before risk confirmation")
	}
}

func TestCLIRejectsFilenameEncryptionWithoutEncryption(t *testing.T) {
	if err := runBackup([]string{"--encrypt-filename"}); err == nil {
		t.Fatal("filename encryption accepted without encryption")
	}
}
