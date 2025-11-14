package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Configuration holds all runtime configuration for the bot
type Configuration struct {
	GitHubToken      string
	SlackBotToken    string
	SlackAppToken    string
	SlackChannelID   string
	MessagePattern   string
	DefaultOwner     string
	DefaultRepo      string
	LogLevel         string
}

// Custom error types
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("configuration error [%s]: %s", e.Field, e.Message)
}

type AuthenticationError struct {
	Service string
	Message string
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("%s authentication failed: %s", e.Service, e.Message)
}

type ProcessingError struct {
	Operation string
	Cause     error
}

func (e *ProcessingError) Error() string {
	return fmt.Sprintf("processing error [%s]: %v", e.Operation, e.Cause)
}

// validateConfiguration validates all configuration fields
func validateConfiguration(config *Configuration) error {
	// Validate required tokens
	if config.GitHubToken == "" {
		return &ConfigError{Field: "GitHubToken", Message: "GitHub token is required"}
	}
	
	if config.SlackBotToken == "" {
		return &ConfigError{Field: "SlackBotToken", Message: "Slack bot token is required"}
	}
	
	if config.SlackAppToken == "" {
		return &ConfigError{Field: "SlackAppToken", Message: "Slack app token is required"}
	}
	
	// Validate token formats
	if !strings.HasPrefix(config.SlackBotToken, "xoxb-") {
		return &ConfigError{Field: "SlackBotToken", Message: "Slack bot token must start with 'xoxb-'"}
	}
	
	if !strings.HasPrefix(config.SlackAppToken, "xapp-") {
		return &ConfigError{Field: "SlackAppToken", Message: "Slack app token must start with 'xapp-'"}
	}
	
	// Validate message pattern (regex)
	if config.MessagePattern != "" {
		_, err := regexp.Compile(config.MessagePattern)
		if err != nil {
			return &ConfigError{Field: "MessagePattern", Message: fmt.Sprintf("Invalid regex pattern: %v", err)}
		}
	}
	
	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	
	if !validLogLevels[strings.ToLower(config.LogLevel)] {
		return &ConfigError{Field: "LogLevel", Message: "Log level must be one of: debug, info, warn, error"}
	}
	
	return nil
}