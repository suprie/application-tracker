package matcher

import (
	"fmt"
	"strings"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/rag"
)

// BuildMatchPrompt creates the LLM prompt for fit matching a master profile
// against a job description.
func BuildMatchPrompt(masterProfile string, jd *domain.JobDescription) string {
	var b strings.Builder

	b.WriteString("You are an ATS (Applicant Tracking System) semantic matching engine.\n")
	b.WriteString("Compare the candidate's master profile against the job description below.\n")
	b.WriteString("Produce a fit assessment as a JSON object.\n\n")

	b.WriteString("=== CANDIDATE MASTER PROFILE ===\n")
	b.WriteString(masterProfile)
	b.WriteString("\n\n")

	fmt.Fprintf(&b, "=== JOB DESCRIPTION ===\n")
	if jd.Company != nil {
		fmt.Fprintf(&b, "Company: %s\n", *jd.Company)
	}
	if jd.RoleTitle != nil {
		fmt.Fprintf(&b, "Role: %s\n", *jd.RoleTitle)
	}
	if jd.Seniority != nil {
		fmt.Fprintf(&b, "Seniority: %s\n", *jd.Seniority)
	}
	if jd.EmploymentType != nil {
		fmt.Fprintf(&b, "Employment Type: %s\n", *jd.EmploymentType)
	}
	if jd.WorkArrangement != nil {
		fmt.Fprintf(&b, "Work Arrangement: %s\n", *jd.WorkArrangement)
	}
	if jd.Location != nil {
		fmt.Fprintf(&b, "Location: %s\n", *jd.Location)
	}
	b.WriteString("\n")
	b.WriteString("Requirements:\n")
	b.WriteString(jd.RequirementsJSON)
	b.WriteString("\n\n")
	b.WriteString("Responsibilities:\n")
	b.WriteString(jd.ResponsibilitiesJSON)
	b.WriteString("\n\n")
	b.WriteString("Keywords:\n")
	b.WriteString(jd.KeywordsJSON)
	b.WriteString("\n")

	return b.String()
}

// BuildMatchPromptWithChunks creates the LLM prompt for fit matching using
// RAG-retrieved profile chunks instead of the full master profile. Only the
// chunks most relevant to the JD are included, keeping the prompt focused.
func BuildMatchPromptWithChunks(chunks []rag.Chunk, jd *domain.JobDescription) string {
	var b strings.Builder

	b.WriteString("You are an ATS (Applicant Tracking System) semantic matching engine.\n")
	b.WriteString("Compare the candidate's relevant experience against the job description below.\n")
	b.WriteString("Produce a fit assessment as a JSON object.\n\n")

	b.WriteString("=== RETRIEVED RELEVANT EXPERIENCE ===\n")
	for _, c := range chunks {
		b.WriteString(formatChunk(c))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "=== JOB DESCRIPTION ===\n")
	if jd.Company != nil {
		fmt.Fprintf(&b, "Company: %s\n", *jd.Company)
	}
	if jd.RoleTitle != nil {
		fmt.Fprintf(&b, "Role: %s\n", *jd.RoleTitle)
	}
	if jd.Seniority != nil {
		fmt.Fprintf(&b, "Seniority: %s\n", *jd.Seniority)
	}
	if jd.EmploymentType != nil {
		fmt.Fprintf(&b, "Employment Type: %s\n", *jd.EmploymentType)
	}
	if jd.WorkArrangement != nil {
		fmt.Fprintf(&b, "Work Arrangement: %s\n", *jd.WorkArrangement)
	}
	if jd.Location != nil {
		fmt.Fprintf(&b, "Location: %s\n", *jd.Location)
	}
	b.WriteString("\n")
	b.WriteString("Requirements:\n")
	b.WriteString(jd.RequirementsJSON)
	b.WriteString("\n\n")
	b.WriteString("Responsibilities:\n")
	b.WriteString(jd.ResponsibilitiesJSON)
	b.WriteString("\n\n")
	b.WriteString("Keywords:\n")
	b.WriteString(jd.KeywordsJSON)
	b.WriteString("\n")

	return b.String()
}

// formatChunk renders a chunk with a type-appropriate label for prompt display.
func formatChunk(c rag.Chunk) string {
	label := string(c.Type)
	switch c.Type {
	case rag.ChunkExperience:
		if title, ok := c.Metadata["title"]; ok && title != "" {
			label = "Experience"
		}
		if company, ok := c.Metadata["company"]; ok && company != "" {
			label += fmt.Sprintf(" — %s", company)
		}
		if duration, ok := c.Metadata["duration"]; ok && duration != "" {
			label += " " + duration
		}
	case rag.ChunkSkill:
		if cat, ok := c.Metadata["category"]; ok {
			label = fmt.Sprintf("Skills — %s", cat)
		}
	case rag.ChunkEducation:
		if degree, ok := c.Metadata["degree"]; ok && degree != "" {
			label = "Education"
		}
		if school, ok := c.Metadata["school"]; ok && school != "" {
			label += fmt.Sprintf(" — %s", school)
		}
	}
	return fmt.Sprintf("[%s]\n%s", label, c.Content)
}
