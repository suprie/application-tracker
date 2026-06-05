package ats

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type RustPDFExtractor struct {
	BinaryPath string
}

func (e RustPDFExtractor) Extract(ctx context.Context, filePath string) (*ExtractResult, error) {

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		e.BinaryPath,
		filePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run rust pdf extractor: %w", err)
	}

	var result ExtractResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("decode extractor output: %w", err)
	}
	return &result, nil

}
