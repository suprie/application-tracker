package ranker

import (
	"fmt"
	"strings"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/rag"
)

// BuildPrompt creates the experience ranker prompt from the profile summary,
// BM25-retrieved chunks, and job description.
func BuildPrompt(profileSummary string, chunks []rag.Chunk, jd *domain.JobDescription) string {
	var b strings.Builder

	b.WriteString("You are an Experience Ranker Agent.\n\n")
	b.WriteString("Your task is to select the strongest candidate experiences for a specific job.\n")
	b.WriteString("You are NOT writing the cover letter.\n\n")

	b.WriteString("Input:\n")
	b.WriteString("1. Job description\n")
	b.WriteString("2. Retrieved experience chunks from BM25\n")
	b.WriteString("3. Candidate profile summary, if available\n\n")

	b.WriteString("Goal:\n")
	b.WriteString("Select the best experiences that should be used by a cover letter writer.\n\n")

	b.WriteString("Important:\n")
	b.WriteString("BM25 retrieval is only a recall mechanism. Do not blindly trust keyword similarity.\n")
	b.WriteString("Your job is to re-rank the retrieved chunks based on strategic fit.\n\n")

	b.WriteString("Selection rules:\n")
	b.WriteString("- Select exactly 3 experiences.\n")
	b.WriteString("- Select exactly 3 to 5 skills.\n")
	b.WriteString("- Prefer experiences with concrete outcomes.\n")
	b.WriteString("- Prefer rare, senior-level, career-defining achievements.\n")
	b.WriteString("- Prefer business impact over pure technical similarity.\n")
	b.WriteString("- Prefer product and organizational impact when relevant.\n")
	b.WriteString("- Use technical achievements when they strongly support the role requirements.\n")
	b.WriteString("- Do not invent facts.\n")
	b.WriteString("- Do not use experiences not present in the input.\n")
	b.WriteString("- Do not select multiple experiences that tell the same story unless necessary.\n")
	b.WriteString("- Avoid selecting an experience only because it shares keywords with the job description.\n\n")

	b.WriteString("Scoring rubric:\n")
	b.WriteString("For each experience, score from 0 to 5:\n\n")
	b.WriteString("1. relevance_to_job\n")
	b.WriteString("How directly this experience matches the job responsibilities.\n\n")
	b.WriteString("2. business_or_product_impact\n")
	b.WriteString("Revenue, user impact, product adoption, engagement, reliability, or customer value.\n\n")
	b.WriteString("3. technical_depth\n")
	b.WriteString("Architecture, performance, scalability, reliability, migration, API design, or production complexity.\n\n")
	b.WriteString("4. seniority_signal\n")
	b.WriteString("Leadership, ownership, cross-functional collaboration, mentoring, decision making, or scope.\n\n")
	b.WriteString("5. rarity\n")
	b.WriteString("How uncommon or differentiating this experience is compared to typical candidates.\n\n")
	b.WriteString("6. narrative_fit\n")
	b.WriteString("How well this experience helps tell a coherent career story for this specific job.\n\n")

	b.WriteString("Final score:\n")
	b.WriteString("final_score =\n")
	b.WriteString("  relevance_to_job * 0.25 +\n")
	b.WriteString("  business_or_product_impact * 0.25 +\n")
	b.WriteString("  technical_depth * 0.20 +\n")
	b.WriteString("  seniority_signal * 0.15 +\n")
	b.WriteString("  rarity * 0.10 +\n")
	b.WriteString("  narrative_fit * 0.05\n\n")

	b.WriteString("Return JSON only.\n\n")

	// --- Input data ---

	if profileSummary != "" {
		b.WriteString("=== CANDIDATE PROFILE SUMMARY ===\n")
		b.WriteString(profileSummary)
		b.WriteString("\n\n")
	}

	b.WriteString("=== RETRIEVED EXPERIENCE CHUNKS (BM25) ===\n")
	for _, c := range chunks {
		b.WriteString(fmt.Sprintf("[%s] %s\n\n", c.ID, c.Content))
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

// ProfileSummary extracts a short summary string from chunks. Returns the
// summary chunk content if one exists, otherwise empty string.
func ProfileSummary(chunks []rag.Chunk) string {
	for _, c := range chunks {
		if c.Type == rag.ChunkSummary {
			return c.Content
		}
	}
	return ""
}
