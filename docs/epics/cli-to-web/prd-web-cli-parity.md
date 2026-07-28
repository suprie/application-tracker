---
prd: prd-web-cli-parity
epic: cli-to-web
milestone: "0.1.0"
status: accepted
description: >
  Give every CLI command a web-reachable equivalent — API endpoint plus
  frontend page — so the browser is a complete substitute for the CLI in
  day-to-day use.
---

## Overview

The CLI (`cmd/ats/main.go`) implements parse-cv, parse-jd, match, rank,
cover-letter, apply, list, detail, migrate, and serve. This PRD tracks
mapping each user-facing command to a web path: a Gin endpoint under
`internal/server/`, backed by the same `internal/service` functions the CLI
calls, and a Svelte page under `web/src/pages/` to drive it.

## Problem Statement

Using the tracker day-to-day currently requires shelling out to `ats <cmd>`
for actions like matching a JD against a profile or generating a cover
letter. Most of this has already been ported to a web server + SPA; one
command (`parse-cv`, which builds the master profile from a CV) has no web
equivalent, forcing a drop back to the CLI for onboarding a new profile.

## Goals

| Priority | Goal |
|----------|------|
| P0 | Every JD/company CRUD and action command (list, detail, match, rank, cover-letter, apply) has an API endpoint and UI page |
| P0 | Long-running LLM operations run through the async task runner, not inline in the request |
| P0 | CV upload (parse-cv) has an API endpoint and UI page |
| P1 | `ats serve` remains the single entry point serving both API and built SPA assets |

## Non-Goals

- Auth/multi-user
- Deployment/hosting
- Mobile UI
- Removing `migrate` or other CLI-only ops commands from the CLI

## Design

_Stub — fill in once work on Phase 3/4 starts. Existing Phases 1-2 code is
the de facto design for those parts (see `internal/server/`, `web/src/`)._

## Phases

### Phase 1 — Backend API layer

- [x] `cli-web/jd-crud` JD list/get/create/patch endpoints
- [x] `cli-web/jd-actions` match/rank/cover-letter/apply endpoints
- [x] `cli-web/company-crud` company list/get/create/update endpoints
- [x] `cli-web/task-queue` async task runner + poll/cancel endpoints

### Phase 2 — Frontend SPA

- [x] `cli-web/spa-scaffold` Svelte+Vite app scaffold with router
- [x] `cli-web/spa-pages` List/Detail/New/Companies/CompanyDetail pages
- [x] `cli-web/spa-embed` embed built assets into Go binary, serve via `ats serve`

### Phase 3 — CV upload gap

- [ ] `cli-web/cv-upload-api` POST endpoint accepting a CV file, running parse-cv, writing the master profile
- [ ] `cli-web/cv-upload-ui` upload page/form wired to the new endpoint

### Phase 4 — Verification

- [ ] `cli-web/parity-check` confirm every CLI command has an equivalent API+UI path; update epic README PRD Index status

## Open Questions

- Where should the uploaded master profile be stored — overwrite `generated/master_profile.yaml`, or support multiple named profiles?
- Should CV upload run synchronously or go through the async task runner like match/rank/cover-letter?
