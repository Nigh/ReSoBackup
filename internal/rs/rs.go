package rs

import (
	"bytes"
	"fmt"

	"github.com/klauspost/reedsolomon"
)

func Encode(data []byte, totalShares, threshold int) ([][]byte, error) {
	parity := totalShares - threshold
	enc, err := reedsolomon.New(threshold, parity)
	if err != nil {
		return nil, fmt.Errorf("create reed-solomon encoder: %w", err)
	}
	shards, err := enc.Split(data)
	if err != nil {
		return nil, fmt.Errorf("split data into shards: %w", err)
	}
	if err := enc.Encode(shards); err != nil {
		return nil, fmt.Errorf("encode shards: %w", err)
	}
	return shards, nil
}

func Reconstruct(shards [][]byte, totalShares, threshold int, payloadSize int64) ([]byte, error) {
	parity := totalShares - threshold
	enc, err := reedsolomon.New(threshold, parity)
	if err != nil {
		return nil, fmt.Errorf("create reed-solomon encoder: %w", err)
	}
	if err := enc.Reconstruct(shards); err != nil {
		return nil, fmt.Errorf("reconstruct shards: %w", err)
	}
	var buf bytes.Buffer
	if err := enc.Join(&buf, shards, int(payloadSize)); err != nil {
		return nil, fmt.Errorf("join shards: %w", err)
	}
	return buf.Bytes(), nil
}
