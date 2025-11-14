# Data Model: Slack-to-GitHub Bot

**Date**: 2025-11-14  
**Feature**: Slack-to-GitHub Bot  
**Purpose**: Define core data structures and entities

## Core Entities

### 1. Configuration

**Purpose**: Runtime configuration for the bot including tokens, patterns, and operational settings.

**Attributes**:
- `GitHubToken` (string, required) - GitHub personal access token
- `SlackBotToken` (string, required) - Slack bot user OAuth token (`xoxb-*`)
- `SlackAppToken` (string, required) - Slack app-level token (`xapp-*`)
- `SlackChannelID` (string, optional) - Specific channel ID to monitor
- `MessagePattern` (string, optional) - Regex pattern for message matching (default: match all)
- `DefaultOwner` (string, optional) - Default GitHub repository owner
- `DefaultRepo` (string, optional) - Default GitHub repository name

**Validation Rules**:
- All token fields must be non-empty when provided
- SlackBotToken must start with `xoxb-`
- SlackAppToken must start with `xapp-`
- MessagePattern must be valid regex syntax
- GitHubToken must have PR approval permissions (validated at startup)

**State Transitions**: Immutable after validation - no runtime state changes.

### 2. SlackMessage

**Purpose**: Represents incoming Slack messages with metadata for processing.

**Attributes**:
- `Text` (string) - Message content
- `Channel` (string) - Channel ID where message was posted
- `User` (string) - User ID who posted the message
- `Timestamp` (string) - Slack message timestamp
- `ThreadTimestamp` (string, optional) - Parent thread timestamp if reply

**Validation Rules**:
- Text must be non-empty
- Channel must match configured channel (if specified)
- User must not be the bot itself (prevent self-processing)

**Relationships**: No persistent relationships - processed and discarded.

### 3. PatternMatch

**Purpose**: Results of pattern matching against Slack messages.

**Attributes**:
- `Pattern` (string) - Regex pattern that matched
- `MatchedText` (string) - Portion of text that matched
- `PRReferences` ([]PRReference) - Extracted PR references from the message
- `SourceMessage` (SlackMessage) - Original message that generated the match

**Validation Rules**:
- Pattern must be valid regex
- MatchedText must not be empty
- At least one PRReference required for processing

**State Transitions**: Created → Processed → Discarded

### 4. PRReference

**Purpose**: Extracted GitHub pull request information from Slack messages.

**Attributes**:
- `Owner` (string, optional) - Repository owner (may be inferred from config)
- `Repository` (string, optional) - Repository name (may be inferred from config)
- `Number` (int, required) - Pull request number
- `URL` (string, optional) - Full GitHub PR URL if extracted from URL pattern

**Validation Rules**:
- Number must be positive integer
- Owner and Repository required (either extracted or from default config)
- URL must be valid GitHub PR URL format if provided

**Relationships**: Many PRReference can exist per PatternMatch.

### 5. ApprovalRequest

**Purpose**: Request to approve a GitHub pull request with context.

**Attributes**:
- `Owner` (string, required) - Repository owner
- `Repository` (string, required) - Repository name  
- `PRNumber` (int, required) - Pull request number
- `Message` (string, optional) - Approval comment message
- `SourceChannel` (string) - Slack channel that triggered the approval
- `SourceUser` (string) - Slack user who triggered the approval
- `Timestamp` (time.Time) - When the approval was requested

**Validation Rules**:
- Owner and Repository must be non-empty
- PRNumber must be positive
- Message length limited to GitHub's comment limits (65536 characters)

**State Transitions**: Created → InProgress → Completed|Failed

### 6. ApprovalResult

**Purpose**: Result of GitHub PR approval operation.

**Attributes**:
- `Request` (ApprovalRequest) - Original approval request
- `Success` (bool) - Whether the approval succeeded
- `ReviewID` (int64, optional) - GitHub review ID if successful
- `Error` (string, optional) - Error message if failed
- `ProcessedAt` (time.Time) - When the approval was processed
- `RetryAttempts` (int) - Number of retry attempts made

**Validation Rules**:
- If Success is true, ReviewID must be provided
- If Success is false, Error must be provided
- ProcessedAt must be after Request.Timestamp

**State Transitions**: Created → Final (no further changes)

## Entity Relationships

```
Configuration (1) ──┐
                   │
SlackMessage (1) ──┼──> PatternMatch (0..1) ──> PRReference (1..n) ──> ApprovalRequest (1..n)
                   │                                                              │
                   └──────────────────────────────────────────────────────────> ApprovalResult (1)
```

## Data Flow

1. **SlackMessage** received from websocket connection
2. **Configuration** provides matching criteria and GitHub context
3. **PatternMatch** created if message matches configured patterns
4. **PRReference**(s) extracted from matched message text
5. **ApprovalRequest**(s) created for each valid PR reference
6. **ApprovalResult** generated after GitHub API interaction

## Memory Management

- All entities are short-lived (processing lifespan: seconds)
- No persistent storage - stateless operation
- Garbage collected after request processing completes
- Maximum concurrent entities limited by message processing rate (target: <100 concurrent)

## Validation Strategy

- **Input Validation**: All external inputs validated at creation
- **Configuration Validation**: Performed once at startup with clear error messages
- **Runtime Validation**: Minimal validation during processing for performance
- **Error Propagation**: Validation errors bubble up with context for debugging