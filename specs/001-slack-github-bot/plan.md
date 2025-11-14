# Implementation Plan: Slack-to-GitHub Bot

**Branch**: `001-slack-github-bot` | **Date**: 2025-11-14 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-slack-github-bot/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Go binary (module: lgtm) that listens to Slack messages via websocket connection and automatically approves GitHub pull requests when specific patterns (like "LGTM") are detected. Uses command-line flags for configuration including GitHub token, Slack tokens, channel ID, and message patterns.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.25  
**Primary Dependencies**: github.com/urfave/cli/v2, github.com/slack-go/slack, github.com/google/go-github/v75/github, github.com/gofri/go-github-ratelimit/v2, golang.org/x/oauth2  
**Storage**: N/A (stateless bot)  
**Testing**: Go standard testing (go test)  
**Target Platform**: Linux/macOS/Windows (cross-platform binary)
**Project Type**: Single CLI application  
**Performance Goals**: Process messages within 2 seconds, handle 100+ messages/minute  
**Constraints**: <5MB memory usage, startup time <10 seconds  
**Scale/Scope**: Single binary, ~5-10 Go files, supports 1 channel per instance

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Clear is Better than Clever**: Simple Go binary with clear CLI interface and straightforward logic flow
- [x] **Lazy and Pragmatic**: Minimal feature set (listen -> pattern match -> approve PR), no over-engineering
- [x] **Minimal Dependencies**: Dependencies justified - urfave/cli (user requirement), slack-go/slack (WebSocket complexity), google/go-github (API complexity)
- [x] **Clear Boundaries**: Single responsibility: message processing and PR approval
- [x] **Simple Directory Structure**: Flat Go project structure with clear package organization

*If any checks fail, document justification in Complexity Tracking section below.*

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
lgtm/
├── main.go              # CLI entry point with urfave/cli (package main)
├── slack.go             # Slack websocket client and message processing 
├── github.go            # GitHub API client and PR approval logic
├── config.go            # Configuration and flag handling
├── matcher.go           # Pattern matching logic
├── go.mod               # Module definition (module lgtm)
└── go.sum               # Dependencies
```

**Structure Decision**: Single Go package (`package main`) with multiple files organized by functionality. This keeps the implementation simple and avoids unnecessary package boundaries for a small CLI application. All code is co-located and easy to navigate.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| urfave/cli/v2 dependency | User specified CLI framework requirement | Standard library flag package lacks subcommand support and structured help |
| Slack/GitHub SDK dependencies | WebSocket and API complexity | Standard library HTTP client requires significant boilerplate for auth, websockets, and error handling |
