package main

import (
	"encoding/json"
	"fmt"
	"github.com/hypermodeinc/modus/sdk/go/pkg/http"
)

// GitHubUser represents a GitHub user or bot
type GitHubUser struct {
	Login     string `json:"login"`
	HTMLURL   string `json:"html_url"`
	Type      string `json:"type"`
	SiteAdmin bool   `json:"site_admin"`
}

// GitHubLabel represents a GitHub issue label
type GitHubLabel struct {
	Name string `json:"name"`
}

type GitHubReactions struct {
	URL        string `json:"url"`
	TotalCount int    `json:"total_count"`
	PlusOne    int    `json:"+1"`
	MinusOne   int    `json:"-1"`
	Laugh      int    `json:"laugh"`
	Hooray     int    `json:"hooray"`
	Confused   int    `json:"confused"`
	Heart      int    `json:"heart"`
	Rocket     int    `json:"rocket"`
	Eyes       int    `json:"eyes"`
}

type GitHubIssue struct {
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	State       string         `json:"state"`
	Number      int            `json:"number"`
	HTMLURL     string         `json:"html_url"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
	ClosedAt    string         `json:"closed_at"`
	User        GitHubUser     `json:"user"`
	Labels      []GitHubLabel  `json:"labels"`
	Comments    int            `json:"comments"`
	CommentsURL string         `json:"comments_url"`
	Reactions   GitHubReactions `json:"reactions"` // Updated to use GitHubReactions struct
}

// FetchIssueDetails fetches the details of a GitHub issue
func FetchIssueDetails(repo string, issueNumber int) (*GitHubIssue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%d", repo, issueNumber)

	options := &http.RequestOptions{
		Method: "GET",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	request := http.NewRequest(url, options)
	response, err := http.Fetch(request)
	if err != nil {
		return nil, fmt.Errorf("error fetching issue details: %w", err)
	}

	if response.Status != 200 {
		fmt.Printf("Unexpected response status: %d\n", response.Status)
		fmt.Printf("Response body: %s\n", string(response.Body))
		return nil, fmt.Errorf("unexpected status code: %d", response.Status)
	}

	var issue GitHubIssue
	if err := json.Unmarshal(response.Body, &issue); err != nil {
		return nil, fmt.Errorf("error parsing issue details: %w", err)
	}

	return &issue, nil
}

// IssueClosedHandler handles the closure of an issue
func IssueClosedHandler(repo string, issueNumber int) {
	issue, err := FetchIssueDetails(repo, issueNumber)
	if err != nil {
		fmt.Printf("Error fetching issue details: %v\n", err)
		return
	}

	// Print the full issue details
	fmt.Printf("Issue Details:\n%+v\n", issue)
}
