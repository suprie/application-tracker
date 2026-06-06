package cvparser

import (
	"strings"
	"testing"
)

func TestBuildCVParserPrompt(t *testing.T) {
	cvText := "John Doe\nSoftware Engineer"

	result := BuildCVParserPrompt(cvText)

	// The prompt must include the CV text
	if !strings.Contains(result, cvText) {
		t.Errorf("prompt does not contain the CV text")
	}

	// The prompt must include the expected schema keys
	requiredKeys := []string{
		"name:",
		"headline:",
		"location:",
		"email:",
		"phone:",
		"linkedin:",
		"summary:",
		"total_years_experience:",
		"domains:",
		"skills:",
		"  languages:",
		"  mobile:",
		"  architecture:",
		"  testing:",
		"  ci_cd:",
		"  tools:",
		"  leadership:",
		"experience:",
		"  -",
		"    title:",
		"    company:",
		"    start_date:",
		"    end_date:",
		"education:",
		"    school:",
		"    degree:",
		"parsing_warnings:",
	}

	for _, key := range requiredKeys {
		if !strings.Contains(result, key) {
			t.Errorf("prompt is missing schema key: %q", key)
		}
	}

	// The prompt must include the core rules
	if !strings.Contains(result, "Extract facts only") {
		t.Errorf("prompt missing rule: Extract facts only")
	}
	if !strings.Contains(result, "Never invent information") {
		t.Errorf("prompt missing rule: Never invent information")
	}
	if !strings.Contains(result, "Use null when missing") {
		t.Errorf("prompt missing rule: Use null when missing")
	}
	if !strings.Contains(result, "Output YAML") {
		t.Errorf("prompt missing rule: Output YAML")
	}
}

func TestBuildCVParserPrompt_EmptyCV(t *testing.T) {
	result := BuildCVParserPrompt("")

	if !strings.Contains(result, "CV TEXT:") {
		t.Errorf("prompt should include CV TEXT section even with empty input")
	}
}

func TestBuildCVParserPrompt_PreservesSpecialChars(t *testing.T) {
	cvText := "C++ developer with <script>alert('xss')</script> experience"
	result := BuildCVParserPrompt(cvText)

	if !strings.Contains(result, cvText) {
		t.Errorf("prompt should preserve special characters in CV text")
	}
}
