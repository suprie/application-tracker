package service

import (
	"context"
	"fmt"
	"log"

	"suprie/application_tracker/internal/cvextractor"
	"suprie/application_tracker/internal/cvparser"
	"suprie/application_tracker/internal/fileutil"
	"suprie/application_tracker/internal/rag"
	"suprie/application_tracker/internal/textutils"
)

// RunParseCV is the CLI wrapper for ParseCV.
func RunParseCV(filename string) {
	d := NewDeps(nil, nil)
	if err := ParseCV(context.Background(), d, filename); err != nil {
		log.Fatalf("%v", err)
	}
}

// ParseCV extracts text from a CV PDF, parses it into a YAML profile via the
// LLM, saves it to d.ProfilePath, and caches profile chunks for BM25
// retrieval. It is the server-callable core (though currently only the CLI
// invokes it, since the web does not parse CVs).
func ParseCV(ctx context.Context, d Deps, filename string) error {
	extractor := cvextractor.RustPDFExtractor{BinaryPath: d.ReaderBinPath}

	fmt.Printf("Processing: %s\n", filename)

	result, err := extractor.Extract(ctx, filename)
	if err != nil {
		return fmt.Errorf("extracting CV: %w", err)
	}

	prompt := cvparser.BuildCVParserPrompt(textutils.NormalizeText(result.Text))

	client := d.LLM("parse_cv")
	response, err := client.Generate(ctx, prompt, nil)
	if err != nil {
		return fmt.Errorf("generating response: %w", err)
	}

	yamlContent := fileutil.StripMarkdownFence(response)
	if err := fileutil.SaveYAML(d.ProfilePath, yamlContent); err != nil {
		return fmt.Errorf("saving yaml file: %w", err)
	}

	// Chunk and cache the profile for BM25 retrieval in match/cover-letter.
	cacheChunks(d.ProfilePath, yamlContent)
	return nil
}

// cacheChunks splits the master profile into chunks and persists them so
// match and cover-letter can retrieve relevant sections without re-parsing.
// Failure is non-fatal — callers fall back to the full-profile prompt.
func cacheChunks(profilePath, yamlContent string) {
	chunks, err := rag.ChunkProfile([]byte(yamlContent))
	if err != nil {
		fmt.Printf("⚠️  Skipping profile chunking — %v\n", err)
		return
	}
	if len(chunks) == 0 {
		return
	}

	store := rag.NewChunkStore()
	if err := store.Update(chunks, profilePath); err != nil {
		fmt.Printf("⚠️  Skipping chunk cache — %v\n", err)
		return
	}
	if err := store.Save(rag.ChunksFileName); err != nil {
		fmt.Printf("⚠️  Skipping chunk cache — %v\n", err)
		return
	}

	fmt.Printf("✅ Profile chunked into %d sections (%s)\n", len(chunks), rag.ChunksFileName)
}
