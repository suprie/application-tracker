package coverletter

import (
	"fmt"
	"strings"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/rag"
	"suprie/application_tracker/internal/ranker"
)

// BuildCoverLetterPrompt creates the LLM prompt for generating a tailored
// cover letter from the master profile and job description.
func BuildCoverLetterPrompt(masterProfile string, jd *domain.JobDescription) string {
	var b strings.Builder

	b.WriteString("You are a professional cover letter writer.\n")
	b.WriteString("Write a tailored, compelling cover letter for the job described below.\n")
	b.WriteString("Use the candidate's master profile to highlight relevant experience.\n\n")
	b.WriteString(`
Before writing:
1. Identify the top 3 matching experiences.
2. Identify the top 3 matching skills.
3. Identify any gaps.
4. Focus only on the strongest matches.
Writing rules:
- Maximum 350 words.
- Sound like an experienced engineer.
- Avoid recruiter buzzwords.
- Avoid repeating the resume.
- Use concrete achievements.
- Mention company-specific context when relevant.
- Prioritize impact over responsibilities.
- Do not invent experience.
- Do not mention gaps unless they are critical.
	`)

	b.WriteString("IMPORTANT — structure your response as follows:\n")
	b.WriteString("- \"opening\": just the salutation (e.g. \"Dear Hiring Manager,\")\n")
	b.WriteString("- \"closing\": just the closing phrase (e.g. \"Sincerely,\")\n")
	b.WriteString("- \"opening_paragraphs\": 1 paragraphs — A brief introduction who am i \n")
	b.WriteString("- \"body_paragraphs\": 2-3 paragraphs — the complete letter body\n")
	b.WriteString("- \"closing_paragraphs\": 1 paragraphs — some kind of looking forward to hear from you\n")
	b.WriteString("- \"your_name\": your full name (used in the signature)\n")
	b.WriteString("Respond as a JSON object with the fields specified in the schema.\n\n")

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
	b.WriteString("\nRequirements:\n")
	b.WriteString(jd.RequirementsJSON)
	b.WriteString("\n\nResponsibilities:\n")
	b.WriteString(jd.ResponsibilitiesJSON)
	b.WriteString("\n")

	return b.String()
}

// BuildCoverLetterPromptWithChunks creates the LLM prompt for cover letter
// generation using RAG-retrieved profile chunks instead of the full master
// profile. Only the chunks most relevant to the JD are included.
func BuildCoverLetterPromptWithChunks(chunks []rag.Chunk, jd *domain.JobDescription) string {
	var b strings.Builder

	b.WriteString("You are a professional cover letter writer.\n")
	b.WriteString("Write a tailored, compelling cover letter for the job described below.\n")
	b.WriteString("Use the candidate's relevant experience (retrieved below) to highlight matches.\n\n")
	b.WriteString(`
Before writing:
1. Identify the top 3 matching experiences from the retrieved chunks.
2. Identify the top 3 matching skills from the retrieved chunks.
3. Focus only on the strongest matches present in the retrieved chunks.
Writing rules:
- Maximum 350 words.
- Sound like an experienced engineer.
- Avoid recruiter buzzwords.
- Avoid repeating the resume.
- Use concrete achievements from the retrieved chunks.
- Mention company-specific context when relevant.
- Prioritize impact over responsibilities.
- Do not invent experience.
- Do not mention gaps.
	`)

	b.WriteString("IMPORTANT — structure your response as follows:\n")
	b.WriteString("- \"opening\": just the salutation (e.g. \"Dear Hiring Manager,\")\n")
	b.WriteString("- \"closing\": just the closing phrase (e.g. \"Sincerely,\")\n")
	b.WriteString("- \"opening_paragraphs\": 1 paragraph — A brief introduction of who you are\n")
	b.WriteString("- \"body_paragraphs\": 2-3 paragraphs — the complete letter body\n")
	b.WriteString("- \"closing_paragraphs\": 1 paragraph — looking forward to hearing from them\n")
	b.WriteString("- \"your_name\": your full name (used in the signature)\n")
	b.WriteString("Respond as a JSON object with the fields specified in the schema.\n\n")

	b.WriteString("=== RETRIEVED RELEVANT EXPERIENCE ===\n")
	for _, c := range chunks {
		label := formatCoverLetterChunkLabel(c)
		b.WriteString(fmt.Sprintf("[%s]\n%s\n\n", label, c.Content))
	}

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
	b.WriteString("\nRequirements:\n")
	b.WriteString(jd.RequirementsJSON)
	b.WriteString("\n\nResponsibilities:\n")
	b.WriteString(jd.ResponsibilitiesJSON)
	b.WriteString("\n")

	return b.String()
}

