package service

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/dto"
	"suprie/application_tracker/internal/jdextractor"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/textutils"
	"time"
)

func RunParseJD(filepath string) {

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("read file %s : %v", filepath, err)
	}

	normalizeText := textutils.NormalizeText(string(bytes))

	prompt := jdextractor.BuildJDParserPrompt(
		normalizeText,
	)

	lmClient := llm.NewLMStudioClient()
	schema := jdextractor.BuildJSONSchema()

	response, err := lmClient.Generate(context.Background(), prompt, &schema)
	if err != nil {
		log.Fatalf("extracting JD %v", err)
	}

	var jsonResponse dto.JobDescriptionResponse

	if err := json.Unmarshal([]byte(response), &jsonResponse); err != nil {
		log.Fatalf("unmarshal response: %v", err)
	}

	dto := mapToJobDescription(&jsonResponse)

	log.Printf("Response: %+v\n", dto)
}

func mapToJobDescription(response *dto.JobDescriptionResponse) domain.JobDescription {
	requirementJSON, _ := json.Marshal(response.Requirements)
	keywordJson, _ := json.Marshal(response.Keywords)
	parsingWarnings, _ := json.Marshal(response.ParsingWarnings)

	return domain.JobDescription{
		Company:            response.Company,
		RoleTitle:          response.RoleTitle,
		Seniority:          response.Seniority,
		EmploymentType:     response.EmploymentType,
		WorkArrangement:    response.WorkArrangement,
		RequirementsJson:   string(requirementJSON),
		KeywordsJson:       string(keywordJson),
		ParsingWarningJson: string(parsingWarnings),
		CreatedAt:          time.Now(),
	}
}
