package jdextractor

import (
	"strings"
	"testing"
)

func TestBuildJDParserPrompt_ContainsInput(t *testing.T) {
	jdText := "Senior Go Engineer at Acme Corp"

	result := BuildJDParserPrompt(jdText)

	if !strings.Contains(result, jdText) {
		t.Errorf("prompt does not contain the input JD text")
	}
}

func TestBuildJDParserPrompt_ContainsCriticalRules(t *testing.T) {
	result := BuildJDParserPrompt("dummy JD")

	rules := []string{
		"Extract facts only",
		"Never invent information",
		"Use null when information is missing",
		"Output valid JSON only",
		"Do not use markdown",
		"Do not use triple backticks",
		"must_have",
		"nice_to_have",
	}

	for _, rule := range rules {
		if !strings.Contains(result, rule) {
			t.Errorf("prompt missing critical rule: %q", rule)
		}
	}
}

func TestBuildJDParserPrompt_EmptyInput(t *testing.T) {
	result := BuildJDParserPrompt("")

	if !strings.Contains(result, "INPUT JOB DESCRIPTION") {
		t.Errorf("prompt should include INPUT JOB DESCRIPTION section even with empty input")
	}
}

func TestBuildJDParserPrompt_PreservesSpecialChars(t *testing.T) {
	jdText := "C++ & Rust developer with <script>alert('xss')</script> experience"
	result := BuildJDParserPrompt(jdText)

	if !strings.Contains(result, jdText) {
		t.Errorf("prompt should preserve special characters in JD text")
	}
}

func TestBuildJDParserPrompt_NotEmpty(t *testing.T) {
	result := BuildJDParserPrompt("some JD content")

	if len(result) == 0 {
		t.Errorf("prompt should not be empty")
	}
}
