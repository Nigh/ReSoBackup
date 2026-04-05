package app

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/crypto/scrypt"
	"golang.org/x/term"

	"reed-solomon-backup/internal/cryptox"
	"reed-solomon-backup/internal/format"
	"reed-solomon-backup/internal/rs"
)

type BackupOptions struct {
	InputPath string
	Shares    int
	Threshold int
	Password  string
	OutputDir string
}

type RestoreOptions struct {
	InputPath string
	Password  string
	OutputDir string
}

func RunBackup(opts BackupOptions) error {
	if err := validateBackupOptions(opts); err != nil {
		return err
	}
	if err := confirmBackupWarnings(opts); err != nil {
		return err
	}

	password, err := readPasswordIfNeeded(opts.Password, "Enter backup password: ")
	if err != nil {
		return err
	}

	plain, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return fmt.Errorf("read input file: %w", err)
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = filepath.Dir(opts.InputPath)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("generate salt: %w", err)
	}
	masterKey, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return fmt.Errorf("derive key: %w", err)
	}

	originalName := filepath.Base(opts.InputPath)
	encryptedName, err := cryptox.EncryptToString(masterKey, []byte(originalName))
	if err != nil {
		return fmt.Errorf("encrypt file name: %w", err)
	}
	prefix := safeName(encryptedName)

	batchID := make([]byte, 16)
	if _, err := rand.Read(batchID); err != nil {
		return fmt.Errorf("generate batch id: %w", err)
	}

	payload, err := cryptox.Encrypt(masterKey, plain)
	if err != nil {
		return fmt.Errorf("encrypt file content: %w", err)
	}

	shares, err := rs.Encode(payload, opts.Shares, opts.Threshold)
	if err != nil {
		return err
	}

	meta := format.Metadata{
		Version:              1,
		BatchID:              base64.RawURLEncoding.EncodeToString(batchID),
		Salt:                 base64.RawURLEncoding.EncodeToString(salt),
		KDF:                  format.KDFInfo{Name: "scrypt", N: 32768, R: 8, P: 1, KeyLen: 32},
		Shares:               opts.Shares,
		Threshold:            opts.Threshold,
		OriginalFileSize:     int64(len(plain)),
		EncryptedFileName:    encryptedName,
		EncryptedPayloadSize: int64(len(payload)),
		Prefix:               prefix,
	}

	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	metaPath := filepath.Join(outputDir, prefix+".rsmeta")
	if err := os.WriteFile(metaPath, metaBytes, 0o600); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}

	for i, shard := range shares {
		shareFile := filepath.Join(outputDir, fmt.Sprintf("%s.rs.%03d", prefix, i+1))
		wrapped, err := wrapShard(batchID, i, opts.Shares, opts.Threshold, shard)
		if err != nil {
			return err
		}
		if err := os.WriteFile(shareFile, wrapped, 0o600); err != nil {
			return fmt.Errorf("write share %d: %w", i+1, err)
		}
	}

	fmt.Printf("Backup created successfully.\nMetadata: %s\nShares prefix: %s\n", metaPath, filepath.Join(outputDir, prefix+".rs.###"))
	return nil
}

func RunRestore(opts RestoreOptions) error {
	if opts.InputPath == "" {
		return errors.New("--input is required")
	}
	password, err := readPasswordIfNeeded(opts.Password, "Enter restore password: ")
	if err != nil {
		return err
	}

	metaPath, prefix, dir, err := discoverSet(opts.InputPath)
	if err != nil {
		return err
	}

	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}
	var meta format.Metadata
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return fmt.Errorf("parse metadata: %w", err)
	}

	salt, err := base64.RawURLEncoding.DecodeString(meta.Salt)
	if err != nil {
		return fmt.Errorf("decode salt: %w", err)
	}
	masterKey, err := scrypt.Key([]byte(password), salt, meta.KDF.N, meta.KDF.R, meta.KDF.P, meta.KDF.KeyLen)
	if err != nil {
		return fmt.Errorf("derive key: %w", err)
	}

	originalNameBytes, err := cryptox.DecryptString(masterKey, meta.EncryptedFileName)
	if err != nil {
		return errors.New("invalid password or corrupted metadata")
	}
	originalName := string(originalNameBytes)

	shardFiles, err := filepath.Glob(filepath.Join(dir, prefix+".rs.*"))
	if err != nil {
		return fmt.Errorf("find share files: %w", err)
	}
	sort.Strings(shardFiles)

	shards := make([][]byte, meta.Shares)
	available := 0
	for _, path := range shardFiles {
		if strings.HasSuffix(path, ".rsmeta") {
			continue
		}
		idx, data, hdr, err := unwrapShard(path)
		if err != nil {
			continue
		}
		if int(hdr.TotalShares) != meta.Shares || int(hdr.Threshold) != meta.Threshold {
			continue
		}
		if idx >= 0 && idx < meta.Shares && shards[idx] == nil {
			shards[idx] = data
			available++
		}
	}
	if available < meta.Threshold {
		return fmt.Errorf("not enough shares: have %d, need %d", available, meta.Threshold)
	}

	payload, err := rs.Reconstruct(shards, meta.Shares, meta.Threshold, meta.EncryptedPayloadSize)
	if err != nil {
		return err
	}
	plain, err := cryptox.Decrypt(masterKey, payload)
	if err != nil {
		return errors.New("failed to decrypt payload, password may be incorrect or shares corrupted")
	}
	if int64(len(plain)) != meta.OriginalFileSize {
		plain = plain[:min(len(plain), int(meta.OriginalFileSize))]
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = "."
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	outPath := filepath.Join(outputDir, filepath.Base(originalName))
	if err := os.WriteFile(outPath, plain, 0o600); err != nil {
		return fmt.Errorf("write restored file: %w", err)
	}

	fmt.Printf("File restored successfully: %s\n", outPath)
	return nil
}

