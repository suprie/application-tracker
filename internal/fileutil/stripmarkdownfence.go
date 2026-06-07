package fileutil

import (
	"strings"
)

func StripMarkdownFence(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```yaml")
	s = strings.TrimPrefix(s, "```yml")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
