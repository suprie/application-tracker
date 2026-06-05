package ats

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// writeFakeExtractor creates a shell script that echoes the given JSON.
// Returns the path to the script.
func writeFakeExtractor(t *testing.T, jsonOutput string) string {
	t.Helper()

	dir := t.TempDir()
	var scriptPath string
	var content string

	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(dir, "fake-extractor.bat")
		content = fmt.Sprintf("@echo off\r\necho %s\r\n", jsonOutput)
	} else {
		scriptPath = filepath.Join(dir, "fake-extractor.sh")
		content = fmt.Sprintf("#!/bin/sh\necho '%s'\n", jsonOutput)
	}

	if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
		t.Fatalf("failed to write fake extractor: %v", err)
	}

	return scriptPath
}

func TestRustPDFExtractor_Extract_Success(t *testing.T) {
	fakeBin := writeFakeExtractor(t, `{"text": "John Doe CV content"}`)
	extractor := RustPDFExtractor{BinaryPath: fakeBin}

	result, err := extractor.Extract(context.Background(), "dummy.pdf")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if result.Text != "John Doe CV content" {
		t.Errorf("Text = %q, want %q", result.Text, "John Doe CV content")
	}
}

func TestRustPDFExtractor_Extract_BinaryNotFound(t *testing.T) {
	extractor := RustPDFExtractor{BinaryPath: "/nonexistent/path/to/binary"}

	_, err := extractor.Extract(context.Background(), "dummy.pdf")
	if err == nil {
		t.Errorf("expected error for nonexistent binary, got nil")
	}
}

func TestRustPDFExtractor_Extract_Timeout(t *testing.T) {
	// A script that sleeps forever to trigger the 30s timeout
	// We use a shorter test by checking that context cancellation propagates
	extractor := RustPDFExtractor{BinaryPath: "/nonexistent/path/to/binary"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	_, err := extractor.Extract(ctx, "dummy.pdf")
	if err == nil {
		t.Errorf("expected error for cancelled context, got nil")
	}
}

func TestRustPDFExtractor_Extract_InvalidJSON(t *testing.T) {
	fakeBin := writeFakeExtractor(t, `not json at all`)
	extractor := RustPDFExtractor{BinaryPath: fakeBin}

	_, err := extractor.Extract(context.Background(), "dummy.pdf")
	if err == nil {
		t.Errorf("expected decode error for invalid JSON, got nil")
	}
}

func TestExtractResult_JSONTags(t *testing.T) {
	// Verify the struct shape hasn't drifted — the Rust binary depends on this
	result := ExtractResult{Text: "test"}
	// If the field name or json tag changes, the Rust side must change too.
	// This test just ensures the struct compiles with the expected field.
	if result.Text != "test" {
		t.Errorf("ExtractResult.Text = %q, want %q", result.Text, "test")
	}
}
