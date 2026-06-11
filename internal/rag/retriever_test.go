package rag

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRetrieve_BM25_TopK(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profile.yaml")
	writeProfile(t, profilePath, "name: X")

	store := NewChunkStore()
	chunks := []Chunk{
		{ID: "exp-go", Type: ChunkExperience, Content: "Go backend engineer with Kubernetes and Docker experience building cloud infrastructure"},
		{ID: "exp-python", Type: ChunkExperience, Content: "Python data engineer with machine learning and TensorFlow experience"},
		{ID: "exp-frontend", Type: ChunkExperience, Content: "Frontend developer with React TypeScript and CSS experience"},
		{ID: "skill-languages", Type: ChunkSkill, Content: "Skills — Languages: Go, Rust, Python, TypeScript"},
	}
	store.Update(chunks, profilePath)

	// Query about Kubernetes/Go should rank exp-go highest.
	got, err := Retrieve("Go Kubernetes cloud infrastructure", store, 2)
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(got))
	}
	if got[0].ID != "exp-go" {
		t.Errorf("top chunk = %q, want 'exp-go'", got[0].ID)
	}
}

func TestRetrieve_BM25_KGreaterThanLen(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profile.yaml")
	writeProfile(t, profilePath, "name: X")

	store := NewChunkStore()
	store.Update(
		[]Chunk{{ID: "only", Type: ChunkSummary, Content: "the only chunk"}},
		profilePath,
	)

	got, err := Retrieve("only", store, 100)
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 chunk (k > len), got %d", len(got))
	}
}

func TestRetrieve_BM25_EmptyStore(t *testing.T) {
	store := NewChunkStore()
	got, err := Retrieve("query", store, 5)
	if err != nil {
		t.Fatalf("Retrieve empty store: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil from empty store, got %v", got)
	}
}

func TestRetrieve_BM25_KZeroDefaults(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profile.yaml")
	writeProfile(t, profilePath, "name: X")

	store := NewChunkStore()
	chunks := make([]Chunk, 10)
	for i := 0; i < 10; i++ {
		chunks[i] = Chunk{ID: string(rune('a' + i)), Type: ChunkSummary, Content: "content " + strings.Repeat("x ", i+1)}
	}
	store.Update(chunks, profilePath)

	got, err := Retrieve("content", store, 0)
	if err != nil {
		t.Fatalf("Retrieve k=0: %v", err)
	}
	if len(got) != 5 {
		t.Errorf("k=0 should default to 5, got %d", len(got))
	}
}

func TestRetrieve_BM25_KeywordMatch(t *testing.T) {
	// Verifies that BM25 ranks exact keyword matches higher than unrelated chunks.
	chunks := []Chunk{
		{ID: "k8s", Type: ChunkExperience, Content: "Built Kubernetes operators and managed Docker containers for cloud deployments"},
		{ID: "unrelated", Type: ChunkExperience, Content: "Managed restaurant operations and customer service"},
	}

	got, _ := Retrieve("Kubernetes Docker cloud", &ChunkStore{chunks: chunks}, 2)
	if got[0].ID != "k8s" {
		t.Errorf("BM25 should rank keyword match first, got %q", got[0].ID)
	}
}

func TestTokenize(t *testing.T) {
	tokens := tokenize("Built Kubernetes operators (Go) — 2024!")
	if len(tokens) == 0 {
		t.Fatal("expected non-empty tokens")
	}
	for _, tok := range tokens {
		if len(tok) < 2 {
			t.Errorf("token %q too short", tok)
		}
		if tok != strings.ToLower(tok) {
			t.Errorf("token %q not lowercased", tok)
		}
	}
	// Punctuation and single-char tokens should be dropped.
	for _, tok := range tokens {
		if tok == "(" || tok == ")" || tok == "—" || tok == "!" {
			t.Errorf("punctuation should be dropped, got %q", tok)
		}
	}
}

func TestTokenize_Empty(t *testing.T) {
	tokens := tokenize("")
	if len(tokens) != 0 {
		t.Errorf("expected empty tokens, got %v", tokens)
	}
}

func TestTokenize_Numbers(t *testing.T) {
	tokens := tokenize("Go 1.25 with 99th percentile")
	found := false
	for _, tok := range tokens {
		if tok == "25" || tok == "99" {
			found = true
		}
	}
	if !found {
		t.Error("expected numeric tokens to be preserved")
	}
}
