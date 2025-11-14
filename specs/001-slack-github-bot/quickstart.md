# Quickstart Guide: LGTM Bot

**Date**: 2025-11-14  
**Purpose**: Get the Slack-to-GitHub bot running quickly

## Prerequisites

- Go 1.25 or later
- GitHub personal access token with repository permissions
- Slack bot tokens (bot user OAuth token and app-level token)

## 1. Setup GitHub Token

### Create Personal Access Token

1. Go to GitHub Settings > Developer settings > Personal access tokens > Fine-grained tokens
2. Click "Generate new token"
3. Select repository access (specific repos or all repos)
4. Grant permissions:
   - **Contents**: Read and Write
   - **Pull requests**: Write  
   - **Issues**: Read
5. Copy the generated token (starts with `ghp_`)

### Verify Permissions

```bash
# Test token with curl
curl -H "Authorization: token YOUR_GITHUB_TOKEN" \
     https://api.github.com/user

# Should return your user information
```

## 2. Setup Slack Bot

### Create Slack App

1. Go to [Slack API Dashboard](https://api.slack.com/apps)
2. Click "Create New App" > "From scratch"
3. Enter app name (e.g., "LGTM Bot") and select workspace
4. Go to "OAuth & Permissions"
5. Add Bot Token Scopes:
   - `channels:read` - View basic channel info
   - `chat:write` - Send messages (for logging)
   - `app_mentions:read` - Listen for app mentions

### Enable Socket Mode

1. Go to "Socket Mode" in your app settings
2. Toggle "Enable Socket Mode"  
3. Generate App-Level Token with `connections:write` scope
4. Copy the app-level token (starts with `xapp-`)

### Install Bot to Workspace

1. Go to "OAuth & Permissions"
2. Click "Install to Workspace"
3. Authorize the app
4. Copy the Bot User OAuth Token (starts with `xoxb-`)

## 3. Quick Start

### Get Channel ID

```bash
# Method 1: From Slack URL
# https://workspace.slack.com/archives/C1234567890
# Channel ID is: C1234567890

# Method 2: Using Slack API (optional)
curl -H "Authorization: Bearer xoxb-YOUR-BOT-TOKEN" \
     https://slack.com/api/conversations.list
```

### Run the Bot

```bash
# Clone and build (when repository exists)
git clone https://github.com/yourusername/lgtm
cd lgtm
go build -o lgtm ./cmd/lgtm

# Set environment variables
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"
export SLACK_BOT_TOKEN="xoxb-xxxxxxxxxxxxxxxxxxxx"
export SLACK_APP_TOKEN="xapp-xxxxxxxxxxxxxxxxxxxx"
export SLACK_CHANNEL_ID="C1234567890"

# Start the bot
./lgtm
```

### Expected Output

```
INFO: Starting LGTM bot...
INFO: Authenticated as GitHub user: yourusername
INFO: Connected to Slack workspace: your-workspace
INFO: Monitoring channel: #general (C1234567890)
INFO: Pattern: .* (all messages)
INFO: Bot ready - listening for messages...
```

## 4. Test the Bot

### Send Test Message

In your Slack channel, post a message with a GitHub PR:
```
This looks good to me! https://github.com/owner/repo/pull/123
```

### Expected Bot Behavior

```
INFO: Message matched pattern in #general by @youruser
INFO: Extracted PR: owner/repo#123
INFO: Approved PR owner/repo#123 (Review ID: 987654321)
```

## 5. Common Patterns

### LGTM Pattern
```bash
# Match "LGTM" or "looks good to me" (case insensitive)
export SLACK_MESSAGE_PATTERN="(?i)lgtm|looks good to me"
./lgtm
```

### PR Number Pattern
```bash
# Match #123 style PR references  
export SLACK_MESSAGE_PATTERN="#\d+"
export GITHUB_OWNER="yourorg"
export GITHUB_REPO="yourrepo"  
./lgtm
```

### Specific Channel Only
```bash
# Monitor only #code-review channel
export SLACK_CHANNEL_ID="C7890123456"
export SLACK_MESSAGE_PATTERN="(?i)approve|lgtm"
./lgtm
```

## 6. Validation and Troubleshooting

### Validate Configuration
```bash
./lgtm validate
```

Expected output if successful:
```
✓ Configuration valid
✓ GitHub token authenticated as: yourusername  
✓ GitHub token has required permissions
✓ Slack bot token authenticated
✓ Slack app token authenticated
✓ All validations passed
```

### Common Issues

#### Invalid GitHub Token
```
ERROR: GitHub authentication failed: Bad credentials
```
**Solution**: Verify your `GITHUB_TOKEN` is correct and not expired.

#### Insufficient GitHub Permissions
```
ERROR: GitHub token lacks required permissions for PR approval
```
**Solution**: Ensure your token has `Pull requests: Write` permission.

#### Invalid Slack Tokens
```
ERROR: Slack authentication failed: invalid_auth
```
**Solution**: Verify `SLACK_BOT_TOKEN` and `SLACK_APP_TOKEN` are correct.

#### Bot Not Receiving Messages
```
INFO: Bot ready - listening for messages...
(no messages appear)
```
**Solutions**:
- Verify the bot is added to the channel
- Check `SLACK_CHANNEL_ID` is correct
- Ensure Socket Mode is enabled in Slack app settings

#### Pattern Not Matching
```
INFO: Message received but no pattern match
```
**Solutions**:
- Test your regex pattern with an online validator
- Use `.*` to match all messages for testing
- Check message content doesn't have unexpected characters

### Debug Mode

```bash
export LOG_LEVEL="debug"
./lgtm
```

This will show detailed logs for troubleshooting.

## 7. Production Deployment

### Environment Variables File

```bash
# Create .env file (never commit to git)
cat > .env << EOF
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx
SLACK_BOT_TOKEN=xoxb-xxxxxxxxxxxxxxxxxxxx
SLACK_APP_TOKEN=xapp-xxxxxxxxxxxxxxxxxxxx
SLACK_CHANNEL_ID=C1234567890
SLACK_MESSAGE_PATTERN=(?i)lgtm|looks good to me
GITHUB_OWNER=yourorg
GITHUB_REPO=yourrepo
LOG_LEVEL=info
EOF

# Load and run
set -a && source .env && set +a
./lgtm
```

### Systemd Service (Linux)

```ini
# /etc/systemd/system/lgtm-bot.service
[Unit]
Description=LGTM Slack to GitHub Bot
After=network.target

[Service]
Type=simple
User=lgtm
EnvironmentFile=/opt/lgtm/.env
ExecStart=/opt/lgtm/lgtm
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable lgtm-bot
sudo systemctl start lgtm-bot
```

### Docker Deployment

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o lgtm ./cmd/lgtm

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/lgtm .
CMD ["./lgtm"]
```

```bash
# Build and run
docker build -t lgtm-bot .
docker run --env-file .env lgtm-bot
```

You're now ready to use the LGTM bot! The bot will automatically approve GitHub pull requests when it detects matching patterns in your Slack messages.