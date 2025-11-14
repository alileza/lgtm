package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// PatternMatcher handles message pattern matching
type PatternMatcher struct {
	pattern *regexp.Regexp
}

// NewPatternMatcher creates a new pattern matcher with compiled regex
func NewPatternMatcher(pattern string) (*PatternMatcher, error) {
	if pattern == "" {
		pattern = ".*" // Match all messages by default
	}
	
	compiledPattern, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	
	return &PatternMatcher{
		pattern: compiledPattern,
	}, nil
}

// PatternMatch represents a successful pattern match
type PatternMatch struct {
	Pattern       string
	MatchedText   string
	PRReferences  []PRReference
	SourceMessage *SlackMessage
}

// PRReference represents a GitHub pull request reference
type PRReference struct {
	Owner      string
	Repository string
	Number     int
	URL        string
}

// SlackMessage represents a Slack message
type SlackMessage struct {
	Text      string
	Channel   string
	User      string
	Timestamp string
	ThreadTS  string
}

// Match tests if a message matches the configured pattern
func (pm *PatternMatcher) Match(message string) (*PatternMatch, error) {
	if !pm.pattern.MatchString(message) {
		return nil, nil // No match
	}
	
	// Find the matched substring
	matchedText := pm.pattern.FindString(message)
	
	// Create pattern match and extract PR references
	patternMatch := &PatternMatch{
		Pattern:     pm.pattern.String(),
		MatchedText: matchedText,
	}
	
	// Extract PR references from the entire message (not just matched text)
	prRefs, err := pm.ExtractPRReferences(message)
	if err != nil {
		return nil, err
	}
	patternMatch.PRReferences = prRefs
	
	return patternMatch, nil
}

// ExtractPRReferences finds GitHub PR references in text
func (pm *PatternMatcher) ExtractPRReferences(text string) ([]PRReference, error) {
	var references []PRReference
	
	// GitHub PR URL pattern: https://github.com/owner/repo/pull/123 (matches your bash script)
	prURLPattern := regexp.MustCompile(`https?://github\.com/([^[:space:]/]+)/([^[:space:]/]+)/pull/([0-9]+)`)
	
	// Simple PR number pattern: #123, PR-456, PR #123
	prNumberPattern := regexp.MustCompile(`(?:#|PR-?)\s*(\d+)`)
	
	// Extract full URLs first
	urlMatches := prURLPattern.FindAllStringSubmatch(text, -1)
	for _, match := range urlMatches {
		if len(match) == 4 {
			number, err := strconv.Atoi(match[3])
			if err != nil {
				continue
			}
			
			// Validate the URL format
			if err := validateGitHubURL(match[0]); err != nil {
				continue
			}
			
			references = append(references, PRReference{
				Owner:      match[1],
				Repository: match[2],
				Number:     number,
				URL:        match[0],
			})
		}
	}
	
	// Extract simple PR numbers (these will need default owner/repo)
	numberMatches := prNumberPattern.FindAllStringSubmatch(text, -1)
	for _, match := range numberMatches {
		if len(match) == 2 {
			number, err := strconv.Atoi(match[1])
			if err != nil {
				continue
			}
			
			// Only add if we don't already have this PR from a URL
			alreadyExists := false
			for _, existing := range references {
				if existing.Number == number {
					alreadyExists = true
					break
				}
			}
			
			if !alreadyExists {
				references = append(references, PRReference{
					Number: number,
					// Owner and Repository will need to be filled from config
				})
			}
		}
	}
	
	return references, nil
}

// validateGitHubURL validates that a URL is a proper GitHub URL
func validateGitHubURL(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTPS URLs are supported")
	}
	
	if !strings.EqualFold(parsedURL.Host, "github.com") {
		return err
	}
	
	return nil
}