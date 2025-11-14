package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"
)

// Global log level variable
var logLevel string

// logDebug logs only if level is debug
func logDebug(format string, v ...interface{}) {
	if logLevel == "debug" {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// logInfo logs for info level and above
func logInfo(format string, v ...interface{}) {
	if logLevel == "debug" || logLevel == "info" {
		log.Printf("[INFO] "+format, v...)
	}
}

// logWarn logs for warn level and above
func logWarn(format string, v ...interface{}) {
	if logLevel == "debug" || logLevel == "info" || logLevel == "warn" {
		log.Printf("[WARN] "+format, v...)
	}
}

// logError logs for all levels
func logError(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

func main() {
	app := &cli.App{
		Name:  "lgtm",
		Usage: "Slack-to-GitHub bot that monitors Slack messages and approves GitHub pull requests",
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "Start the bot to monitor Slack messages and approve GitHub PRs",
				Action:  runCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "github-token",
						Usage:    "GitHub personal access token",
						EnvVars:  []string{"GITHUB_TOKEN"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "slack-bot-token",
						Usage:    "Slack bot user OAuth token",
						EnvVars:  []string{"SLACK_BOT_TOKEN"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "slack-app-token",
						Usage:    "Slack app-level token for Socket Mode",
						EnvVars:  []string{"SLACK_APP_TOKEN"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    "slack-channel-id",
						Usage:   "Specific channel ID to monitor (empty = all channels)",
						EnvVars: []string{"SLACK_CHANNEL_ID"},
					},
					&cli.StringFlag{
						Name:    "slack-pattern",
						Usage:   "Regex pattern for message matching",
						EnvVars: []string{"SLACK_MESSAGE_PATTERN"},
						Value:   ".*",
					},
					&cli.StringFlag{
						Name:    "github-owner",
						Usage:   "Default repository owner",
						EnvVars: []string{"GITHUB_OWNER"},
					},
					&cli.StringFlag{
						Name:    "github-repo",
						Usage:   "Default repository name",
						EnvVars: []string{"GITHUB_REPO"},
					},
					&cli.StringFlag{
						Name:    "log-level",
						Usage:   "Logging level (debug, info, warn, error)",
						EnvVars: []string{"LOG_LEVEL"},
						Value:   "info",
					},
				},
			},
			{
				Name:   "validate",
				Usage:  "Validate configuration and tokens without starting the bot",
				Action: validateCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "github-token",
						Usage:    "GitHub personal access token",
						EnvVars:  []string{"GITHUB_TOKEN"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "slack-bot-token",
						Usage:    "Slack bot user OAuth token",
						EnvVars:  []string{"SLACK_BOT_TOKEN"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "slack-app-token",
						Usage:    "Slack app-level token for Socket Mode",
						EnvVars:  []string{"SLACK_APP_TOKEN"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    "log-level",
						Usage:   "Logging level (debug, info, warn, error)",
						EnvVars: []string{"LOG_LEVEL"},
						Value:   "info",
					},
				},
			},
			{
				Name:   "version",
				Usage:  "Display version information",
				Action: versionCommand,
			},
		},
		Action: func(c *cli.Context) error {
			// Default action is to run
			return runCommand(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runCommand(c *cli.Context) error {
	fmt.Println("Starting LGTM bot...")
	
	// Parse configuration from CLI flags
	config, err := parseConfig(c)
	if err != nil {
		return err
	}
	
	// Set global log level
	logLevel = strings.ToLower(config.LogLevel)
	
	// Validate configuration
	if err := validateConfiguration(config); err != nil {
		return fmt.Errorf("%v\n\nTroubleshooting:\n- Ensure all required environment variables are set: GITHUB_TOKEN, SLACK_BOT_TOKEN, SLACK_APP_TOKEN\n- Check token formats: Slack bot token should start with 'xoxb-', app token with 'xapp-'\n- Verify regex pattern syntax if using custom SLACK_MESSAGE_PATTERN", err)
	}
	
	// Show configuration summary
	logInfo("Configuration loaded - Pattern: '%s', Channel: %s, Log Level: %s", 
		config.MessagePattern, 
		func() string { if config.SlackChannelID != "" { return config.SlackChannelID } else { return "all channels" } }(),
		config.LogLevel)
	
	// Create pattern matcher
	matcher, err := NewPatternMatcher(config.MessagePattern)
	if err != nil {
		return fmt.Errorf("failed to create pattern matcher: %v\n\nTroubleshooting:\n- Check your SLACK_MESSAGE_PATTERN environment variable for valid regex syntax\n- Test your pattern at https://regex101.com/\n- Use '.*' to match all messages (default)", err)
	}
	
	logDebug("Pattern matcher initialized with pattern: %s", config.MessagePattern)
	
	// Set up graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create GitHub client
	githubClient, err := NewGitHubClient(config)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %v\n\nTroubleshooting:\n- Verify your GITHUB_TOKEN environment variable is set and valid\n- Ensure the token has 'repo' scope for private repositories or 'public_repo' for public ones\n- Check GitHub token at https://github.com/settings/tokens", err)
	}
	
	// Validate GitHub permissions
	if err := githubClient.ValidatePermissions(ctx); err != nil {
		return fmt.Errorf("GitHub permission validation failed: %v\n\nTroubleshooting:\n- Ensure your GitHub token has the correct permissions\n- For private repos: token needs 'repo' scope\n- For public repos: token needs 'public_repo' scope\n- Verify the default repository exists and is accessible", err)
	}
	
	// Create Slack client
	slackClient, err := NewSlackClient(config, matcher, githubClient)
	if err != nil {
		return fmt.Errorf("failed to create Slack client: %v\n\nTroubleshooting:\n- Verify SLACK_BOT_TOKEN starts with 'xoxb-'\n- Verify SLACK_APP_TOKEN starts with 'xapp-'\n- Check that your Slack app has Socket Mode enabled\n- Ensure bot has been added to the target channel", err)
	}
	
	// Handle shutdown signals
	go handleShutdown(cancel)
	
	logInfo("Bot ready - listening for messages...")
	
	// Start Slack client (blocking)
	if err := slackClient.Start(ctx); err != nil && ctx.Err() == nil {
		return fmt.Errorf("Slack client error: %v\n\nTroubleshooting:\n- Check that Slack app has correct OAuth scopes (app_mentions:read, channels:history, chat:write)\n- Verify Socket Mode is enabled in Slack app settings\n- Ensure bot token and app token are both valid and active\n- Check Slack app event subscriptions are configured", err)
	}
	
	logInfo("Bot shutdown complete")
	return nil
}

// handleShutdown listens for shutdown signals and cancels the context
func handleShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-sigChan
	log.Printf("[MAIN] Received signal %v, shutting down...", sig)
	cancel()
}

func validateCommand(c *cli.Context) error {
	fmt.Println("Validating configuration...")
	
	// Parse configuration from CLI flags
	config, err := parseConfig(c)
	if err != nil {
		return err
	}
	
	// Validate configuration
	if err := validateConfiguration(config); err != nil {
		return err
	}
	
	fmt.Printf("✓ Configuration valid\n")
	return nil
}

// parseConfig creates a Configuration struct from CLI context
func parseConfig(c *cli.Context) (*Configuration, error) {
	config := &Configuration{
		GitHubToken:    c.String("github-token"),
		SlackBotToken:  c.String("slack-bot-token"),
		SlackAppToken:  c.String("slack-app-token"),
		SlackChannelID: c.String("slack-channel-id"),
		MessagePattern: c.String("slack-pattern"),
		DefaultOwner:   c.String("github-owner"),
		DefaultRepo:    c.String("github-repo"),
		LogLevel:       c.String("log-level"),
	}
	
	return config, nil
}

func versionCommand(c *cli.Context) error {
	fmt.Printf("lgtm version 1.0.0\n")
	fmt.Printf("Go version: %s\n", "go1.25")
	return nil
}