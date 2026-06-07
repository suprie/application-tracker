package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStripMarkdownFence(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strips yaml fence",
			input: "```yaml\nname: John\n```",
			want:  "name: John",
		},
		{
			name:  "strips yml fence",
			input: "```yml\nname: John\n```",
			want:  "name: John",
		},
		{
			name:  "strips bare fence",
			input: "```\nname: John\n```",
			want:  "name: John",
		},
		{
			name:  "no fence returns as-is",
			input: "name: John",
			want:  "name: John",
		},
		{
			name:  "strips trailing fence only",
			input: "name: John\n```",
			want:  "name: John",
		},
		{
			name:  "strips leading fence only",
			input: "```\nname: John",
			want:  "name: John",
		},
		{
			name:  "trims surrounding whitespace",
			input: "  \n  name: John  \n  ",
			want:  "name: John",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "fence with whitespace around content",
			input: "```yaml\n  name: John  \n```",
			want:  "name: John",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripMarkdownFence(tt.input)
			if got != tt.want {
				t.Errorf("StripMarkdownFence(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSaveYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "test.yaml")
	content := "name: John Doe"

	err := SaveYAML(path, content)
	if err != nil {
		t.Fatalf("SaveYAML failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read saved file: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content = %q, want %q", string(data), content)
	}
}

func TestSaveYAML_DirPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "test.yaml")

	err := SaveYAML(path, "content")
	if err != nil {
		t.Fatalf("SaveYAML failed: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "subdir"))
	if err != nil {
		t.Fatalf("cannot stat created dir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected a directory at subdir path")
	}
}
