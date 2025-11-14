# Research: Slack-to-GitHub Bot Implementation

**Date**: 2025-11-14  
**Feature**: Slack-to-GitHub Bot  
**Purpose**: Resolve technical dependencies and implementation approaches

## Slack WebSocket SDK Selection

### Decision: `github.com/slack-go/slack`

**Rationale**: Official community-maintained library with comprehensive Socket Mode support, active maintenance (4.9k stars, regular releases through 2024), and minimal dependencies aligning with Go best practices.

**Key Features**:
- Modern Socket Mode support (recommended over legacy RTM)
- Automatic reconnection handling
- Full Slack API coverage
- Active maintenance with community Slack channel support

**Authentication Requirements**:
- `SLACK_BOT_TOKEN`: Bot User OAuth Token (`xoxb-*`)
- `SLACK_APP_TOKEN`: App-Level Token (`xapp-*`) for Socket Mode

**Alternatives Considered**:
- `nlopes/slack`: Legacy, now redirects to slack-go/slack
- `lestrrat-go/slack`: Has mock servers but less comprehensive API coverage

## GitHub API Client Selection

### Decision: `github.com/google/go-github/v75/github`

**Rationale**: Official Google-maintained library with comprehensive GitHub API v3 coverage, active development, semantic versioning, and extensive community adoption. Full support for PR approval via `CreateReview` API.

**Key Features**:
- Complete GitHub API coverage
- Context support for cancellation/timeouts
- Built-in rate limit awareness
- Structured response handling

**Authentication**:
- Personal Access Token with OAuth2 wrapper
- Required permissions: `Contents` (Read/Write), `Pull requests` (Write), `Issues` (Read)

**Rate Limiting Enhancement**: `github.com/gofri/go-github-ratelimit/v2/github_ratelimit`
- Automatic rate limit detection and handling
- Callback support for rate limit events
- Recommended for production use

**Alternatives Considered**:
- Direct HTTP client: Too much boilerplate for auth, error handling, and rate limiting
- Third-party GitHub libraries: Less comprehensive and not officially maintained

## Pattern Matching Strategy

### Decision: Standard Go `regexp` package

**Rationale**: Built-in package sufficient for GitHub PR URL extraction and message pattern matching. No external dependency needed.

**Patterns**:
- GitHub PR URLs: `https://github\.com/([^/]+)/([^/]+)/pull/(\d+)`
- Simple PR references: `(?:#|PR-?)(\d+)`
- Custom message patterns: User-configurable regex

## Error Handling and Resilience

### Decision: Exponential backoff with circuit breaker pattern

**Rationale**: GitHub API and Slack WebSocket connections can be unreliable. Implement retry logic with exponential backoff and proper error categorization.

**Implementation**:
- Retry on transient failures (5xx, network errors)
- Exponential backoff: 2^attempt seconds
- Maximum 3 retry attempts
- Proper error categorization (auth vs permission vs network)

## Configuration Management

### Decision: Environment variables + CLI flags with `urfave/cli/v2`

**Rationale**: User specifically requested urfave/cli/v2. Environment variables provide secure token storage while CLI flags offer runtime flexibility.

**Flag Structure**:
- `GITHUB_TOKEN` (env) / `--github-token` (flag)
- `SLACK_BOT_TOKEN` (env) / `--slack-bot-token` (flag)  
- `SLACK_APP_TOKEN` (env) / `--slack-app-token` (flag)
- `SLACK_CHANNEL_ID` (env) / `--slack-channel-id` (flag)
- `SLACK_MESSAGE_PATTERN` (env) / `--slack-pattern` (flag, default: watch all)

## Security Considerations

### Token Management
- Never log token values
- Use fine-grained personal access tokens with minimal permissions
- Validate token permissions on startup with clear error messages

### Input Validation
- Validate all GitHub URLs and PR numbers
- Sanitize user input patterns
- Implement URL parsing validation

## Performance and Scale

### Memory Usage
- Target: <5MB memory usage
- Stateless design - no persistent storage
- Efficient regex compilation and reuse

### Processing Speed
- Target: Process messages within 2 seconds
- Handle 100+ messages/minute
- Serial request processing to respect rate limits

## Implementation Dependencies Summary

**Required Dependencies**:
1. `github.com/slack-go/slack` - Slack WebSocket client
2. `github.com/google/go-github/v75/github` - GitHub API client
3. `github.com/gofri/go-github-ratelimit/v2/github_ratelimit` - Rate limiting
4. `golang.org/x/oauth2` - OAuth2 for GitHub authentication
5. `github.com/urfave/cli/v2` - CLI framework (user requirement)

**Standard Library Usage**:
- `regexp` - Pattern matching
- `net/url` - URL parsing and validation
- `context` - Request cancellation and timeouts
- `encoding/json` - Configuration parsing (if needed)

All dependencies are well-maintained, have minimal transitive dependencies, and align with Go best practices for a production CLI application.