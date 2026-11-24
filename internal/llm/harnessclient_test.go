//go:build !windows

package llm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHarnessGenerate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	for _, cli := range []string{"claude", "codex"} {
		t.Run(cli, func(t *testing.T) {
			stub := "#!/bin/sh\n[ -z \"$(ls -A)\" ] || exit 1\ncat\n"
			if err := os.WriteFile(filepath.Join(dir, cli), []byte(stub), 0700); err != nil {
				t.Fatal(err)
			}
			format := &ResponseFormat{JSONSchema: JSONSchema{Schema: map[string]any{"type": "object"}}}
			got, err := NewHarnessClient(cli, "match").Generate(t.Context(), "test prompt", format)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(got, "test prompt\n") || !strings.HasSuffix(got, `{"type":"object"}`) {
				t.Errorf("unexpected stdin: %q", got)
			}
		})
	}
}
