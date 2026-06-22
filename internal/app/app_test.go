package app

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestWrapUnwrapShardRoundtrip(t *testing.T) {
	batchID := make([]byte, 16)
	for i := range batchID {
		batchID[i] = byte(i)
	}
	data := []byte("test shard data for wrap/unwrap roundtrip")

	wrapped, err := wrapShard(batchID, 3, 8, 5, data)
	if err != nil {
		t.Fatalf("wrapShard: %v", err)
	}

	tmpFile := t.TempDir() + "/test.rs.004"
	if err := os.WriteFile(tmpFile, wrapped, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	idx, got, hdr, err := unwrapShard(tmpFile)
	if err != nil {
		t.Fatalf("unwrapShard: %v", err)
	}

	if idx != 3 {
		t.Fatalf("expected index 3, got %d", idx)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("data mismatch: got %q, want %q", got, data)
	}
	if string(hdr.Magic[:]) != "RSBK" {
		t.Fatalf("magic mismatch: got %q", string(hdr.Magic[:]))
	}
	if hdr.TotalShares != 8 {
		t.Fatalf("TotalShares mismatch: got %d", hdr.TotalShares)
	}
	if hdr.Threshold != 5 {
		t.Fatalf("Threshold mismatch: got %d", hdr.Threshold)
	}
}

func TestUnwrapShardInvalidMagic(t *testing.T) {
	tmpFile := t.TempDir() + "/bad.rs.001"
	badData := make([]byte, 40)
	copy(badData[:4], "BAAD")
	if err := os.WriteFile(tmpFile, badData, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, _, _, err := unwrapShard(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid magic")
	}
}

func TestUnwrapShardTooShort(t *testing.T) {
	tmpFile := t.TempDir() + "/short.rs.001"
	if err := os.WriteFile(tmpFile, []byte{1, 2, 3}, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, _, _, err := unwrapShard(tmpFile)
	if err == nil {
		t.Fatal("expected error for short shard")
	}
}

func TestWrapShardHeader(t *testing.T) {
	batchID := make([]byte, 16)
	data := []byte("header test")

	wrapped, err := wrapShard(batchID, 2, 10, 4, data)
	if err != nil {
		t.Fatalf("wrapShard: %v", err)
	}

	if string(wrapped[:4]) != "RSBK" {
		t.Fatalf("magic: got %q", string(wrapped[:4]))
	}
	if wrapped[4] != 1 {
		t.Fatalf("version: got %d", wrapped[4])
	}
	totalShares := binary.BigEndian.Uint16(wrapped[5:7])
	if totalShares != 10 {
		t.Fatalf("TotalShares: got %d", totalShares)
	}
	threshold := binary.BigEndian.Uint16(wrapped[7:9])
	if threshold != 4 {
		t.Fatalf("Threshold: got %d", threshold)
	}
	index := binary.BigEndian.Uint16(wrapped[9:11])
	if index != 2 {
		t.Fatalf("Index: got %d", index)
	}
	dataSize := binary.BigEndian.Uint32(wrapped[27:31])
	if dataSize != uint32(len(data)) {
		t.Fatalf("DataSize: got %d, want %d", dataSize, len(data))
	}
	if !bytes.Equal(wrapped[31:], data) {
		t.Fatal("data content mismatch")
	}
}

func TestDiscoverSetRsmeta(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/test.rsmeta", []byte("{}"), 0o600)

	metaPath, prefix, _, err := discoverSet(dir + "/test.rsmeta")
	if err != nil {
		t.Fatalf("discoverSet: %v", err)
	}
	if prefix != "test" {
		t.Fatalf("prefix: got %q", prefix)
	}
	if filepath.Base(metaPath) != "test.rsmeta" {
		t.Fatalf("metaPath: got %q", metaPath)
	}
}

func TestDiscoverSetRsFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/myfile.rsmeta", []byte("{}"), 0o600)

	metaPath, prefix, _, err := discoverSet(dir + "/myfile.rs.001")
	if err != nil {
		t.Fatalf("discoverSet: %v", err)
	}
	if prefix != "myfile" {
		t.Fatalf("prefix: got %q", prefix)
	}
	if filepath.Base(metaPath) != "myfile.rsmeta" {
		t.Fatalf("metaPath: got %q", metaPath)
	}
}

func TestDiscoverSetInvalidFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/test.txt", []byte(""), 0o600)

	_, _, _, err := discoverSet(dir + "/test.txt")
	if err == nil {
		t.Fatal("expected error for non .rsmeta/.rs.NNN file")
	}
}

func TestValidateBackupOptions(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/input.txt", []byte("test"), 0o600)

	tests := []struct {
		name    string
		opts    BackupOptions
		wantErr bool
	}{
		{"valid", BackupOptions{InputPath: dir + "/input.txt", Shares: 8, Threshold: 5}, false},
		{"no input", BackupOptions{Shares: 8, Threshold: 5}, true},
		{"shares too low", BackupOptions{InputPath: dir + "/input.txt", Shares: 2, Threshold: 1}, true},
		{"shares too high", BackupOptions{InputPath: dir + "/input.txt", Shares: 129, Threshold: 1}, true},
		{"threshold zero", BackupOptions{InputPath: dir + "/input.txt", Shares: 8, Threshold: 0}, true},
		{"threshold > shares", BackupOptions{InputPath: dir + "/input.txt", Shares: 5, Threshold: 6}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBackupOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateBackupOptions: err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestSafeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc", "abc"},
		{"abc=", "abc"},
		{"a/b+c", "a_b-c"},
		{"abc===", "abc"},
	}
	for _, tt := range tests {
		got := safeName(tt.input)
		if got != tt.want {
			t.Errorf("safeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal.txt", "normal.txt"},
		{"path/to\\file:name", "path_to_file_name"},
		{"file*?\"<>|name", "file______name"},
		{"file with spaces", "file_with_spaces"},
	}
	for _, tt := range tests {
		got := sanitizeFilename(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
