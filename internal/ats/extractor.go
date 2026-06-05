package ats

import (
	"context"
)

type ExtractResult struct {
	Text  string `json:"text"`
}

type CVExtractor interface {
	Extract(ctx context.Context, filePath string) (*ExtractResult, error)
}
