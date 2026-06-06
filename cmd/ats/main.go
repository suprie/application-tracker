package main

import (
	"context"
	"fmt"
	"os"

	"suprie/application_tracker/internal/cvextractor"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/cvparser"
	"suprie/application_tracker/internal/textutils"
	"suprie/application_tracker/internal/utilities"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: ats <file.pdf>")
		os.Exit(1)
	}

	extractor := cvextractor.RustPDFExtractor{
		BinaryPath: "./bin/ats-reader",
	}

	filename := os.Args[1]
	fmt.Printf("Processing: %s\n", filename)

	result, err := extractor.Extract(context.Background(), filename)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	normalizeText := textutils.NormalizeText(result.Text)

	prompt := cvparser.BuildCVParserPrompt(normalizeText)

	lmClient := llm.LMStudioClient{
		BaseURL: "http://localhost:1234",
		Model:   "gemma-3n-e4b",
	}

	response, err := lmClient.Generate(context.Background(), prompt)
	if err != nil {
		panic(err)
	}

	utilities.SaveYAML(
		"generated/master_profile.yaml",
		utilities.StripMarkdownFence(response),
	)
}
