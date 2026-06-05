package textutils

import (
	"testing"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "converts Windows line endings",
			input: "line1\r\nline2\r\nline3",
			want:  "line1\nline2\nline3",
		},
		{
			name:  "collapses three or more newlines to two",
			input: "para1\n\n\n\npara2\n\n\npara3",
			want:  "para1\n\npara2\n\npara3",
		},
		{
			name:  "trims leading and trailing whitespace",
			input: "  \n  hello world  \n  ",
			want:  "hello world",
		},
		{
			name:  "handles empty string",
			input: "",
			want:  "",
		},
		{
			name:  "handles whitespace-only string",
			input: "   \n\n\n   ",
			want:  "",
		},
		{
			name:  "keeps single newline intact",
			input: "line1\nline2",
			want:  "line1\nline2",
		},
		{
			name:  "keeps double newline (paragraph break) intact",
			input: "para1\n\npara2",
			want:  "para1\n\npara2",
		},
		{
			name:  "mixed Windows and Unix endings",
			input: "a\r\n\r\n\r\n\r\nb",
			want:  "a\n\nb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeText(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
