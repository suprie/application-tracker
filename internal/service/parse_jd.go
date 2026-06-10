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
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/repository"
	"suprie/application_tracker/internal/textutils"
)

// RunParseJD extracts a job description from a PDF and optionally persists it.
// Pass nil for repo to skip persistence (print only).
// applyURL is the URL of the job posting (optional).
func RunParseJD(filepath string, applyURL string, repo repository.JobDescriptionRepository) {

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("read file %s : %v", filepath, err)
	}

	normalizeText := textutils.NormalizeText(string(bytes))

	prompt := jdextractor.BuildJDParserPrompt(
		normalizeText,
	)

	lmClient := llm.NewLMStudioClient("parse_jd")
	schema := jdextractor.BuildJSONSchema()

	response, err := lmClient.Generate(context.Background(), prompt, &schema)
	if err != nil {
		log.Fatalf("extracting JD %v", err)
	}

	var jsonResponse dto.JobDescriptionResponse

	if err = json.Unmarshal([]byte(response), &jsonResponse); err != nil {
		log.Fatalf("unmarshal response: %v", err)
	}

	jd, err := mapToJobDescription(&jsonResponse)
	if err != nil {
		log.Fatalf("map job description: %v", err)
	}

	if applyURL != "" {
		jd.ApplyURL = &applyURL
	}

	if repo != nil {
		if err := repo.Create(context.Background(), &jd); err != nil {
			log.Fatalf("saving job description: %v", err)
		}
		log.Printf("Saved job description (id=%d)\n", jd.ID)
	}

	log.Printf("Response: %+v\n", jd)
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
		Company:                response.Company,
		RoleTitle:              response.RoleTitle,
		Seniority:              response.Seniority,
		EmploymentType:         response.EmploymentType,
		WorkArrangement:        response.WorkArrangement,
		Location:               response.Location,
		RequirementsJSON:       string(requirementJSON),
		ResponsibilitiesJSON:   string(responsibilitiesJSON),
		KeywordsJSON:           string(keywordJSON),
		ParsingWarningJSON:     string(parsingWarningsJSON),
		CreatedAt:              time.Now(),
	}, nil
}
