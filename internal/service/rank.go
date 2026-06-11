package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/rag"
	"suprie/application_tracker/internal/ranker"
	"suprie/application_tracker/internal/repository"
)

// RunRank runs the experience ranker for the given JD and persists the result.
// Run this before cover-letter so the two LLM calls are in separate processes.
func RunRank(jdID int, masterProfilePath string, repo repository.JobDescriptionRepository) {
	jd, err := repo.GetByID(context.Background(), jdID)
	if err != nil {
		log.Fatalf("loading job description id=%d: %v", jdID, err)
	}

	// Load chunks or fall back to chunking the profile on the fly.
	store := rag.NewChunkStore()
	chunks := loadOrChunk(store, masterProfilePath)
	if len(chunks) == 0 {
		log.Fatal("no profile chunks available — run 'ats parse-cv' first")
	}

	// BM25 retrieval to narrow down candidates.
	query := buildCoverLetterQuery(jd)
	retrieved, err := rag.Retrieve(query, store, 8)
	if err != nil || len(retrieved) == 0 {
		log.Fatal("BM25 retrieval returned no chunks")
	}

	fmt.Printf("🔍 BM25 retrieved %d chunks — running experience ranker\n", len(retrieved))

	// Run the ranker.
	summary := ranker.ProfileSummary(store.Chunks())
	rankerPrompt := ranker.BuildPrompt(summary, retrieved, jd)
	rankerSchema := ranker.BuildJSONSchema()

	rankerClient := llm.NewLMStudioClient("ranker")
	rankerResp, err := rankerClient.Generate(context.Background(), rankerPrompt, &rankerSchema)
	if err != nil {
		log.Fatalf("ranker failed: %v", err)
	}

	var ranked ranker.Response
	if err := json.Unmarshal([]byte(rankerResp), &ranked); err != nil {
		log.Fatalf("unmarshal ranker response: %v", err)
	}

	// Persist.
	if err := repo.UpdateRankerResult(context.Background(), jdID, rankerResp); err != nil {
		log.Fatalf("saving ranker result: %v", err)
	}

	// Print summary.
	fmt.Printf("\n=== Ranked Experiences for JD #%d ===\n", jdID)
	if jd.Company != nil {
		fmt.Printf("Company: %s\n", *jd.Company)
	}
	if jd.RoleTitle != nil {
		fmt.Printf("Role:    %s\n", *jd.RoleTitle)
	}
	fmt.Println()

	for i, exp := range ranked.SelectedExperiences {
		fmt.Printf("%d. %s (score: %.1f)\n", i+1, exp.Title, exp.Scores.FinalScore)
		fmt.Printf("   Why: %s\n", exp.WhySelected)
		fmt.Println()
	}

	fmt.Printf("Skills: ")
	for i, s := range ranked.SelectedSkills {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%s", s.Skill)
	}
	fmt.Println()

	if ranked.RecommendedNarrative != "" {
		fmt.Printf("\nNarrative: %s\n", ranked.RecommendedNarrative)
	}

	if len(ranked.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range ranked.Warnings {
			fmt.Printf("  ⚠️  %s\n", w)
		}
	}

	fmt.Println("\n✅ Ranker result saved to database.")
}

// loadOrChunk tries to load cached chunks, or chunks the profile on the fly.
func loadOrChunk(store *rag.ChunkStore, profilePath string) []rag.Chunk {
	if err := store.Load(rag.ChunksFileName); err == nil && !store.IsStale(profilePath) && store.Len() > 0 {
		return store.Chunks()
	}

	// Chunk on the fly.
	profileBytes, err := os.ReadFile(profilePath)
	if err != nil {
		return nil
	}
	chunks, err := rag.ChunkProfile(profileBytes)
	if err != nil {
		return nil
	}
	store.Update(chunks, profilePath)
	return chunks
}

// buildCoverLetterQuery constructs a retrieval query from JD fields.
func buildCoverLetterQuery(jd *domain.JobDescription) string {
	var parts []string
	if jd.RoleTitle != nil {
		parts = append(parts, *jd.RoleTitle)
	}
	parts = append(parts, jd.RequirementsJSON)
	parts = append(parts, jd.ResponsibilitiesJSON)
	return strings.Join(parts, " ")
}
