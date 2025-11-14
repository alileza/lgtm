# Internal Structure Contracts

**Date**: 2025-11-14  
**Purpose**: Define Go file organization and function boundaries

## File Structure

```
lgtm/
├── main.go              # CLI entry point with urfave/cli (package main)
├── slack.go             # Slack websocket client and message processing 
├── github.go            # GitHub API client and PR approval logic
├── config.go            # Configuration and flag handling
├── matcher.go           # Pattern matching logic
├── go.mod               # Module definition (module lgtm)
└── go.sum               # Dependencies
```

**All files use `package main` - no internal packages needed for this simple CLI application.**

## Core Interfaces

### 1. MessageProcessor

**File**: `slack.go`

```go
type MessageProcessor interface {
    // ProcessMessage handles incoming Slack messages
    // Returns true if message was processed, false if ignored
    ProcessMessage(ctx context.Context, msg *SlackMessage) (bool, error)
    
    // Start begins message processing
    Start(ctx context.Context) error
    
    // Stop gracefully shuts down message processing
    Stop(ctx context.Context) error
}
```

**Implementation**: `SlackClient`

**Responsibilities**:
- Connect to Slack via Socket Mode
- Filter messages by channel (if configured)
- Pass messages to pattern matcher
- Handle WebSocket reconnections

### 2. PatternMatcher

**File**: `matcher.go`

```go
type PatternMatcher interface {
    // Match tests if a message matches the configured pattern
    Match(message string) (*PatternMatch, error)
    
    // ExtractPRReferences finds GitHub PR references in matched text
    ExtractPRReferences(text string) ([]PRReference, error)
}
```

**Implementation**: `RegexMatcher`

**Responsibilities**:
- Compile and cache regex patterns
- Extract PR URLs and numbers from message text
- Validate extracted PR references

### 3. PRApprover

**File**: `github.go`

```go
type PRApprover interface {
    // ApprovePR approves a GitHub pull request
    ApprovePR(ctx context.Context, req *ApprovalRequest) (*ApprovalResult, error)
    
    // ValidatePermissions checks if the token has required permissions
    ValidatePermissions(ctx context.Context) error
    
    // GetAuthenticatedUser returns the authenticated user info
    GetAuthenticatedUser(ctx context.Context) (*User, error)
}
```

**Implementation**: `GitHubClient`

**Responsibilities**:
- Authenticate with GitHub using personal access token
- Approve pull requests via GitHub API
- Handle rate limiting and retries
- Validate token permissions

### 4. ConfigProvider

**File**: `config.go`

```go
type ConfigProvider interface {
    // Load parses configuration from flags and environment variables
    Load() (*Config, error)
    
    // Validate checks configuration completeness and validity
    Validate(cfg *Config) error
}
```

**Implementation**: `FlagConfig`

**Responsibilities**:
- Parse command-line flags and environment variables
- Validate configuration values
- Provide default values

## Data Types

### SlackMessage

```go
type SlackMessage struct {
    Text      string    `json:"text"`
    Channel   string    `json:"channel"`
    User      string    `json:"user"`
    Timestamp string    `json:"ts"`
    ThreadTS  string    `json:"thread_ts,omitempty"`
}
```

### PatternMatch

```go
type PatternMatch struct {
    Pattern       string         `json:"pattern"`
    MatchedText   string         `json:"matched_text"`
    PRReferences  []PRReference  `json:"pr_references"`
    SourceMessage *SlackMessage  `json:"source_message"`
}
```

### PRReference

```go
type PRReference struct {
    Owner      string `json:"owner"`
    Repository string `json:"repository"`
    Number     int    `json:"number"`
    URL        string `json:"url,omitempty"`
}
```

### ApprovalRequest

```go
type ApprovalRequest struct {
    Owner         string    `json:"owner"`
    Repository    string    `json:"repository"`
    PRNumber      int       `json:"pr_number"`
    Message       string    `json:"message,omitempty"`
    SourceChannel string    `json:"source_channel"`
    SourceUser    string    `json:"source_user"`
    Timestamp     time.Time `json:"timestamp"`
}
```

### ApprovalResult

```go
type ApprovalResult struct {
    Request        *ApprovalRequest `json:"request"`
    Success        bool             `json:"success"`
    ReviewID       int64            `json:"review_id,omitempty"`
    Error          string           `json:"error,omitempty"`
    ProcessedAt    time.Time        `json:"processed_at"`
    RetryAttempts  int              `json:"retry_attempts"`
}
```

### Config

```go
type Config struct {
    GitHubToken      string `json:"github_token"`
    SlackBotToken    string `json:"slack_bot_token"`
    SlackAppToken    string `json:"slack_app_token"`
    SlackChannelID   string `json:"slack_channel_id,omitempty"`
    MessagePattern   string `json:"message_pattern"`
    DefaultOwner     string `json:"default_owner,omitempty"`
    DefaultRepo      string `json:"default_repo,omitempty"`
    LogLevel         string `json:"log_level"`
}
```

## Error Types

### Custom Errors

```go
// ConfigError represents configuration-related errors
type ConfigError struct {
    Field   string
    Message string
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("configuration error [%s]: %s", e.Field, e.Message)
}

// AuthenticationError represents authentication failures
type AuthenticationError struct {
    Service string // "github" or "slack"
    Message string
}

func (e *AuthenticationError) Error() string {
    return fmt.Sprintf("%s authentication failed: %s", e.Service, e.Message)
}

// ProcessingError represents message processing errors
type ProcessingError struct {
    Operation string
    Cause     error
}

func (e *ProcessingError) Error() string {
    return fmt.Sprintf("processing error [%s]: %v", e.Operation, e.Cause)
}
```

## Function Dependencies

```
main.go (CLI setup)
    ↓
config.go (configuration parsing)
    ↓
slack.go (message processing) ←→ matcher.go (pattern matching) ←→ github.go (PR approval)
```

**All functions in the same `package main` can call each other directly - no package boundaries.**

## Contract Guarantees

### Thread Safety
- All interfaces must be safe for concurrent use
- Implementations must handle concurrent access properly
- Context cancellation must be respected

### Error Handling
- All methods return errors for failure cases
- Errors must be wrapped with sufficient context
- Temporary vs permanent failures must be distinguished

### Context Support
- All long-running operations must accept `context.Context`
- Context cancellation must be handled gracefully
- Timeouts and deadlines must be respected

### Resource Management
- Interfaces must provide cleanup methods where appropriate
- Implementations must not leak resources (goroutines, connections)
- Graceful shutdown must be supported

## Testing Interfaces

### Mockable Design
- All interfaces can be easily mocked for testing
- No direct dependencies on external services in business logic
- Dependency injection through interface parameters

### Test Doubles

```go
// Mock implementations for testing
type MockMessageProcessor struct {
    messages chan *SlackMessage
    errors   chan error
}

type MockPatternMatcher struct {
    shouldMatch bool
    prRefs      []PRReference
}

type MockPRApprover struct {
    shouldSucceed bool
    reviewID      int64
}
```

## Function Visibility

### Exported (Public)
- Functions that may need testing or CLI access
- Main data types and structs
- Error types

### Unexported (Private)  
- Helper functions
- Implementation details
- Third-party client wrappers
- Internal state management

**Since everything is in `package main`, most functions can remain unexported (lowercase names) except those specifically needed for testing or external access. This keeps the API surface minimal while maintaining simplicity.**