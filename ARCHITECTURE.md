# Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      cmd/ats (CLI)                          │
│  parse-cv │ parse-jd │ match │ rank │ cover-letter │ ...    │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                   internal/service                          │
│  Orchestration layer — one file per command.                │
│                                                             │
│  parse_cv.go   parse_jd.go   match.go   apply.go            │
│  cover_letter.go   rank.go   list.go   detail.go            │
│                                                             │
│  Each service wires together extraction, LLM, repository.   │
└──┬──────────┬──────────┬──────────┬──────────┬─────────────┘
   │          │          │          │          │
   ▼          ▼          ▼          ▼          ▼
┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────────┐
│cvextr│ │cvpars│ │jdextr│ │matchr│ │coverlettr│
│actor │ │er    │ │actor │ │er    │ │(template)│
│      │ │      │ │      │ │      │ │          │
│ Rust │ │ LLM  │ │LLM   │ │LLM   │ │LLM+LaTeX │
│ PDF→ │ │CV→   │ │JD→   │ │fit   │ │cover     │
│ text │ │YAML  │ │JSON  │ │score │ │letter    │
└──────┘ └──────┘ └──────┘ └──────┘ └──────────┘

┌──────────┐  ┌──────────┐  ┌──────────────────────┐
│  ranker  │  │   rag    │  │        llm           │
│          │  │          │  │                      │
│ LLM      │  │ Chunk    │  │ LMStudioClient       │
│ re-rank  │  │ Profile  │  │   ├─ OpenAIProvider   │
│ BM25→    │  │ BM25     │  │   ├─ DeepSeekProvider │
│ curated  │  │ Retrieve │  │   └─ OllamaProvider   │
│ picks    │  │ Store    │  │                      │
└──────────┘  └──────────┘  └──────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                    Data Layer                                │
│                                                              │
│  domain/              repository/          repository/sqlite/│
│  JobDescription       interface            SQLite impl       │
│  + status consts      + CRUD               + migrations/    │
│                       + ranker result       (4 migrations)   │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                    Utilities                                 │
│  textutils/         fileutil/          migration/            │
│  NormalizeText      SaveYAML           golang-migrate driver │
│                     StripMarkdownFence                       │
└──────────────────────────────────────────────────────────────┘
```

## Data Flow

### CV Processing
```
PDF → [cvextractor:Rust] → text → [cvparser:LLM] → master_profile.yaml
                                                         │
                                               [rag:ChunkProfile]
                                                         │
                                                  chunks.json
```

### JD Processing
```
PDF → [os.ReadFile] → text → [jdextractor:LLM+JSON Schema] → SQLite
```

### Match (with RAG)
```
JD ← SQLite   Profile ← master_profile.yaml
        │            │
        ▼            ▼
   [rag:BM25 retrieve top-5 chunks]
        │
        ▼
   [matcher:LLM] → fit_score + summary → SQLite
```

### Cover Letter (with Ranker)
```
Step 1:  ats rank <id>
         JD + chunks ──→ [rag:BM25 retrieve 8] ──→ [ranker:LLM] ──→ SQLite

Step 2:  ats cover-letter <id>
         JD + cached ranker result ──→ [coverletter:LLM] ──→ LaTeX ──→ PDF
```

### Status Lifecycle
```
Draft → Fit match → Applied → Rejected / Offer
```

## Package Inventory

| Package | Files | Tests | Purpose |
|---------|-------|-------|---------|
| `cmd/ats` | 1 | — | CLI entry point |
| `internal/domain` | 1 | — | `JobDescription` model |
| `internal/dto` | 1 | — | Shared DTOs |
| `internal/repository` | 1 | — | Repository interface |
| `internal/repository/sqlite` | 2 | 10 | SQLite CRUD |
| `internal/migration` | 2 | 4 | Schema migrations |
| `internal/llm` | 5 | 15 | LLM client + providers |
| `internal/cvextractor` | 2 | 5 | PDF→text (Rust) |
| `internal/cvparser` | 1 | 3 | CV→YAML prompt |
| `internal/jdextractor` | 3 | 12 | JD→JSON prompt + schema |
| `internal/matcher` | 3 | — | Fit-match prompt + schema |
| `internal/coverletter` | 5 | 3 | Cover letter prompt + schema + LaTeX |
| `internal/ranker` | 3 | — | Experience ranker prompt + schema |
| `internal/rag` | 4 | 17 | Chunking + BM25 retrieval + store |
| `internal/service` | 8 | — | Orchestration layer |
| `internal/textutils` | 1 | 6 | Text normalization |
| `internal/fileutil` | 1 | 5 | File I/O helpers |
| `rust/` | 2 | — | Rust PDF extractor |
