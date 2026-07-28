package service

import (
	"log"
	"os"

	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/repository"
)

// Deps bundles the shared dependencies and configuration used by the service
// cores. Long-lived callers (the HTTP server) construct it once; the CLI
// wrappers construct one per invocation via NewDeps.
type Deps struct {
	JDRepo        repository.JobDescriptionRepository
	CompanyRepo   repository.CompanyRepository
	LLM           func(stage string) llm.LLMClient
	ProfilePath   string // master profile YAML (default generated/master_profile.yaml)
	GeneratedDir  string // output dir for cover letters etc. (default generated)
	ReaderBinPath string // Rust PDF extractor binary (default ./bin/ats-reader)
	Logger        *log.Logger
}

// NewDeps wires a Deps with default configuration derived from environment
// variables. Either repository may be nil when the caller does not need it
// (e.g. parse-cv needs no repository).
func NewDeps(jd repository.JobDescriptionRepository, company repository.CompanyRepository) Deps {
	return Deps{
		JDRepo:        jd,
		CompanyRepo:   company,
		LLM:           DefaultLLMFactory,
		ProfilePath:   envOr("ATS_PROFILE", "generated/master_profile.yaml"),
		GeneratedDir:  envOr("ATS_GENERATED_DIR", "generated"),
		ReaderBinPath: envOr("ATS_READER", "./bin/ats-reader"),
		Logger:        log.Default(),
	}
}

// DefaultLLMFactory builds an LLM client for a pipeline stage from environment
// variables (see llm.NewLMStudioClient). It is the seam used for dependency
// injection in tests.
func DefaultLLMFactory(stage string) llm.LLMClient {
	return llm.NewLMStudioClient(stage)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
