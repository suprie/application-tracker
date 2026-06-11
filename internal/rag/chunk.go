package rag

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ChunkType classifies a profile chunk for prompt formatting.
type ChunkType string

const (
	ChunkSummary    ChunkType = "summary"
	ChunkExperience ChunkType = "experience"
	ChunkSkill      ChunkType = "skill"
	ChunkEducation  ChunkType = "education"
)

// Chunk is a retrievable unit of the master profile.
type Chunk struct {
	ID       string            `json:"id"`
	Type     ChunkType         `json:"type"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
}

// profileYAML mirrors the master profile YAML schema produced by cvparser.
type profileYAML struct {
	Name                string   `yaml:"name"`
	Headline            string   `yaml:"headline"`
	Location            string   `yaml:"location"`
	Email               string   `yaml:"email"`
	Phone               string   `yaml:"phone"`
	Linkedin            string   `yaml:"linkedin"`
	Summary             string   `yaml:"summary"`
	TotalYearsExp       any      `yaml:"total_years_experience"`
	Domains             []string `yaml:"domains"`
	Skills              skillsYAML
	Experience          []experienceEntry `yaml:"experience"`
	Education           []educationEntry  `yaml:"education"`
	ParsingWarnings     []string          `yaml:"parsing_warnings"`
}

type skillsYAML struct {
	Languages    []string `yaml:"languages"`
	Mobile       []string `yaml:"mobile"`
	Architecture []string `yaml:"architecture"`
	Testing      []string `yaml:"testing"`
	CICD         []string `yaml:"ci_cd"`
	Tools        []string `yaml:"tools"`
	Leadership   []string `yaml:"leadership"`
}

type experienceEntry struct {
	Title      string   `yaml:"title"`
	Company    string   `yaml:"company"`
	StartDate  string   `yaml:"start_date"`
	EndDate    string   `yaml:"end_date"`
	TeamSize   any      `yaml:"team_size"`
	Highlights []string `yaml:"highlights"`
}

type educationEntry struct {
	School    string `yaml:"school"`
	Degree    string `yaml:"degree"`
	StartDate string `yaml:"start_date"`
	EndDate   string `yaml:"end_date"`
}

// ChunkProfile parses a master profile YAML document and splits it into
// retrievable chunks. Parsing warnings are skipped — they carry no signal
// for matching or retrieval.
func ChunkProfile(yamlBytes []byte) ([]Chunk, error) {
	var p profileYAML
	if err := yaml.Unmarshal(yamlBytes, &p); err != nil {
		return nil, fmt.Errorf("unmarshal profile yaml: %w", err)
	}

	var chunks []Chunk

	// Summary chunk — roll up top-level bio fields.
	summaryParts := collectNonEmpty(
		p.Summary,
		formatOptional("Headline", p.Headline),
		formatOptional("Location", p.Location),
		formatOptional("Email", p.Email),
		formatOptional("LinkedIn", p.Linkedin),
		formatOptional("Total years of experience", formatAny(p.TotalYearsExp)),
	)

	if len(p.Domains) > 0 {
		summaryParts = append(summaryParts,
			fmt.Sprintf("Domains: %s", strings.Join(p.Domains, ", ")))
	}

	if len(summaryParts) > 0 {
		chunks = append(chunks, Chunk{
			ID:      "summary",
			Type:    ChunkSummary,
			Content: strings.Join(summaryParts, ". "),
			Metadata: map[string]string{
				"name": p.Name,
			},
		})
	}

	// Experience chunks — one per role.
	for i, exp := range p.Experience {
		content := formatExperience(exp)
		if content == "" {
			continue
		}
		chunks = append(chunks, Chunk{
			ID:      fmt.Sprintf("exp-%d", i),
			Type:    ChunkExperience,
			Content: content,
			Metadata: map[string]string{
				"title":    exp.Title,
				"company":  exp.Company,
				"duration": formatDateRange(exp.StartDate, exp.EndDate),
			},
		})
	}

	// Skill chunks — one per category with at least one item.
	skillCategories := map[string][]string{
		"Languages":    p.Skills.Languages,
		"Mobile":       p.Skills.Mobile,
		"Architecture": p.Skills.Architecture,
		"Testing":      p.Skills.Testing,
		"CI/CD":        p.Skills.CICD,
		"Tools":        p.Skills.Tools,
		"Leadership":   p.Skills.Leadership,
	}
	for category, items := range skillCategories {
		if len(items) == 0 {
			continue
		}
		chunks = append(chunks, Chunk{
			ID:      fmt.Sprintf("skill-%s", strings.ToLower(category)),
			Type:    ChunkSkill,
			Content: fmt.Sprintf("Skills — %s: %s", category, strings.Join(items, ", ")),
			Metadata: map[string]string{
				"category": category,
			},
		})
	}

	// Education chunks — one per entry.
	for i, edu := range p.Education {
		content := formatEducation(edu)
		if content == "" {
			continue
		}
		chunks = append(chunks, Chunk{
			ID:      fmt.Sprintf("edu-%d", i),
			Type:    ChunkEducation,
			Content: content,
			Metadata: map[string]string{
				"school":   edu.School,
				"degree":   edu.Degree,
				"duration": formatDateRange(edu.StartDate, edu.EndDate),
			},
		})
	}

	return chunks, nil
}

// ---- helpers ----

func formatExperience(exp experienceEntry) string {
	if exp.Title == "" && exp.Company == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Role: %s", exp.Title))
	if exp.Company != "" {
		b.WriteString(fmt.Sprintf(" at %s", exp.Company))
	}
	b.WriteString(formatDateRange(exp.StartDate, exp.EndDate))
	if v := formatAny(exp.TeamSize); v != "" {
		b.WriteString(fmt.Sprintf(". Team size: %s", v))
	}
	if len(exp.Highlights) > 0 {
		b.WriteString(". Highlights: ")
		for i, h := range exp.Highlights {
			if i > 0 {
				b.WriteString("; ")
			}
			b.WriteString(h)
		}
	}
	return b.String()
}

func formatEducation(edu educationEntry) string {
	if edu.Degree == "" && edu.School == "" {
		return ""
	}
	parts := collectNonEmpty(
		edu.Degree,
		optionalPrefix("from", edu.School),
	)
	if len(parts) == 0 {
		return ""
	}
	result := strings.Join(parts, " ")
	result += formatDateRange(edu.StartDate, edu.EndDate)
	return result
}

func formatDateRange(start, end string) string {
	if start == "" && end == "" {
		return ""
	}
	s := strings.TrimSpace(start)
	e := strings.TrimSpace(end)
	// "Present" means current role.
	if e == "" || strings.EqualFold(e, "present") {
		e = "Present"
	}
	if s == "" {
		return fmt.Sprintf(" (until %s)", e)
	}
	return fmt.Sprintf(" (%s–%s)", s, e)
}

func formatOptional(label, value string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf("%s: %s", label, value)
}

func formatAny(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case int:
		if val == 0 {
			return ""
		}
		return fmt.Sprintf("%d", val)
	case float64:
		if val == 0 {
			return ""
		}
		return fmt.Sprintf("%.0f", val)
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

func optionalPrefix(prefix, value string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf("%s %s", prefix, value)
}

func collectNonEmpty(values ...string) []string {
	var out []string
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			out = append(out, v)
		}
	}
	return out
}
