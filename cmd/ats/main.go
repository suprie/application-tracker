package main

import (
	"context"
	"fmt"
	"os"

	"suprie/application_tracker/internal/cvextractor"
	"suprie/application_tracker/internal/cvparser"
	"suprie/application_tracker/internal/jdextractor"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/textutils"
	"suprie/application_tracker/internal/utilities"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage: ats <file.pdf>")
		os.Exit(1)
	}

	command := os.Args[1]
	filename := os.Args[2]

	switch command {
	case "parse-cv":
		runParseCV(filename)
	case "parse-jd":
		runParseJD(filename)
	}

}

func runParseJD(filepath string) {

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	normalizeText := textutils.NormalizeText(string(bytes))

	prompt := jdextractor.BuildJDParserPrompt(
		normalizeText,
	)

	println(prompt)
	lmClient := llm.LMStudioClient{
		BaseURL: "http://localhost:1234",
		Model:   "gemma-4-12b-qat",
	}

	schema := jdextractor.BuildJSONSchema()

	response, err := lmClient.Generate(context.Background(), prompt, &schema)
	if err != nil {
		panic(err)
	}

	println("Response : %s", response)
}

func runParseCV(filename string) {
	extractor := cvextractor.RustPDFExtractor{
		BinaryPath: "./bin/ats-reader",
	}

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

	response, err := lmClient.Generate(context.Background(), prompt, nil)
	if err != nil {
		panic(err)
	}

	utilities.SaveYAML(
		"generated/master_profile.yaml",
		utilities.StripMarkdownFence(response),
	)

}
