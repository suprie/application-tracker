package textutils

import (
	"regexp"
	"strings"
)

var multiNewLineRe = regexp.MustCompile(`\n{3,}`)

func NormalizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = multiNewLineRe.ReplaceAllString(s, "\n\n")

	return strings.TrimSpace(s)
}
