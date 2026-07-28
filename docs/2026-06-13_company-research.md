# Company Research Feature — Implementation Plan

**Date**: 2026-06-13  
**Goal**: Add a company research database with backend API and web frontend.  
Integrate company lookup into JD parsing and cover letter generation.

## Overview

Currently `job_descriptions.company` is free-text from the LLM parser. This
feature adds a normalized `companies` table with manual research data, a REST
API (Gin), a Svelte+Vite web frontend (embedded in the Go binary), and soft
integration into the JD → cover-letter pipeline.

## Key Design Decisions

- **Gin** for the HTTP server — lighter than `net/http` for REST APIs, wide
  ecosystem, idiomatic Go.
- **Svelte + Vite** for the frontend — compiles to static assets embedded via
  `//go:embed`. Svelte is lightweight with no virtual DOM, Vite gives fast HMR
  during development.
- **Soft integration**: JD parsing does company lookup by `normalized_name` but
  doesn't block if unmatched — just logs a hint.
- **Normalized name matching**: Strip suffixes (Inc, Ltd, LLC, Corp, GmbH,
  etc.) + lowercase for dedup lookups.
- **Single binary**: The Vite-built SPA output goes into `web/dist/`, embedded
  in the Go binary. `ats serve` runs the full stack.

## Schema

```sql
CREATE TABLE companies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    normalized_name TEXT NOT NULL UNIQUE,
    website_url TEXT,
    industry TEXT,
    size TEXT,
    country TEXT,
    notes TEXT,
    source TEXT,
    research_summary TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## API Routes (Gin)

```
POST   /api/companies          — create company (JSON body)
GET    /api/companies          — list all companies (JSON, optional ?q= search)
GET    /api/companies/:id      — get one company (JSON)
PUT    /api/companies/:id      — update company (JSON body)
GET    /                        — serve embedded Svelte SPA
```

## Frontend Routes (Svelte)

```
/                              — company list + search + add button
/companies/:id                 — company detail + edit form
/companies/new                 — add company form
```

## File Inventory

### New Files

| File | Purpose |
|------|---------|
| `internal/migration/migrations/000005_create_companies.up.sql` | Create companies table |
| `internal/migration/migrations/000005_create_companies.down.sql` | Drop companies table |
| `internal/domain/company.go` | Company domain model |
| `internal/dto/company_dto.go` | JSON request/response DTOs |
| `internal/repository/sqlite/company.go` | SQLite company CRUD |
| `internal/textutils/normalize_company.go` | Company name normalization |
| `internal/textutils/normalize_company_test.go` | Normalization tests |
| `internal/service/company.go` | Company service layer |
| `internal/server/server.go` | Gin server setup, routes, middleware |
| `internal/server/company_handler.go` | REST handlers for /api/companies |
| `web/package.json` | Svelte+Vite project config |
| `web/vite.config.js` | Vite config (output to web/dist/) |
| `web/src/app.html` | Svelte shell |
| `web/src/routes/+page.svelte` | Company list + add form |
| `web/src/routes/companies/[id]/+page.svelte` | Company detail + edit |
| `web/src/lib/api.js` | Fetch helper for backend API |

### Modified Files

| File | Change |
|------|--------|
| `internal/repository/repository.go` | Add `CompanyRepository` interface |
| `internal/service/parse_jd.go` | Add company lookup after extraction |
| `internal/service/cover_letter.go` | Look up company, pass research to prompt |
| `internal/coverletter/prompt.go` | Accept optional `*domain.Company` context |
| `cmd/ats/main.go` | Add `serve`, `company add`, `company list` commands |
| `go.mod` | Add `github.com/gin-gonic/gin` dependency |

## Dependency Order

```
Step 1-2:  Migration + Domain model        (A — foundation)
Step 3:    Repository interface + impl      (A — data layer)
Step 4:    Normalization utility + tests    (A — shared utility)
Step 5:    DTOs                             (A — shared types)
Step 6:    Service layer                    (B — business logic)
Step 7:    Gin HTTP server + handlers       (B — API)
Step 8:    Svelte+Vite frontend             (B — web UI)
Step 9:    Wire into JD parsing             (B — integration)
Step 10:   Wire into cover letter prompt    (B — integration)
Step 11:   CLI serve command                (B — entry point)
Step 12:   CLI company commands             (C — convenience)
Step 13:   Update ARCHITECTURE.md           (C — docs)
```

## Implementation Notes

### Migration
Follow existing pattern: `00000X_name.up.sql` / `00000X_name.down.sql`, parsed
by `internal/migration/migrate.go` from embedded `migrations/` directory.

### Repository Pattern
- Interface in `internal/repository/repository.go`: `GetByID`, `GetByNormalizedName`,
  `Create`, `List`, `Update`
- SQLite impl in `internal/repository/sqlite/company.go`:
  - Constructor: `NewCompanyRepository(db *sql.DB)`
  - Compile-time interface check
  - Shared `scanner` interface for row scanning
  - `selectColumns`/`insertColumns` string constants
  - `scanCompany` helper function

### Normalization
`NormalizeCompanyName(raw string) string`:
- Trim whitespace, lowercase
- Strip: `Inc.`, `Inc`, `Ltd.`, `Ltd`, `LLC`, `Corp.`, `Corp`, `Corporation`,
  `GmbH`, `S.A.`, `B.V.`, `Pvt Ltd`, `Limited`, `Co.`, `Co`
- Remove punctuation, collapse whitespace

### JD Parsing Integration
In `RunParseJD`, after `mapToJobDescription`:
```go
if repo != nil && companyRepo != nil {
    ctx := context.Background()
    if jd.Company != nil {
        c, err := companyRepo.GetByNormalizedName(ctx, textutils.NormalizeCompanyName(*jd.Company))
        if err == nil && c != nil {
            log.Printf("✅ Company matched: %s (id=%d)", c.Name, c.ID)
        } else {
            log.Printf("📝 Company '%s' not in database — add via web UI", *jd.Company)
        }
    }
}
```

### Cover Letter Integration
In `buildCoverLetterPrompt`, after loading the JD, look up the company. If
found and has research data, inject a `=== COMPANY RESEARCH ===` section into
the prompt with industry, size, notes, and research_summary.

### Gin Server
```go
r := gin.Default()
r.GET("/", serveSPA("web/dist"))
r.StaticFS("/assets/", embeddedAssets)
api := r.Group("/api")
{
    api.POST("/companies", createCompany)
    api.GET("/companies", listCompanies)
    api.GET("/companies/:id", getCompany)
    api.PUT("/companies/:id", updateCompany)
}
```

### Svelte Frontend
- `@sveltejs/adapter-static` for SPA output
- Tailwind CSS for styling (lightweight, no JS runtime needed)
- Fetch wrapper in `src/lib/api.js` calls the Gin backend
- Build: `cd web && npm run build` → output to `web/dist/`
- Dist embedded via `//go:embed web/dist/*` in Go

## Verification

1. `go build ./...` — compiles cleanly
2. `./main migrate up` — creates companies table
3. `cd web && npm install && npm run build` — builds frontend
4. `./main serve` — starts on :8080, browse to http://localhost:8080
5. Add a company via the web UI (or `curl -X POST ...`)
6. `./main parse-jd some_jd.pdf` — shows "Company matched" or hint to add
7. `./main cover-letter <id>` — if company exists, prompt includes research
8. `go test ./...` — all existing + new normalization tests pass
