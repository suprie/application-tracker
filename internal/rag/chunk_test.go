package rag

import (
	"strings"
	"testing"
)

// fullProfileYAML is a realistic master profile for testing.
const fullProfileYAML = `
name: Alice Engineer
headline: Senior Backend Engineer
location: Berlin, Germany
email: alice@example.com
phone: "+49123456789"
linkedin: linkedin.com/in/alice-engineer
summary: Experienced backend engineer with a focus on distributed systems and Go.
total_years_experience: 9
domains:
  - Fintech
  - Cloud Infrastructure
skills:
  languages:
    - Go
    - Python
    - TypeScript
  mobile: []
  architecture:
    - Kubernetes
    - Docker
    - Terraform
    - gRPC
  testing:
    - Table-driven tests
  ci_cd:
    - GitHub Actions
    - ArgoCD
  tools:
    - Prometheus
    - Grafana
  leadership:
    - Mentoring
    - Code review
experience:
  - title: Senior Backend Engineer
    company: Acme Corp
    start_date: "2020"
    end_date: "2024"
    team_size: 5
    highlights:
      - Built Kubernetes operators for automated DB provisioning
      - Migrated monolith to event-driven microservices
  - title: Backend Engineer
    company: StartupCo
    start_date: "2018"
    end_date: "2020"
    team_size: 8
    highlights:
      - Designed real-time fraud detection pipeline
      - Reduced p99 latency by 40%
  - title: ""
    company: ""
    start_date: ""
    end_date: ""
education:
  - school: TU Berlin
    degree: M.Sc. Computer Science
    start_date: "2016"
    end_date: "2018"
  - school: University of Munich
    degree: B.Sc. Computer Science
    start_date: "2013"
    end_date: "2016"
parsing_warnings:
  - Missing phone number format
`

func TestChunkProfile_ParsesFullYAML(t *testing.T) {
	chunks, err := ChunkProfile([]byte(fullProfileYAML))
	if err != nil {
		t.Fatalf("ChunkProfile returned error: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected non-empty chunks")
	}
}

func TestChunkProfile_SummaryChunk(t *testing.T) {
	chunks, err := ChunkProfile([]byte(fullProfileYAML))
	if err != nil {
		t.Fatalf("ChunkProfile: %v", err)
	}

	var summary *Chunk
	for i := range chunks {
		if chunks[i].Type == ChunkSummary {
			summary = &chunks[i]
			break
		}
	}
	if summary == nil {
		t.Fatal("expected a summary chunk")
	}
	if summary.ID != "summary" {
		t.Errorf("expected summary ID 'summary', got %q", summary.ID)
	}
	if !strings.Contains(summary.Content, "Experienced backend engineer") {
		t.Errorf("summary content missing bio text: %s", summary.Content)
	}
	if !strings.Contains(summary.Content, "Fintech") {
		t.Errorf("summary content missing domains: %s", summary.Content)
	}
	if v, ok := summary.Metadata["name"]; !ok || v != "Alice Engineer" {
		t.Errorf("summary metadata name = %q, want 'Alice Engineer'", v)
	}
}

func TestChunkProfile_ExperienceChunks(t *testing.T) {
	chunks, err := ChunkProfile([]byte(fullProfileYAML))
	if err != nil {
		t.Fatalf("ChunkProfile: %v", err)
	}

	count := 0
	for _, c := range chunks {
		if c.Type == ChunkExperience {
			count++
			if c.Content == "" {
				t.Errorf("experience chunk %s has empty content", c.ID)
			}
			if c.Metadata["title"] == "" {
				t.Errorf("experience chunk %s missing title metadata", c.ID)
			}
			if c.Metadata["company"] == "" {
				t.Errorf("experience chunk %s missing company metadata", c.ID)
			}
		}
	}
	if count != 2 {
		t.Errorf("expected 2 experience chunks, got %d", count)
	}
}

func TestChunkProfile_SkillChunks(t *testing.T) {
	chunks, err := ChunkProfile([]byte(fullProfileYAML))
	if err != nil {
		t.Fatalf("ChunkProfile: %v", err)
	}

	count := 0
	for _, c := range chunks {
		if c.Type == ChunkSkill {
			count++
			if c.Content == "" {
				t.Errorf("skill chunk %s has empty content", c.ID)
			}
			if c.Metadata["category"] == "" {
				t.Errorf("skill chunk %s missing category metadata", c.ID)
			}
		}
	}
	// languages, architecture, testing, ci_cd, tools, leadership = 6 (mobile is empty)
	if count != 6 {
		t.Errorf("expected 6 skill chunks, got %d", count)
	}
}

func TestChunkProfile_EducationChunks(t *testing.T) {
	chunks, err := ChunkProfile([]byte(fullProfileYAML))
	if err != nil {
		t.Fatalf("ChunkProfile: %v", err)
	}

	count := 0
	for _, c := range chunks {
		if c.Type == ChunkEducation {
			count++
			if c.Content == "" {
				t.Errorf("education chunk %s has empty content", c.ID)
			}
		}
	}
	if count != 2 {
		t.Errorf("expected 2 education chunks, got %d", count)
	}
}

func TestChunkProfile_SkipsEmptyExperience(t *testing.T) {
	chunks, err := ChunkProfile([]byte(fullProfileYAML))
	if err != nil {
		t.Fatalf("ChunkProfile: %v", err)
	}
	// The third experience entry has empty title and company.
	for _, c := range chunks {
		if c.Type == ChunkExperience && c.ID == "exp-2" {
			t.Errorf("empty experience entry should have been skipped")
		}
	}
}

func TestChunkProfile_NoParsingWarnings(t *testing.T) {
	chunks, err := ChunkProfile([]byte(fullProfileYAML))
	if err != nil {
		t.Fatalf("ChunkProfile: %v", err)
	}
	for _, c := range chunks {
		if strings.Contains(c.Content, "parsing_warning") ||
			strings.Contains(c.Content, "Missing phone") {
			t.Errorf("chunk %s contains parsing warning content: %s", c.ID, c.Content)
		}
	}
}

func TestChunkProfile_InvalidYAML(t *testing.T) {
	_, err := ChunkProfile([]byte(`: bad yaml : :`))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestChunkProfile_EmptyYAML(t *testing.T) {
	chunks, err := ChunkProfile([]byte(``))
	if err != nil {
		t.Fatalf("ChunkProfile: %v", err)
	}
	// Empty YAML produces no chunks — that's fine.
	if len(chunks) != 0 {
		t.Logf("got %d chunks from empty YAML (acceptable)", len(chunks))
	}
}
