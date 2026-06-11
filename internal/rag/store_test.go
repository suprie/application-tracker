package rag

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChunkStore_SaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "chunks.json")
	profilePath := filepath.Join(dir, "profile.yaml")

	writeProfile(t, profilePath, "name: Test User\nsummary: A profile.")

	store := NewChunkStore()
	chunks := []Chunk{
		{ID: "summary", Type: ChunkSummary, Content: "A profile.", Metadata: map[string]string{"name": "Test User"}},
		{ID: "exp-0", Type: ChunkExperience, Content: "Role: Dev at Co", Metadata: map[string]string{"title": "Dev"}},
	}

	if err := store.Update(chunks, profilePath); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := store.Save(storePath); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded := NewChunkStore()
	if err := loaded.Load(storePath); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Chunks()) != 2 {
		t.Fatalf("expected 2 chunks after load, got %d", len(loaded.Chunks()))
	}
	if loaded.Chunks()[0].ID != "summary" {
		t.Errorf("chunk[0].ID = %q, want 'summary'", loaded.Chunks()[0].ID)
	}
	if loaded.IsStale(profilePath) {
		t.Error("store should NOT be stale for unchanged profile")
	}
}

func TestChunkStore_IsStale_ChangedProfile(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "chunks.json")
	profilePath := filepath.Join(dir, "profile.yaml")

	writeProfile(t, profilePath, "name: Original")
	store := NewChunkStore()
	store.Update([]Chunk{{ID: "s", Type: ChunkSummary, Content: "x"}}, profilePath)
	store.Save(storePath)

	writeProfile(t, profilePath, "name: Modified")

	loaded := NewChunkStore()
	loaded.Load(storePath)
	if !loaded.IsStale(profilePath) {
		t.Error("store should be stale after profile change")
	}
}

func TestChunkStore_IsStale_EmptyStore(t *testing.T) {
	s := NewChunkStore()
	if !s.IsStale("nonexistent.yaml") {
		t.Error("empty store should be stale")
	}
}

func TestChunkStore_Load_NonexistentFile(t *testing.T) {
	s := NewChunkStore()
	err := s.Load("/nonexistent/path/chunks.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestChunkStore_Update_ReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profile.yaml")
	storePath := filepath.Join(dir, "chunks.json")

	writeProfile(t, profilePath, "name: V1")

	store := NewChunkStore()
	store.Update([]Chunk{{ID: "a", Type: ChunkSummary, Content: "first"}}, profilePath)
	store.Save(storePath)

	store.Update([]Chunk{{ID: "b", Type: ChunkExperience, Content: "second"}}, profilePath)
	store.Save(storePath)

	loaded := NewChunkStore()
	loaded.Load(storePath)
	if len(loaded.Chunks()) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(loaded.Chunks()))
	}
	if loaded.Chunks()[0].ID != "b" {
		t.Errorf("chunk ID = %q, want 'b'", loaded.Chunks()[0].ID)
	}
}

func TestChunkStore_Len(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profile.yaml")
	writeProfile(t, profilePath, "name: X")

	store := NewChunkStore()
	if store.Len() != 0 {
		t.Errorf("empty store Len = %d", store.Len())
	}

	store.Update([]Chunk{
		{ID: "1", Type: ChunkSummary, Content: "a"},
		{ID: "2", Type: ChunkExperience, Content: "b"},
		{ID: "3", Type: ChunkSkill, Content: "c"},
	}, profilePath)

	if store.Len() != 3 {
		t.Errorf("store Len = %d, want 3", store.Len())
	}
}

func writeProfile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}
}
