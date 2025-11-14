# CLI Interface Contract

**Date**: 2025-11-14  
**Purpose**: Define command-line interface structure and behavior

## Main Command

### `lgtm`

**Description**: Slack-to-GitHub bot that monitors Slack messages and approves GitHub pull requests.

**Usage**: 
```bash
lgtm [global options] command [command options] [arguments...]
```

**Global Options**:

| Flag | Environment Variable | Type | Required | Default | Description |
|------|---------------------|------|----------|---------|-------------|
| `--github-token` | `GITHUB_TOKEN` | string | Yes | - | GitHub personal access token |
| `--slack-bot-token` | `SLACK_BOT_TOKEN` | string | Yes | - | Slack bot user OAuth token |
| `--slack-app-token` | `SLACK_APP_TOKEN` | string | Yes | - | Slack app-level token for Socket Mode |
| `--slack-channel-id` | `SLACK_CHANNEL_ID` | string | No | "" | Specific channel ID to monitor (empty = all channels) |
| `--slack-pattern` | `SLACK_MESSAGE_PATTERN` | string | No | ".*" | Regex pattern for message matching |
| `--github-owner` | `GITHUB_OWNER` | string | No | "" | Default repository owner |
| `--github-repo` | `GITHUB_REPO` | string | No | "" | Default repository name |
| `--log-level` | `LOG_LEVEL` | string | No | "info" | Logging level (debug, info, warn, error) |
| `--help, -h` | - | - | - | - | Show help |
| `--version, -v` | - | - | - | - | Show version |

## Commands

### `run` (default)

**Description**: Start the bot to monitor Slack messages and approve GitHub PRs.

**Usage**:
```bash
lgtm run [options]
lgtm [options]  # 'run' is default command
```

**Behavior**:
1. Validate all configuration and tokens
2. Test GitHub token permissions
3. Connect to Slack via Socket Mode
4. Listen for messages matching the configured pattern
5. Extract PR references and approve them on GitHub
6. Run until interrupted (Ctrl+C)

**Exit Codes**:
- `0`: Normal shutdown
- `1`: Configuration error (invalid tokens, missing required flags)
- `2`: Authentication error (invalid or insufficient permissions)
- `3`: Connection error (unable to connect to Slack or GitHub)
- `4`: Runtime error (unexpected error during operation)

**Output**:
```
INFO: Starting LGTM bot...
INFO: Authenticated as GitHub user: username
INFO: Connected to Slack workspace: workspace-name  
INFO: Monitoring channel: #channel-name (C1234567890)
INFO: Pattern: "(?i)lgtm|looks good to me"
INFO: Bot ready - listening for messages...
INFO: Message matched pattern in #channel-name by @user
INFO: Extracted PR: owner/repo#123
INFO: Approved PR owner/repo#123 (Review ID: 987654321)
```

### `validate`

**Description**: Validate configuration and tokens without starting the bot.

**Usage**:
```bash
lgtm validate [options]
```

**Behavior**:
1. Parse and validate all configuration
2. Test GitHub token authentication and permissions
3. Test Slack token authentication
4. Report validation results and exit

**Exit Codes**:
- `0`: All validations passed
- `1`: Configuration validation failed
- `2`: GitHub authentication failed
- `3`: Slack authentication failed

**Output**:
```
✓ Configuration valid
✓ GitHub token authenticated as: username
✓ GitHub token has required permissions
✓ Slack bot token authenticated
✓ Slack app token authenticated
✓ All validations passed
```

### `version`

**Description**: Display version information.

**Usage**:
```bash
lgtm version
```

**Output**:
```
lgtm version 1.0.0
Go version: go1.25
Platform: linux/amd64
```

## Configuration Priority

Configuration values are resolved in this order (highest to lowest priority):

1. Command-line flags
2. Environment variables
3. Default values

## Error Handling

### Invalid Configuration
```bash
$ lgtm --github-token=""
ERROR: GitHub token is required
Use --help for usage information
```

### Authentication Failure
```bash
$ lgtm --github-token="invalid"
ERROR: GitHub authentication failed: Bad credentials
Verify your GitHub token has the required permissions
```

### Pattern Compilation Error
```bash
$ lgtm --slack-pattern="[invalid"
ERROR: Invalid message pattern: error parsing regexp: missing closing ]
```

## Environment Variable Support

All configuration can be provided via environment variables:

```bash
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"
export SLACK_BOT_TOKEN="xoxb-xxxxxxxxxxxxxxxxxxxx"
export SLACK_APP_TOKEN="xapp-xxxxxxxxxxxxxxxxxxxx" 
export SLACK_CHANNEL_ID="C1234567890"
export SLACK_MESSAGE_PATTERN="(?i)lgtm|looks good to me"

lgtm  # Uses environment variables
```

## Signal Handling

- `SIGINT` (Ctrl+C): Graceful shutdown with cleanup
- `SIGTERM`: Graceful shutdown with cleanup
- `SIGUSR1`: Reload configuration (if supported in future versions)

## Logging Format

**Structured logging with consistent format**:
```
[TIMESTAMP] [LEVEL] [COMPONENT] MESSAGE [key=value ...]

Example:
2025-11-14T10:30:45Z INFO slack Message received channel=C1234567890 user=U9876543210 pattern_match=true
2025-11-14T10:30:46Z INFO github Approving PR owner=myorg repo=myrepo pr_number=123
2025-11-14T10:30:47Z INFO github PR approved review_id=987654321 pr_url=https://github.com/myorg/myrepo/pull/123
```

## Usage Examples

### Basic Usage
```bash
lgtm --github-token="$GITHUB_TOKEN" \
     --slack-bot-token="$SLACK_BOT_TOKEN" \
     --slack-app-token="$SLACK_APP_TOKEN" \
     --slack-channel-id="C1234567890" \
     --slack-pattern="(?i)lgtm|approve"
```

### Monitor All Channels
```bash
lgtm --github-token="$GITHUB_TOKEN" \
     --slack-bot-token="$SLACK_BOT_TOKEN" \
     --slack-app-token="$SLACK_APP_TOKEN"
     # No --slack-channel-id means monitor all channels
```

### With Default Repository
```bash
lgtm --github-token="$GITHUB_TOKEN" \
     --slack-bot-token="$SLACK_BOT_TOKEN" \
     --slack-app-token="$SLACK_APP_TOKEN" \
     --github-owner="myorg" \
     --github-repo="myrepo" \
     --slack-pattern="#\d+"  # Match #123 style PR references
```