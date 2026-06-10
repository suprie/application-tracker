package coverletter

import (
	"strings"
	"testing"
)

func TestRenderLaTeX(t *testing.T) {
	data := TemplateData{
		YourName:       "Jane Doe",
		YourAddress:    "123 Main St, Jakarta",
		YourEmail:      "jane@example.com",
		YourPhone:      "+62 812-3456-7890",
		RecipientName:  "John Smith",
		RecipientTitle: "Hiring Manager",
		CompanyName:    "Acme Corp",
		CompanyAddress: "456 Tech Park, Jakarta",
		Subject:        "Application for Backend Engineer",
		Opening:        "Dear John Smith,",
		BodyParagraphs: []string{"First paragraph.", "Second paragraph."},
		Closing:        "Sincerely,",
	}

	tex, err := RenderLaTeX(data)
	if err != nil {
		t.Fatalf("RenderLaTeX: %v", err)
	}

	checks := []string{
		`\documentclass[11pt,a4paper]{letter}`,
		`\usepackage{lmodern}`,
		`\setstretch{1.2}`,
		`\longindentation=0pt`,
		`\signature{\hspace*{-\parindent}Jane Doe}`,
		"123 Main St, Jakarta",
		"jane@example.com",
		"+62 812-3456-7890",
		"John Smith",
		"Hiring Manager",
		`\textbf{Acme Corp}`,
		`\opening{ Dear John Smith, }`,
		`\closing{ Sincerely, }`,
		"First paragraph.",
		"Second paragraph.",
		"Application for Backend Engineer",
		`\begin{document}`,
		`\end{document}`,
		`\begin{letter}`,
		`\end{letter}`,
	}

	for _, want := range checks {
		if !strings.Contains(tex, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestRenderLaTeX_NoSubject(t *testing.T) {
	data := TemplateData{
		YourName:      "Jane Doe",
		RecipientName: "John Smith",
		CompanyName:   "Acme Corp",
		Opening:       "Dear John,",
		Closing:       "Best,",
	}

	tex, err := RenderLaTeX(data)
	if err != nil {
		t.Fatalf("RenderLaTeX: %v", err)
	}

	// Subject section should not appear.
	if strings.Contains(tex, `\textbf{Re:`) {
		t.Error("subject should not appear when empty")
	}
}

func TestRenderLaTeX_NoOptionalFields(t *testing.T) {
	data := TemplateData{
		YourName:      "Jane Doe",
		RecipientName: "John Smith",
		CompanyName:   "Acme Corp",
		Opening:       "Dear John,",
		Closing:       "Sincerely,",
	}

	tex, err := RenderLaTeX(data)
	if err != nil {
		t.Fatalf("RenderLaTeX: %v", err)
	}

	// Should not produce dangling separators.
	if strings.Contains(tex, `\\ \\`) {
		t.Error("should not have empty address/recipient lines")
	}
}
