package service

import (
	"context"
	"fmt"
	"log"
	"suprie/application_tracker/internal/cvextractor"
	"suprie/application_tracker/internal/cvparser"
	"suprie/application_tracker/internal/fileutil"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/textutils"
)

func RunParseCV(filename string) {
	extractor := cvextractor.RustPDFExtractor{
		BinaryPath: "./bin/ats-reader",
	}

	fmt.Printf("Processing: %s\n", filename)

	result, err := extractor.Extract(context.Background(), filename)

	if err != nil {
		log.Fatalf("extracting CV %v", err)
	}

	normalizeText := textutils.NormalizeText(result.Text)

	prompt := cvparser.BuildCVParserPrompt(normalizeText)

	lmClient := llm.NewLMStudioClient()

	response, err := lmClient.Generate(context.Background(), prompt, nil)
	if err != nil {
		log.Fatalf("generating response %v", err)
	}

	if err = fileutil.SaveYAML(
		"generated/master_profile.yaml",
		fileutil.StripMarkdownFence(response),
	); err != nil {
		log.Fatalf("saving yaml file %v", err)
	}

}
