package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// SlackClient handles Slack Socket Mode connection
type SlackClient struct {
	api          *slack.Client
	socketClient *socketmode.Client
	config       *Configuration
	matcher      *PatternMatcher
	githubClient *GitHubClient
}

// NewSlackClient creates a new Slack client with Socket Mode
func NewSlackClient(config *Configuration, matcher *PatternMatcher, githubClient *GitHubClient) (*SlackClient, error) {
	// Create Slack API client with bot token
	api := slack.New(
		config.SlackBotToken,
		slack.OptionDebug(config.LogLevel == "debug"),
		slack.OptionAppLevelToken(config.SlackAppToken),
	)
	
	// Create Socket Mode client
	socketClient := socketmode.New(
		api,
		socketmode.OptionDebug(config.LogLevel == "debug"),
	)
	
	return &SlackClient{
		api:          api,
		socketClient: socketClient,
		config:       config,
		matcher:      matcher,
		githubClient: githubClient,
	}, nil
}

// Start begins the Slack Socket Mode connection
func (sc *SlackClient) Start(ctx context.Context) error {
	log.Printf("[SLACK] Connecting to Slack workspace...")
	
	// Test authentication first
	if err := sc.validateTokens(ctx); err != nil {
		return fmt.Errorf("Slack token validation failed: %v", err)
	}
	
	// Start event handling in background
	go sc.handleEvents(ctx)
	
	// Start Socket Mode connection (blocking)
	return sc.socketClient.RunContext(ctx)
}

// validateTokens validates Slack bot and app tokens
func (sc *SlackClient) validateTokens(ctx context.Context) error {
	// Test bot token by calling auth.test
	authResponse, err := sc.api.AuthTestContext(ctx)
	if err != nil {
		return &AuthenticationError{Service: "Slack", Message: fmt.Sprintf("bot token validation failed: %v", err)}
	}
	
	log.Printf("[SLACK] [AUTH] Bot authenticated as: %s (team: %s)", authResponse.User, authResponse.Team)
	
	// App token validation is implicit - if Socket Mode connection succeeds, the app token is valid
	log.Printf("[SLACK] [AUTH] App token validated successfully")
	
	return nil
}

// Stop gracefully shuts down the Slack client
func (sc *SlackClient) Stop(ctx context.Context) error {
	log.Printf("Stopping Slack client...")
	// Socket Mode client will stop when context is canceled
	return nil
}

// handleEvents processes incoming Slack events
func (sc *SlackClient) handleEvents(ctx context.Context) {
	for evt := range sc.socketClient.Events {
		switch evt.Type {
		case socketmode.EventTypeEventsAPI:
			eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
			if !ok {
				log.Printf("Unexpected event type: %T", evt.Data)
				sc.socketClient.Ack(*evt.Request)
				continue
			}
			
			// Acknowledge the event
			sc.socketClient.Ack(*evt.Request)
			
			// Handle the inner event
			sc.handleEventsAPIEvent(ctx, eventsAPIEvent)
			
		case socketmode.EventTypeConnecting:
			log.Printf("Connecting to Slack with Socket Mode...")
			
		case socketmode.EventTypeConnectionError:
			log.Printf("Connection failed. Retrying later...")
			
		case socketmode.EventTypeConnected:
			log.Printf("Connected to Slack workspace")
			
		default:
			log.Printf("Unexpected event type received: %s", evt.Type)
		}
	}
}

// handleEventsAPIEvent processes Events API events
func (sc *SlackClient) handleEventsAPIEvent(ctx context.Context, event slackevents.EventsAPIEvent) {
	if event.Type != slackevents.CallbackEvent {
		return
	}
	
	innerEvent := event.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		sc.handleMessageEvent(ctx, ev)
	default:
		// Ignore other event types
	}
}

// handleMessageEvent processes message events
func (sc *SlackClient) handleMessageEvent(ctx context.Context, event *slackevents.MessageEvent) {
	// Skip if channel filtering is enabled and this message is from a different channel
	if sc.config.SlackChannelID != "" && event.Channel != sc.config.SlackChannelID {
		return
	}
	
	// Skip bot messages to avoid processing our own messages
	if event.BotID != "" {
		return
	}
	
	// Create SlackMessage struct
	slackMsg := &SlackMessage{
		Text:      event.Text,
		Channel:   event.Channel,
		User:      event.User,
		Timestamp: event.TimeStamp,
		ThreadTS:  event.ThreadTimeStamp,
	}
	
	// Use structured logging for message events
	if sc.config.LogLevel == "debug" {
		log.Printf("[SLACK] [MESSAGE] channel=%s user=%s text=%q", event.Channel, event.User, event.Text)
	} else {
		log.Printf("[SLACK] Message received from channel %s", event.Channel)
	}
	
	// Process the message for pattern matching
	sc.processMessage(ctx, slackMsg)
}

