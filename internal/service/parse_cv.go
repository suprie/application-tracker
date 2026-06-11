package service

import (
	"context"
	"fmt"
	"log"
	"suprie/application_tracker/internal/cvextractor"
	"suprie/application_tracker/internal/cvparser"
	"suprie/application_tracker/internal/fileutil"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/rag"
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

	lmClient := llm.NewLMStudioClient("parse_cv")

	response, err := lmClient.Generate(context.Background(), prompt, nil)
	if err != nil {
		log.Fatalf("generating response %v", err)
	}

	yamlContent := fileutil.StripMarkdownFence(response)
	if err = fileutil.SaveYAML(
		"generated/master_profile.yaml",
		yamlContent,
	); err != nil {
		log.Fatalf("saving yaml file %v", err)
	}

	// Chunk and cache the profile for BM25 retrieval in match/cover-letter.
	cacheChunks(yamlContent)
}

// cacheChunks splits the master profile into chunks and persists them so
// match and cover-letter can retrieve relevant sections without re-parsing.
// Failure is non-fatal — callers fall back to the full-profile prompt.
func cacheChunks(yamlContent string) {
	chunks, err := rag.ChunkProfile([]byte(yamlContent))
	if err != nil {
		fmt.Printf("⚠️  Skipping profile chunking — %v\n", err)
		return
	}
	if len(chunks) == 0 {
		return
	}

	store := rag.NewChunkStore()
	if err := store.Update(chunks, "generated/master_profile.yaml"); err != nil {
		fmt.Printf("⚠️  Skipping chunk cache — %v\n", err)
		return
	}
	if err := store.Save(rag.ChunksFileName); err != nil {
		fmt.Printf("⚠️  Skipping chunk cache — %v\n", err)
		return
	}

	fmt.Printf("✅ Profile chunked into %d sections (%s)\n", len(chunks), rag.ChunksFileName)
}
