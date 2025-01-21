package main

import (
	"encoding/json"
	"fmt"
	"github.com/hypermodeinc/modus/sdk/go/pkg/http"
)

type GitHubIssue struct {
    Title string `json:"title"`
    Body  string `json:"body"`
    State string `json:"state"`
}

func FetchIssueDetails(repo string, issueNumber int) (*GitHubIssue, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%d", repo, issueNumber)

    options := &http.RequestOptions{
        Method: "GET",
        Headers: map[string]string{
            "Content-Type":  "application/json",
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



func SimpleHandler(repo string, issueNumber int) {
    issue, err := FetchIssueDetails(repo, issueNumber)
    if err != nil {
        fmt.Printf("Error fetching issue details: %v\n", err)
        return
    }
	fmt.Printf("Issue Details: %s\n", issue)

    fmt.Printf("Repository: %s\n", repo)
    fmt.Printf("Issue Number: %d\n", issueNumber)
    fmt.Printf("Title: %s\n", issue.Title)
    fmt.Printf("Body: %s\n", issue.Body)
    fmt.Printf("State: %s\n", issue.State)
}
