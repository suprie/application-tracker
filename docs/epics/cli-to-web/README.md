---
title: Web Interface for Application Tracker
slug: cli-to-web
milestone: "0.1.0"
status: in_progress
started: 2026-07-28
owner: Suprie
description: >
  Users manage the whole application-tracking workflow (add a job
  description, match, rank, generate a cover letter, apply, track
  companies) from a web UI instead of the CLI.
---

## Motivation

The application tracker started as a CLI (`ats parse-cv`, `ats parse-jd`,
`ats match`, `ats rank`, `ats cover-letter`, `ats apply`, `ats list`,
`ats detail`). This epic exposes that functionality through a browser so a
user is not required to run commands by hand for day-to-day tracking.

This is written as a **retroactive spec**: most of the work described below
already exists in the codebase (Gin API server, Svelte+Vite frontend served
from the same binary via `ats serve`). The epic documents that shipped
surface and captures the one remaining gap — CV upload — as open work.

## Scope

- Go API server (Gin) exposing job-description and company operations
- Svelte+Vite frontend served as a static SPA
- JD/company CRUD endpoints
- match / rank / cover-letter / apply endpoints
- Async task runner for long-running LLM operations, pollable from the UI
- Embedded SPA serving from the `ats serve` binary
- CV-upload endpoint + UI (gap — not yet built)

## Non-Goals

- Authentication / multi-user support
- Deployment or hosting setup
- Mobile UI
- Removing the CLI — it remains a supported entry point

## PRD Index

| PRD | Description | Status |
|-----|-------------|--------|
| [prd-web-cli-parity](./prd-web-cli-parity.md) | Web parity with existing CLI commands, plus CV upload | accepted |

## Architecture Decisions

### ADR-1: <title> — fill in

## Success Criteria

1. All CLI commands (parse-cv, parse-jd, match, rank, cover-letter, apply, list, detail) are reachable via API + UI.
2. `ats serve` serves the Vite-built SPA and the API from one binary.
3. Long-running LLM operations (match/rank/cover-letter) run async via the task queue and are pollable from the UI.
4. CV upload produces a usable master profile without touching the CLI.

## Risks

| Risk | Mitigation |
|------|------------|
| LLM calls (match/rank/cover-letter) are slow and could block HTTP handlers or degrade UX | Route them through the existing async task runner (`internal/task`); UI polls `/api/tasks/:id` instead of waiting on the request |
