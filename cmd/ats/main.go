package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"suprie/application_tracker/internal/cvextractor"
	"suprie/application_tracker/internal/cvparser"
	"suprie/application_tracker/internal/jdextractor"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/textutils"
	"suprie/application_tracker/internal/fileutil"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage: ats <parse-jd|parse-cv> <file.pdf>")
		os.Exit(1)
	}

	command := os.Args[1]
	filename := os.Args[2]

	switch command {
	case "parse-cv":
		runParseCV(filename)
	case "parse-jd":
		runParseJD(filename)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		fmt.Fprintf(os.Stderr, "Usage: ats <parse-cv|parse-jd> <file>\n")
		os.Exit(1)
	}

}

func runParseJD(filepath string) {

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

	log.Printf("Response: %s", response)
}

func runParseCV(filename string) {
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
