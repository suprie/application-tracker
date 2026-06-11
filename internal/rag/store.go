package rag

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
)

// ChunksFileName is the default path for the persisted chunk store.
const ChunksFileName = "generated/chunks.json"

// storedData is the on-disk representation of the chunk store.
type storedData struct {
	ProfileSum string  `json:"profile_sum"`
	Chunks     []Chunk `json:"chunks"`
}

// ChunkStore holds profile chunks in memory and persists them to a JSON file
// for reuse across CLI invocations. Staleness is detected by comparing the
// master profile's SHA-256 hash against the stored checksum.
type ChunkStore struct {
	chunks     []Chunk
	profileSum string
}

// NewChunkStore returns an empty ChunkStore.
func NewChunkStore() *ChunkStore {
	return &ChunkStore{}
}

// Chunks returns the stored chunks. Callers should check IsStale before using.
func (s *ChunkStore) Chunks() []Chunk {
	return s.chunks
}

// Len returns the number of stored chunks.
func (s *ChunkStore) Len() int {
	return len(s.chunks)
}

// Update replaces the in-memory data and records the profile checksum.
func (s *ChunkStore) Update(chunks []Chunk, profilePath string) error {
	sum, err := fileSum(profilePath)
	if err != nil {
		return fmt.Errorf("checksum profile: %w", err)
	}
	s.chunks = chunks
	s.profileSum = sum
	return nil
}

// Load reads a previously persisted store from path. It returns an error
// if the file does not exist or is malformed.
func (s *ChunkStore) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read store file: %w", err)
	}
	var d storedData
	if err := json.Unmarshal(data, &d); err != nil {
		return fmt.Errorf("parse store file: %w", err)
	}
	s.chunks = d.Chunks
	s.profileSum = d.ProfileSum
	return nil
}

// Save persists the store to path as JSON.
func (s *ChunkStore) Save(path string) error {
	d := storedData{
		ProfileSum: s.profileSum,
		Chunks:     s.chunks,
	}
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal store: %w", err)
	}
	if err := os.MkdirAll("generated", 0o755); err != nil {
		return fmt.Errorf("ensure generated dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write store file: %w", err)
	}
	return nil
}

// IsStale returns true when the profile file's SHA-256 differs from the stored
// checksum, or when the store is empty.
func (s *ChunkStore) IsStale(profilePath string) bool {
	if len(s.chunks) == 0 {
		return true
	}
	current, err := fileSum(profilePath)
	if err != nil {
		return true
	}
	return current != s.profileSum
}

func fileSum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h), nil
}