// BuildCoverLetterPromptWithRanked creates the cover letter prompt using the
// ranker's curated selections instead of raw BM25 chunks. The ranker provides
// scored experiences, selected skills, and a recommended narrative.
func BuildCoverLetterPromptWithRanked(ranked *ranker.Response, jd *domain.JobDescription) string {
	var b strings.Builder

	b.WriteString("You are a professional cover letter writer.\n")
	b.WriteString("Write a tailored, compelling cover letter for the job described below.\n")
	b.WriteString("Use the curated experiences and skills provided — they have been selected\n")
	b.WriteString("specifically for this role based on strategic fit.\n\n")
	b.WriteString(`
Writing rules:
- Maximum 350 words.
- Sound like an experienced engineer.
- Avoid recruiter buzzwords.
- Use the concrete achievements listed in the selected experiences.
- Mention company-specific context when relevant.
- Prioritize impact over responsibilities.
- Do not invent experience.
- Do not mention gaps.
	`)

	b.WriteString("IMPORTANT — structure your response as follows:\n")
	b.WriteString("- \"opening\": just the salutation (e.g. \"Dear Hiring Manager,\")\n")
	b.WriteString("- \"closing\": just the closing phrase (e.g. \"Sincerely,\")\n")
	b.WriteString("- \"opening_paragraphs\": 1 paragraph — A brief introduction of who you are\n")
	b.WriteString("- \"body_paragraphs\": 2-3 paragraphs — the complete letter body\n")
	b.WriteString("- \"closing_paragraphs\": 1 paragraph — looking forward to hearing from them\n")
	b.WriteString("- \"your_name\": your full name (used in the signature)\n")
	b.WriteString("Respond as a JSON object with the fields specified in the schema.\n\n")

	// Ranker's recommended narrative.
	if ranked.RecommendedNarrative != "" {
		b.WriteString("=== RECOMMENDED NARRATIVE ===\n")
		b.WriteString(ranked.RecommendedNarrative)
		b.WriteString("\n\n")
	}

	// Selected experiences.
	b.WriteString("=== SELECTED EXPERIENCES (ranked by strategic fit) ===\n")
	for i, exp := range ranked.SelectedExperiences {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, exp.Title))
		b.WriteString(fmt.Sprintf("   Summary: %s\n", exp.Summary))
		b.WriteString(fmt.Sprintf("   Why selected: %s\n", exp.WhySelected))
		if len(exp.Evidence) > 0 {
			b.WriteString("   Evidence:\n")
			for _, e := range exp.Evidence {
				b.WriteString(fmt.Sprintf("     - %s\n", e))
			}
		}
		b.WriteString("\n")
	}

	// Selected skills.
	if len(ranked.SelectedSkills) > 0 {
		b.WriteString("=== SELECTED SKILLS ===\n")
		for _, s := range ranked.SelectedSkills {
			b.WriteString(fmt.Sprintf("- %s", s.Skill))
			if s.Evidence != "" {
				b.WriteString(fmt.Sprintf(" (%s)", s.Evidence))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Warnings — things to be careful about.
	if len(ranked.Warnings) > 0 {
		b.WriteString("=== WARNINGS ===\n")
		for _, w := range ranked.Warnings {
			b.WriteString(fmt.Sprintf("- %s\n", w))
		}
		b.WriteString("\n")
	}

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
	b.WriteString("\nRequirements:\n")
	b.WriteString(jd.RequirementsJSON)
	b.WriteString("\n\nResponsibilities:\n")
	b.WriteString(jd.ResponsibilitiesJSON)
	b.WriteString("\n")

	return b.String()
}

// formatCoverLetterChunkLabel renders a chunk label for cover letter prompts.
func formatCoverLetterChunkLabel(c rag.Chunk) string {
	switch c.Type {
	case rag.ChunkExperience:
		company := c.Metadata["company"]
		title := c.Metadata["title"]
		duration := c.Metadata["duration"]
		if company != "" && title != "" {
			return fmt.Sprintf("Experience: %s at %s %s", title, company, duration)
		}
		return "Experience"
	case rag.ChunkSkill:
		cat := c.Metadata["category"]
		if cat != "" {
			return fmt.Sprintf("Skills — %s", cat)
		}
		return "Skills"
	case rag.ChunkEducation:
		degree := c.Metadata["degree"]
		school := c.Metadata["school"]
		if degree != "" && school != "" {
			return fmt.Sprintf("Education: %s, %s", degree, school)
		}
		return "Education"
	default:
		return string(c.Type)
	}
}
