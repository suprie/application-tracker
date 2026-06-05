package textutils

import (
	"regexp"
	"strings"
)

func NormalizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	re := regexp.MustCompile(`\n{3,}`)
	s = re.ReplaceAllString(s, "\n\n")

	return strings.TrimSpace(s)
}
