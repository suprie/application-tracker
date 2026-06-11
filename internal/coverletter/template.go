package coverletter

import (
	"strings"
	"text/template"
)

// latexSpecialChars are characters that must be escaped in LaTeX text.
var latexSpecialChars = strings.NewReplacer(
	`\`, `\textbackslash{}`,
	`{`, `\{`,
	`}`, `\}`,
	`$`, `\$`,
	`&`, `\&`,
	`#`, `\#`,
	`_`, `\_`,
	`%`, `\%`,
	`~`, `\textasciitilde{}`,
	`^`, `\textasciicircum{}`,
)

// escapeLaTeX escapes special characters for use in LaTeX body text.
func escapeLaTeX(s string) string {
	return latexSpecialChars.Replace(s)
}

// TemplateData holds all variables for the LaTeX template.
type TemplateData struct {
	YourName          string
	YourAddress       string
	YourEmail         string
	YourPhone         string
	RecipientName     string
	RecipientTitle    string
	CompanyName       string
	CompanyAddress    string
	Subject           string
	Opening           string   // salutation only, e.g. "Dear Hiring Manager,"
	OpeningParagraphs string   // single intro paragraph
	BodyParagraphs    []string // 2-3 body paragraphs
	ClosingParagraphs string   // single closing paragraph
	Closing           string   // closing phrase only, e.g. "Sincerely,"
}

// senderAddress builds the \\-separated address block for \address{}.
func (d TemplateData) senderAddress() string {
	var parts []string
	if d.YourAddress != "" {
		parts = append(parts, d.YourAddress)
	}
	if d.YourEmail != "" {
		parts = append(parts, d.YourEmail)
	}
	if d.YourPhone != "" {
		parts = append(parts, d.YourPhone)
	}
	return strings.Join(parts, " \\\\ ")
}

// hasAddress returns true if any address field is populated.
func (d TemplateData) hasAddress() bool {
	return d.YourAddress != "" || d.YourEmail != "" || d.YourPhone != ""
}

// recipientBlock builds the recipient info for \begin{letter}{...}.
func (d TemplateData) recipientBlock() string {
	var parts []string
	if d.RecipientName != "" {
		parts = append(parts, d.RecipientName)
	}
	if d.RecipientTitle != "" {
		parts = append(parts, d.RecipientTitle)
	}
	if d.CompanyName != "" {
		parts = append(parts, `\textbf{`+d.CompanyName+`}`)
	}
	if d.CompanyAddress != "" {
		parts = append(parts, d.CompanyAddress)
	}
	if len(parts) == 0 {
		return "Hiring Manager"
	}
	return strings.Join(parts, " \\\\ ")
}

// safeOpening returns Opening or a fallback.
func (d TemplateData) safeOpening() string {
	s := strings.TrimSpace(d.Opening)
	if s == "" {
		return "Dear Hiring Manager,"
	}
	return s
}

// safeClosing returns Closing or a fallback.
func (d TemplateData) safeClosing() string {
	s := strings.TrimSpace(d.Closing)
	if s == "" {
		return "Sincerely,"
	}
	return s
}

const latexTemplate = `\documentclass[11pt,a4paper]{letter}
\usepackage[margin=1in]{geometry}
\usepackage{setspace}
\usepackage{parskip}
\usepackage{lmodern}

{{if .YourName}}\signature{\hspace*{-\parindent}{{.YourName}}}{{end}}
\address{ {{senderAddress}} }
\date{\today}

\setstretch{1.2}
\longindentation=0pt

\begin{document}

\begin{letter}{ {{recipientBlock}} }

{{if .Subject}}
\begin{center}
\textbf{Re: {{.Subject | escapeLaTeX}}}
\end{center}
\vspace{0.5em}
{{end}}

\opening{ {{safeOpening}} }

{{if .OpeningParagraphs}}{{.OpeningParagraphs | escapeLaTeX}}

{{end}}
{{range .BodyParagraphs}}
{{. | escapeLaTeX}}

{{end}}
{{if .ClosingParagraphs}}{{.ClosingParagraphs | escapeLaTeX}}

{{end}}
\closing{ {{safeClosing}} }

\end{letter}
\end{document}
`

// RenderLaTeX renders the LaTeX template with the given data.
func RenderLaTeX(data TemplateData) (string, error) {
	tmpl, err := template.New("coverletter").Funcs(template.FuncMap{
		"senderAddress":  data.senderAddress,
		"recipientBlock": data.recipientBlock,
		"hasAddress":     data.hasAddress,
		"safeOpening":    data.safeOpening,
		"safeClosing":    data.safeClosing,
		"escapeLaTeX":    escapeLaTeX,
	}).Parse(latexTemplate)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
