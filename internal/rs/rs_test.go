package rs

import (
	"bytes"
	"testing"
)

func TestEncodeReconstructRoundtrip(t *testing.T) {
	data := []byte("hello, Reed-Solomon! This is a test of erasure coding. 1234567890 abcdefghij.")
	totalShares := 8
	threshold := 5

	shards, err := Encode(data, totalShares, threshold)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	if len(shards) != totalShares {
		t.Fatalf("expected %d shards, got %d", totalShares, len(shards))
	}

	got, err := Reconstruct(shards, totalShares, threshold, int64(len(data)))
	if err != nil {
		t.Fatalf("Reconstruct: %v", err)
	}

	if !bytes.Equal(got, data) {
		t.Fatalf("roundtrip failed: got %q, want %q", got, data)
	}
}

func TestReconstructWithMissingShards(t *testing.T) {
	data := []byte("test data for Reed-Solomon with missing shards recovery")
	totalShares := 6
	threshold := 3

	shards, err := Encode(data, totalShares, threshold)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// Remove 3 shards (keep only threshold count)
	shards[0] = nil
	shards[2] = nil
	shards[4] = nil

	got, err := Reconstruct(shards, totalShares, threshold, int64(len(data)))
	if err != nil {
		t.Fatalf("Reconstruct with missing shards: %v", err)
	}

	if !bytes.Equal(got, data) {
		t.Fatalf("roundtrip with missing shards failed: got %q, want %q", got, data)
	}
}

func TestEncodeInvalidParams(t *testing.T) {
	data := []byte("test")
	_, err := Encode(data, 2, 5) // threshold > totalShares
	if err == nil {
		t.Fatal("expected error for threshold > totalShares")
	}
}

func TestReconstructAllShards(t *testing.T) {
	data := []byte("all shards present reconstruction test data 1234567890")
	totalShares := 4
	threshold := 3

	shards, err := Encode(data, totalShares, threshold)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, err := Reconstruct(shards, totalShares, threshold, int64(len(data)))
	if err != nil {
		t.Fatalf("Reconstruct: %v", err)
	}

	if !bytes.Equal(got, data) {
		t.Fatalf("roundtrip failed: got %q, want %q", got, data)
	}
}