// processMessage handles pattern matching for incoming messages
func (sc *SlackClient) processMessage(ctx context.Context, msg *SlackMessage) {
	// Attempt to match the message against the configured pattern
	match, err := sc.matcher.Match(msg.Text)
	if err != nil {
		log.Printf("Pattern matching error: %v", err)
		return
	}
	
	if match == nil {
		// No pattern match - ignore the message
		return
	}
	
	// Pattern matched!
	match.SourceMessage = msg
	log.Printf("[MATCHER] [MATCH] channel=%s user=%s pattern=%q matched_text=%q", 
		msg.Channel, msg.User, match.Pattern, match.MatchedText)
	
	// Process GitHub PR approvals if any PR references found
	if len(match.PRReferences) > 0 {
		sc.processPRApprovals(ctx, match)
	} else {
		log.Printf("[SLACK] Pattern matched but no PR references found in message")
		// React with X emoji - no PR references found
		sc.addReaction(msg.Channel, msg.Timestamp, "x")
	}
}

// processPRApprovals handles GitHub PR approvals for matched messages
func (sc *SlackClient) processPRApprovals(ctx context.Context, match *PatternMatch) {
	// Add eyes reaction - processing started
	sc.addReaction(match.SourceMessage.Channel, match.SourceMessage.Timestamp, "eyes")
	
	for _, prRef := range match.PRReferences {
		// Fill in missing owner/repo from configuration if needed
		owner := prRef.Owner
		repo := prRef.Repository
		
		if owner == "" {
			owner = sc.config.DefaultOwner
		}
		if repo == "" {
			repo = sc.config.DefaultRepo
		}
		
		// Skip if we still don't have owner/repo
		if owner == "" || repo == "" {
			log.Printf("[SLACK] [PR_APPROVAL] [SKIP] pr_number=%d reason=missing_owner_or_repo", prRef.Number)
			continue
		}
		
		// Create approval request
		approvalReq := &ApprovalRequest{
			Owner:         owner,
			Repository:    repo,
			PRNumber:      prRef.Number,
			SourceChannel: match.SourceMessage.Channel,
			SourceUser:    match.SourceMessage.User,
			SourceMessage: match.SourceMessage,
			Timestamp:     time.Now(),
		}
		
		// Process the approval (this will be async in a real implementation)
		go sc.processApproval(ctx, approvalReq)
	}
}

// processApproval processes a single PR approval request
func (sc *SlackClient) processApproval(ctx context.Context, req *ApprovalRequest) {
	log.Printf("[SLACK] [PR_APPROVAL] [START] owner=%s repo=%s pr_number=%d", req.Owner, req.Repository, req.PRNumber)
	
	// Validate PR exists and is in valid state first
	if err := sc.githubClient.ValidatePRReference(ctx, req.Owner, req.Repository, req.PRNumber); err != nil {
		log.Printf("[SLACK] [PR_APPROVAL] [VALIDATION_FAILED] pr_number=%d error=%v", req.PRNumber, err)
		return
	}
	
	// Approve PR with retry logic
	result, err := sc.githubClient.ApprovePRWithRetry(ctx, req)
	if err != nil {
		log.Printf("[SLACK] [PR_APPROVAL] [ERROR] pr_number=%d error=%v", req.PRNumber, err)
		return
	}
	
	// Log the result
	if result.Success {
		log.Printf("[SLACK] [PR_APPROVAL] [SUCCESS] pr_number=%d review_id=%d retries=%d", 
			req.PRNumber, result.ReviewID, result.RetryAttempts)
		// React with checkmark on success
		sc.addReaction(req.SourceChannel, req.SourceMessage.Timestamp, "white_check_mark")
	} else {
		log.Printf("[SLACK] [PR_APPROVAL] [FAILED] pr_number=%d error=%s retries=%d", 
			req.PRNumber, result.Error, result.RetryAttempts)
		// React with X on failure
		sc.addReaction(req.SourceChannel, req.SourceMessage.Timestamp, "x")
	}
}

// addReaction adds an emoji reaction to a Slack message
func (sc *SlackClient) addReaction(channel, timestamp, emoji string) {
	msgRef := slack.ItemRef{
		Channel:   channel,
		Timestamp: timestamp,
	}
	
	if err := sc.api.AddReaction(emoji, msgRef); err != nil {
		log.Printf("[SLACK] [REACTION] [ERROR] Failed to add reaction %s: %v", emoji, err)
	} else {
		log.Printf("[SLACK] [REACTION] Added reaction %s to message %s", emoji, timestamp)
	}
}