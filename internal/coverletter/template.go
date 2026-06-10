package coverletter

import (
	"strings"
	"text/template"
)

// TemplateData holds all variables for the LaTeX template.
type TemplateData struct {
	YourName       string
	YourAddress    string
	YourEmail      string
	YourPhone      string
	RecipientName  string
	RecipientTitle string
	CompanyName    string
	CompanyAddress string
	Subject        string
	Opening        string // salutation only, e.g. "Dear Hiring Manager,"
	BodyParagraphs []string
	Closing        string // closing phrase only, e.g. "Sincerely,"
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

// recipientBlock builds the recipient info for \begin{letter}{...}.
func (d TemplateData) recipientBlock() string {
	parts := []string{d.RecipientName}
	if d.RecipientTitle != "" {
		parts = append(parts, d.RecipientTitle)
	}
	parts = append(parts, `\textbf{`+d.CompanyName+`}`)
	if d.CompanyAddress != "" {
		parts = append(parts, d.CompanyAddress)
	}
	return strings.Join(parts, " \\\\ ")
}

const latexTemplate = `\documentclass[11pt,a4paper]{letter}
\usepackage[margin=1in]{geometry}
\usepackage{setspace}
\usepackage{parskip}
\usepackage{lmodern}

\signature{\hspace*{-\parindent}{{.YourName}}}
\address{ {{senderAddress}} }
\date{\today}

\setstretch{1.2}
\longindentation=0pt

\begin{document}

\begin{letter}{ {{recipientBlock}} }

{{if .Subject}}
\begin{center}
\textbf{Re: {{.Subject}}}
\end{center}
\vspace{0.5em}
{{end}}

\opening{ {{.Opening}} }

{{range .BodyParagraphs}}
{{.}}

{{end}}
\closing{ {{.Closing}} }

\end{letter}
\end{document}
`

// RenderLaTeX renders the LaTeX template with the given data.
func RenderLaTeX(data TemplateData) (string, error) {
	tmpl, err := template.New("coverletter").Funcs(template.FuncMap{
		"senderAddress":  data.senderAddress,
		"recipientBlock": data.recipientBlock,
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
