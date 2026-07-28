package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/rag"
	"suprie/application_tracker/internal/ranker"
	"suprie/application_tracker/internal/repository"
)

// RunRank is the CLI wrapper for Rank. Run this before cover-letter so the
// two LLM calls are in separate processes.
func RunRank(jdID int, masterProfilePath string, repo repository.JobDescriptionRepository) {
	d := NewDeps(repo, nil)
	d.ProfilePath = masterProfilePath

	ranked, err := Rank(context.Background(), d, jdID)
	if err != nil {
		log.Fatalf("%v", err)
	}

	jd, err := repo.GetByID(context.Background(), jdID)
	if err != nil {
		log.Fatalf("reloading job description id=%d: %v", jdID, err)
	}
	printRankResult(jd, ranked)
	fmt.Println("\n✅ Ranker result saved to database.")
}

// Rank runs the experience ranker for the given JD and persists the result.
func Rank(ctx context.Context, d Deps, jdID int) (*ranker.Response, error) {
	jd, err := d.JDRepo.GetByID(ctx, jdID)
	if err != nil {
		return nil, fmt.Errorf("loading job description id=%d: %w", jdID, err)
	}

	// Load chunks or fall back to chunking the profile on the fly.
	store := rag.NewChunkStore()
	chunks := loadOrChunk(store, d.ProfilePath)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no profile chunks available — run 'ats parse-cv' first")
	}

	// BM25 retrieval to narrow down candidates.
	query := buildCoverLetterQuery(jd)
	retrieved, err := rag.Retrieve(query, store, 8)
	if err != nil {
		return nil, fmt.Errorf("BM25 retrieval: %w", err)
	}
	if len(retrieved) == 0 {
		return nil, fmt.Errorf("BM25 retrieval returned no chunks")
	}

	fmt.Printf("🔍 BM25 retrieved %d chunks — running experience ranker\n", len(retrieved))

	// Run the ranker.
	summary := ranker.ProfileSummary(store.Chunks())
	rankerPrompt := ranker.BuildPrompt(summary, retrieved, jd)
	rankerSchema := ranker.BuildJSONSchema()

	client := d.LLM("ranker")
	rankerResp, err := client.Generate(ctx, rankerPrompt, &rankerSchema)
	if err != nil {
		return nil, fmt.Errorf("ranker failed: %w", err)
	}

	var ranked ranker.Response
	if err := json.Unmarshal([]byte(rankerResp), &ranked); err != nil {
		return nil, fmt.Errorf("unmarshal ranker response: %w", err)
	}

	if err := d.JDRepo.UpdateRankerResult(ctx, jdID, rankerResp); err != nil {
		return nil, fmt.Errorf("saving ranker result: %w", err)
	}

	return &ranked, nil
}

// printRankResult prints the ranked experiences to stdout (CLI only).
func printRankResult(jd *domain.JobDescription, ranked *ranker.Response) {
	fmt.Printf("\n=== Ranked Experiences for JD #%d ===\n", jd.ID)
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
