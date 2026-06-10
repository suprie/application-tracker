package coverletter

import (
	"fmt"
	"strings"

	"suprie/application_tracker/internal/domain"
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
	b.WriteString("- \"body_paragraphs\": 2-3 paragraphs — the complete letter body\n")
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
