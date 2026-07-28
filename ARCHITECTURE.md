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

## Web Interface (`ats serve`)

```
browser ──► internal/server (Gin)
              │
              ├─ GET    /api/jds, /api/jds/:id        ─┐
              ├─ PATCH  /api/jds/:id  (status/url)     │
              ├─ POST   /api/jds/:id/apply              ├─► internal/service
              ├─ POST   /api/jds (parse),                │    cores return errors
              │   /match, /rank, /cover-letter           │    (DI via service.Deps)
              │   → 202 {task_id}                        ─┘
              ├─ GET/DELETE /api/tasks/:id  ◄── internal/task (in-memory runner)
              ├─ GET/POST/PUT /api/companies ─► internal/service/company.go
              └─ / + assets (embedded SPA) ◄── internal/web (//go:embed all:dist)
```

The slow LLM actions (parse/match/rank/cover-letter) are submitted to an
in-memory `internal/task.Runner` (bounded concurrency) and return
`202 {task_id}`; the SPA polls `GET /api/tasks/:id`. State is in-memory —
in-flight jobs are lost on restart (single-user local tool).

The Svelte + Vite + Tailwind frontend lives in `web/` (no `.go` files, so
`go build ./...` skips it) and builds to `internal/web/dist/`, embedded into
the binary by `internal/web/embed.go`. A committed placeholder
`internal/web/dist/index.html` keeps `go build` working before `npm run build`.

### New packages

| Package | Purpose |
|---------|---------|
| `internal/server` | Gin HTTP server: JD/company/task handlers, SPA embed + fallback |
| `internal/task` | In-memory bounded-concurrency async job runner |
| `internal/web` | `//go:embed` of the built frontend (`dist/`) |
| `internal/service/company.go` | Company CRUD service (normalize, dedup, `ErrCompanyExists`) |
| `internal/service/deps.go` | `Deps` (shared repos, LLM factory, paths) injected into cores |
| `internal/domain/company.go`, `status.go` | `Company` model; `NormalizeStatus` shared by CLI + API |
| `internal/repository/sqlite/company.go` | SQLite `CompanyRepository` |
| `internal/dto/{jd,company}_dto.go` | API request/response DTOs + mappers |
| `web/` | Svelte + Vite + Tailwind frontend |

The CLI commands share the same service cores as the server; each `Run*`
function is a thin wrapper that builds a `service.Deps` and `log.Fatalf` on
error, so CLI and web stay in sync.
