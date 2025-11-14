# LGTM Bot

Slack-to-GitHub bot that listens for messages and approves pull requests.

## Install

```bash
go install github.com/alileza/lgtm@latest
```

## Setup

### GitHub Token
Create a personal access token with `Pull requests: Write` permission at https://github.com/settings/tokens

### Slack App
1. Create app at https://api.slack.com/apps
2. Add scopes: `channels:read`, `chat:write`, `app_mentions:read`
3. Enable Socket Mode and generate app-level token

## Run

```bash
export GITHUB_TOKEN="ghp_..."
export SLACK_BOT_TOKEN="xoxb-..."
export SLACK_APP_TOKEN="xapp-..."
export SLACK_MESSAGE_PATTERN="lgtm"

lgtm run
```

### Options

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--github-token` | `GITHUB_TOKEN` | | GitHub personal access token |
| `--slack-bot-token` | `SLACK_BOT_TOKEN` | | Slack bot user OAuth token |
| `--slack-app-token` | `SLACK_APP_TOKEN` | | Slack app-level token |
| `--slack-channel-id` | `SLACK_CHANNEL_ID` | all | Specific channel to monitor |
| `--slack-pattern` | `SLACK_MESSAGE_PATTERN` | `.*` | Regex pattern to match |
| `--github-owner` | `GITHUB_OWNER` | | Default repo owner |
| `--github-repo` | `GITHUB_REPO` | | Default repo name |
| `--log-level` | `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |

## Usage

Bot watches for messages matching the pattern and approves any GitHub PRs found in the message. Reacts with 👀 while processing, ✅ on success, ❌ on failure.