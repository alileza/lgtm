# Tasks: Slack-to-GitHub Bot

**Input**: Design documents from `/specs/001-slack-github-bot/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are OPTIONAL - not explicitly requested in the feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: All files in repository root
- All files use `package main` - no internal packages needed

## Implementation Strategy

**MVP (Minimum Viable Product)**: User Story 1 only - Message Pattern Detection
**Incremental Delivery**: Each user story builds incrementally and is independently testable
**Parallel Development**: Tasks marked [P] can be developed in parallel within each phase

---

## Phase 1: Project Setup

**Goal**: Initialize Go project structure and dependencies

- [x] T001 Initialize Go module with name "lgtm" in go.mod
- [x] T002 [P] Add primary dependencies to go.mod: github.com/urfave/cli/v2, github.com/slack-go/slack, github.com/google/go-github/v75/github, github.com/gofri/go-github-ratelimit/v2, golang.org/x/oauth2
- [x] T003 [P] Create basic main.go with CLI framework setup using urfave/cli/v2
- [x] T004 [P] Create config.go with Configuration struct and validation functions

**Completion Criteria**: Project compiles successfully and CLI help shows basic command structure

---

## Phase 2: Foundational Components

**Goal**: Implement core data structures and utilities needed by all user stories

- [x] T005 [P] Implement Configuration struct in config.go with all required fields (GitHubToken, SlackBotToken, SlackAppToken, SlackChannelID, MessagePattern, DefaultOwner, DefaultRepo)
- [x] T006 [P] Implement configuration validation functions in config.go (token format validation, regex compilation)
- [x] T007 [P] Create basic error types in config.go (ConfigError, AuthenticationError, ProcessingError)
- [x] T008 [P] Implement CLI flag parsing in main.go with support for all environment variables and command-line flags

**Completion Criteria**: Configuration can be loaded from flags/env vars and validated correctly

---

## Phase 3: User Story 1 - Message Pattern Detection (Priority P1)

**Goal**: Implement core MVP functionality - listen to Slack messages and detect patterns

**Independent Test**: Post messages with known patterns to test Slack channel and verify bot detects and logs them

**Acceptance Scenarios**: 
1. Bot detects "deployed to production" pattern and logs event
2. Bot ignores messages that don't match configured patterns  
3. Bot processes first matching pattern when multiple patterns configured

### Implementation Tasks:

- [x] T009 [P] [US1] Create matcher.go with PatternMatcher struct and regex compilation
- [x] T010 [P] [US1] Implement pattern matching function in matcher.go that tests message against configured regex
- [x] T011 [P] [US1] Create slack.go with SlackMessage struct and basic data types
- [x] T012 [US1] Implement Slack Socket Mode connection setup in slack.go using slack-go/slack library
- [x] T013 [US1] Implement message event handling in slack.go with channel filtering support
- [x] T014 [US1] Integrate pattern matching with message processing in slack.go
- [x] T015 [US1] Add structured logging for pattern matches and message processing events
- [x] T016 [US1] Implement graceful shutdown handling in main.go with proper cleanup
- [x] T017 [US1] Add token validation in slack.go to verify bot and app tokens on startup

**Story Completion Criteria**: 
- Bot connects to Slack successfully
- Bot receives messages from specified channel (or all channels)
- Bot matches messages against configured pattern
- Bot logs pattern matches with structured output
- Bot handles graceful shutdown

---

## Phase 4: User Story 2 - GitHub Integration Actions (Priority P2)

**Goal**: Add GitHub PR approval functionality when patterns are detected

**Independent Test**: Post messages with PR approval patterns and verify bot approves corresponding GitHub PRs

**Acceptance Scenarios**:
1. Bot detects "LGTM" pattern with PR URL and approves the GitHub PR
2. Bot logs appropriately when PR is already merged
3. Bot handles GitHub API errors with retry logic

### Implementation Tasks:

- [ ] T018 [P] [US2] Create github.go with GitHub client setup using google/go-github and rate limiting wrapper
- [ ] T019 [P] [US2] Implement GitHub authentication and token validation in github.go
- [ ] T020 [P] [US2] Add PR reference extraction logic to matcher.go (extract URLs and PR numbers)
- [ ] T021 [US2] Implement PR approval function in github.go using GitHub CreateReview API
- [ ] T022 [US2] Add error handling and retry logic for GitHub API calls in github.go
- [ ] T023 [US2] Integrate GitHub PR approval with Slack message processing in slack.go
- [ ] T024 [US2] Add logging for GitHub operations (successful approvals, errors, retries)
- [ ] T025 [US2] Implement PR reference validation (check if PR exists and is approvable)

**Story Completion Criteria**:
- Bot extracts PR references from Slack messages
- Bot authenticates with GitHub successfully  
- Bot approves PRs via GitHub API
- Bot handles API errors gracefully with retries
- Bot logs all GitHub operations for audit

---

## Phase 5: User Story 3 - Advanced Configuration Management (Priority P3)

**Goal**: Add operational flexibility with advanced configuration options

**Independent Test**: Deploy bot with different configuration methods and verify correct settings applied

**Acceptance Scenarios**:
1. Bot loads configuration from multiple channel-pattern mappings
2. Command-line flags override config file settings
3. Bot applies new configuration without restart (if reload signal supported)

### Implementation Tasks:

- [ ] T026 [P] [US3] Extend Configuration struct in config.go to support multiple channel configurations
- [ ] T027 [P] [US3] Add JSON/YAML configuration file support in config.go
- [ ] T028 [P] [US3] Implement configuration file loading with precedence rules (CLI flags > env vars > config file)
- [ ] T029 [US3] Add support for multiple pattern-action combinations in matcher.go
- [ ] T030 [US3] Extend Slack client to support multiple channel monitoring in slack.go
- [ ] T031 [US3] Add configuration validation for complex scenarios (multiple channels, patterns)
- [ ] T032 [US3] Implement validate subcommand in main.go for configuration testing
- [ ] T033 [US3] Add version subcommand in main.go with build information

**Story Completion Criteria**:
- Bot supports configuration files (JSON/YAML)
- Bot handles multiple channel-pattern combinations
- Bot provides validate command for configuration testing
- Bot shows version information
- Configuration precedence works correctly

---

## Phase 6: Polish & Cross-Cutting Concerns

**Goal**: Production readiness improvements and operational features

- [ ] T034 [P] Add comprehensive error messages with actionable guidance for common failures
- [ ] T035 [P] Implement log level configuration (debug, info, warn, error) 
- [ ] T036 [P] Add startup validation sequence that checks all tokens and connections
- [ ] T037 [P] Implement proper signal handling (SIGINT, SIGTERM) for graceful shutdown
- [ ] T038 [P] Add runtime metrics and health check endpoints (optional)
- [ ] T039 [P] Create comprehensive README.md with setup and usage instructions
- [ ] T040 [P] Add example configuration files and deployment scripts

**Completion Criteria**: Bot is production-ready with proper error handling, logging, and documentation

---

## Dependencies & Execution Order

### Story Dependencies
- **User Story 1** (US1): No dependencies - can be implemented after foundational components
- **User Story 2** (US2): Depends on US1 completion (needs pattern matching and Slack integration)  
- **User Story 3** (US3): Independent of US2 - can be implemented in parallel after US1

### Suggested MVP Scope
**MVP = User Story 1 only**: Message pattern detection with Slack integration provides immediate value for automated message monitoring.

### Parallel Execution Examples

**Phase 3 (US1) Parallel Tasks**:
```
Parallel Group A: T009, T010, T011 (different files, no dependencies)
Sequential: T012 → T013 → T014 → T015 → T016 → T017
```

**Phase 4 (US2) Parallel Tasks**:  
```
Parallel Group A: T018, T019, T020 (different files/functions)
Sequential: T021 → T022 → T023 → T024 → T025
```

**Phase 5 (US3) Parallel Tasks**:
```  
Parallel Group A: T026, T027, T029 (independent config features)
Sequential: T028 → T030 → T031 → T032 → T033
```

**Cross-Phase Parallelism**:
- US2 and US3 can be developed in parallel after US1 is complete
- Polish tasks (Phase 6) can start once any user story is complete

---

## Task Summary

**Total Tasks**: 40
- **Setup & Foundational**: 8 tasks
- **User Story 1 (MVP)**: 9 tasks  
- **User Story 2**: 8 tasks
- **User Story 3**: 8 tasks
- **Polish**: 7 tasks

**Parallel Opportunities**: 15+ tasks marked [P] can run in parallel within their phases

**Independent Test Criteria**: Each user story has clear acceptance scenarios and can be tested independently

**Format Validation**: ✅ All tasks follow required checklist format with ID, story labels, and file paths