func validateBackupOptions(opts BackupOptions) error {
	if opts.InputPath == "" {
		return errors.New("--input is required")
	}
	if opts.Shares < 3 || opts.Shares > 128 {
		return errors.New("--shares must be in range [3, 128]")
	}
	if opts.Threshold < 1 {
		return errors.New("--threshold must be >= 1")
	}
	if opts.Threshold > opts.Shares {
		return errors.New("--threshold must be <= shares")
	}
	info, err := os.Stat(opts.InputPath)
	if err != nil {
		return fmt.Errorf("stat input file: %w", err)
	}
	if info.IsDir() {
		return errors.New("input path must be a file, not directory")
	}
	return nil
}

func confirmBackupWarnings(opts BackupOptions) error {
	warnings := collectBackupWarnings(opts)
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

func collectBackupWarnings(opts BackupOptions) []string {
	parity := opts.Shares - opts.Threshold
	var warnings []string

	if parity < 2 || parity*100 < opts.Shares*15 {
		warnings = append(warnings,
			fmt.Sprintf("low redundancy: only %d redundant share(s), so losing more than %d share(s) will make recovery impossible", parity, parity))
	}

	expansion := float64(opts.Shares) / float64(opts.Threshold)
	if expansion > 2.0 {
		warnings = append(warnings,
			"high redundancy: total stored share data will be about "+strconv.FormatFloat(expansion, 'f', 2, 64)+"x the encrypted payload size")
	}

	return warnings
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

func safeName(s string) string {
	s = strings.TrimRight(s, "=")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "+", "-")
	return s
}

func discoverSet(input string) (metaPath, prefix, dir string, err error) {
	abs, err := filepath.Abs(input)
	if err != nil {
		return "", "", "", err
	}
	dir = filepath.Dir(abs)
	base := filepath.Base(abs)
	if strings.HasSuffix(base, ".rsmeta") {
		prefix = strings.TrimSuffix(base, ".rsmeta")
		return abs, prefix, dir, nil
	}
	idx := strings.Index(base, ".rs.")
	if idx < 0 {
		return "", "", "", errors.New("input must be a .rsmeta or .rs.NNN file")
	}
	prefix = base[:idx]
	metaPath = filepath.Join(dir, prefix+".rsmeta")
	if _, err := os.Stat(metaPath); err != nil {
		return "", "", "", fmt.Errorf("metadata file not found: %w", err)
	}
	return metaPath, prefix, dir, nil
}

type shardHeader struct {
	Magic       [4]byte
	Version     uint8
	TotalShares uint16
	Threshold   uint16
	Index       uint16
	BatchID     [16]byte
	DataSize    uint32
}

func wrapShard(batchID []byte, index, shares, threshold int, data []byte) ([]byte, error) {
	var h shardHeader
	copy(h.Magic[:], []byte("RSBK"))
	h.Version = 1
	h.TotalShares = uint16(shares)
	h.Threshold = uint16(threshold)
	h.Index = uint16(index)
	copy(h.BatchID[:], batchID)
	h.DataSize = uint32(len(data))

	out := make([]byte, 31+len(data))
	copy(out[:4], h.Magic[:])
	out[4] = h.Version
	binary.BigEndian.PutUint16(out[5:7], h.TotalShares)
	binary.BigEndian.PutUint16(out[7:9], h.Threshold)
	binary.BigEndian.PutUint16(out[9:11], h.Index)
	copy(out[11:27], h.BatchID[:])
	binary.BigEndian.PutUint32(out[27:31], h.DataSize)
	copy(out[31:], data)
	return out, nil
}

func unwrapShard(path string) (int, []byte, shardHeader, error) {
	var h shardHeader
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, h, err
	}
	if len(b) < 31 {
		return 0, nil, h, io.ErrUnexpectedEOF
	}
	copy(h.Magic[:], b[:4])
	if string(h.Magic[:]) != "RSBK" {
		return 0, nil, h, errors.New("invalid shard magic")
	}
	h.Version = b[4]
	h.TotalShares = binary.BigEndian.Uint16(b[5:7])
	h.Threshold = binary.BigEndian.Uint16(b[7:9])
	h.Index = binary.BigEndian.Uint16(b[9:11])
	copy(h.BatchID[:], b[11:27])
	h.DataSize = binary.BigEndian.Uint32(b[27:31])
	if len(b) < 31+int(h.DataSize) {
		return 0, nil, h, io.ErrUnexpectedEOF
	}
	return int(h.Index), b[31 : 31+int(h.DataSize)], h, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
