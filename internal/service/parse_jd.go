package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/dto"
	"suprie/application_tracker/internal/jdextractor"
	"suprie/application_tracker/internal/repository"
	"suprie/application_tracker/internal/textutils"
)

// RunParseJD is the CLI wrapper for ParseJDText.
func RunParseJD(filepath string, applyURL string, repo repository.JobDescriptionRepository) {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("read file %s : %v", filepath, err)
	}

	d := NewDeps(repo, nil)
	jd, err := ParseJDText(context.Background(), d, string(bytes), applyURL)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("Response: %+v\n", *jd)
}

// ParseJDText extracts a job description from raw text and, when a repository
// is configured, persists it. It is the server-callable core: it takes the JD
// text directly (so the web layer can pass pasted text) and returns errors
// instead of exiting the process.
func ParseJDText(ctx context.Context, d Deps, text, applyURL string) (*domain.JobDescription, error) {
	prompt := jdextractor.BuildJDParserPrompt(textutils.NormalizeText(text))

	client := d.LLM("parse_jd")
	schema := jdextractor.BuildJSONSchema()
	response, err := client.Generate(ctx, prompt, &schema)
	if err != nil {
		return nil, fmt.Errorf("extracting JD: %w", err)
	}

	var jsonResponse dto.JobDescriptionResponse
	if err = json.Unmarshal([]byte(response), &jsonResponse); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	jd, err := mapToJobDescription(&jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("map job description: %w", err)
	}

	if applyURL != "" {
		jd.ApplyURL = &applyURL
	}

	if d.JDRepo != nil {
		if err := d.JDRepo.Create(ctx, &jd); err != nil {
			return nil, fmt.Errorf("saving job description: %w", err)
		}
		d.Logger.Printf("Saved job description (id=%d)", jd.ID)
	}

	return &jd, nil
}

func mapToJobDescription(response *dto.JobDescriptionResponse) (domain.JobDescription, error) {
	requirementJSON, err := json.Marshal(response.Requirements)
	if err != nil {
		return domain.JobDescription{}, fmt.Errorf("marshaling requirements: %w", err)
	}

	responsibilitiesJSON, err := json.Marshal(response.Responsibilities)
	if err != nil {
		return domain.JobDescription{}, fmt.Errorf("marshaling responsibilities: %w", err)
	}

	keywordJSON, err := json.Marshal(response.Keywords)
	if err != nil {
		return domain.JobDescription{}, fmt.Errorf("marshaling keywords: %w", err)
	}

	parsingWarningsJSON, err := json.Marshal(response.ParsingWarnings)
	if err != nil {
		return domain.JobDescription{}, fmt.Errorf("marshaling parsing warnings: %w", err)
	}

	return domain.JobDescription{
		Company:              response.Company,
		RoleTitle:            response.RoleTitle,
		Seniority:            response.Seniority,
		EmploymentType:       response.EmploymentType,
		WorkArrangement:      response.WorkArrangement,
		Location:             response.Location,
		RequirementsJSON:     string(requirementJSON),
		ResponsibilitiesJSON: string(responsibilitiesJSON),
		KeywordsJSON:         string(keywordJSON),
		ParsingWarningJSON:   string(parsingWarningsJSON),
		CreatedAt:            time.Now(),
	}, nil
}
