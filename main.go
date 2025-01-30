package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"github.com/hypermodeinc/modus/sdk/go/pkg/http"
	"github.com/hypermodeinc/modus/sdk/go/pkg/models"
	"github.com/hypermodeinc/modus/sdk/go/pkg/models/openai"
)

type GitHubUser struct {
	Login     string `json:"login"`
	HTMLURL   string `json:"html_url"`
	Type      string `json:"type"`
	SiteAdmin bool   `json:"site_admin"`
}

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
	Reactions   GitHubReactions `json:"reactions"` 
}

type GitHubComment struct {
	User      GitHubUser `json:"user"` 
	Body      string     `json:"body"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

func PostDiscussionToRepo(repo string, title string, body string, token string) error {
    repoOwner := strings.Split(repo, "/")[0]
    repoName := strings.Split(repo, "/")[1]

    // Step 1: Get Repository ID
    query := `{
        "query": "query { repository(owner: \"%s\", name: \"%s\") { id } }"
    }`
    query = fmt.Sprintf(query, repoOwner, repoName)

    options := &http.RequestOptions{
        Method: "POST",
        Headers: map[string]string{
            "Authorization": "Bearer " + token,
            "Content-Type":  "application/json",
        },
        Body: json.RawMessage(query),
    }

    request := http.NewRequest("https://api.github.com/graphql", options)
    response, err := http.Fetch(request)
    if err != nil {
        return fmt.Errorf("error fetching repository ID: %w", err)
    }

    var repoResponse struct {
        Data struct {
            Repository struct {
                ID string `json:"id"`
            } `json:"repository"`
        } `json:"data"`
    }
    if err := json.Unmarshal(response.Body, &repoResponse); err != nil {
        return fmt.Errorf("error parsing repository ID response: %w", err)
    }
    
    repoID := repoResponse.Data.Repository.ID
    if repoID == "" {
        return fmt.Errorf("failed to retrieve repository ID")
    }

    // Step 2: Create Discussion
    discussionMutation := `{
        "query": "mutation { createDiscussion(input: {repositoryId: \"%s\", categoryId: \"%s\", title: \"%s\", body: \"%s\"}) { discussion { url } } }"
    }`
    
    // Replace with a valid Discussion Category ID (can be retrieved from the GitHub API)
    categoryID := "YOUR_DISCUSSION_CATEGORY_ID"

    discussionMutation = fmt.Sprintf(discussionMutation, repoID, categoryID, title, body)

    options.Body = json.RawMessage(discussionMutation)
    request = http.NewRequest("https://api.github.com/graphql", options)
    response, err = http.Fetch(request)
    if err != nil {
        return fmt.Errorf("error creating discussion: %w", err)
    }

    var discussionResponse struct {
        Data struct {
            CreateDiscussion struct {
                Discussion struct {
                    URL string `json:"url"`
                } `json:"discussion"`
            } `json:"createDiscussion"`
        } `json:"data"`
    }
    
    if err := json.Unmarshal(response.Body, &discussionResponse); err != nil {
        return fmt.Errorf("error parsing discussion response: %w", err)
    }

    fmt.Printf("Discussion created successfully: %s\n", discussionResponse.Data.CreateDiscussion.Discussion.URL)
    return nil
}



func GenerateKBArticle(issue *GitHubIssue, comments []GitHubComment) (string, error) {
	prompt := fmt.Sprintf(`
		Generate a detailed markdown article givin a concise summary of the following GitHub issue. Include the problem, the solution (if on exists), and any other relevant details. Please mention the users involved and any significant involvement in the issue.
		
		### Issue Details:
		- **Title**: %s
		- **Description**: %s
		- **State**: %s
		- **Created by**: %s
		- **Created at**: %s
		- **Labels**: %v
		- **Reactions**: 👍 (%d), 👎 (%d), ❤️ (%d), 🎉 (%d), 🚀 (%d), 👀 (%d)

		### Comments:
		%s

		Generate the output in markdown format.
	`,
		issue.Title, issue.Body, issue.State, issue.User.Login, issue.CreatedAt,
		getLabelNames(issue.Labels),
		issue.Reactions.PlusOne, issue.Reactions.MinusOne, issue.Reactions.Heart,
		issue.Reactions.Hooray, issue.Reactions.Rocket, issue.Reactions.Eyes,
		formatComments(comments),
	)

	model, err := models.GetModel[openai.ChatModel]("generate-article")
	if err != nil {
		return "", fmt.Errorf("error fetching model: %w", err)
	}

	input, err := model.CreateInput(
		openai.NewSystemMessage("You are an assistant that generates markdown documentation."),
		openai.NewUserMessage(prompt),
	)
	if err != nil {
		return "", fmt.Errorf("error creating model input: %w", err)
	}

	input.Temperature = 0.7

	output, err := model.Invoke(input)
	if err != nil {
		return "", fmt.Errorf("error invoking model: %w", err)
	}

	return strings.TrimSpace(output.Choices[0].Message.Content), nil
}

func getLabelNames(labels []GitHubLabel) string {
	names := []string{}
	for _, label := range labels {
		names = append(names, label.Name)
	}
	return strings.Join(names, ", ")
}

func formatComments(comments []GitHubComment) string {
	formatted := []string{}
	for _, comment := range comments {
		formatted = append(formatted, fmt.Sprintf("- **%s** (%s): %s", comment.User.Login, comment.CreatedAt, comment.Body))
	}
	return strings.Join(formatted, "\n")
}

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

func FetchIssueComments(commentsURL string) ([]GitHubComment, error) {
	options := &http.RequestOptions{
		Method: "GET",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	request := http.NewRequest(commentsURL, options)
	response, err := http.Fetch(request)
	if err != nil {
		return nil, fmt.Errorf("error fetching issue comments: %w", err)
	}

	if response.Status != 200 {
		fmt.Printf("Unexpected response status: %d\n", response.Status)
		fmt.Printf("Response body: %s\n", string(response.Body))
		return nil, fmt.Errorf("unexpected status code: %d", response.Status)
	}

	var comments []GitHubComment
	if err := json.Unmarshal(response.Body, &comments); err != nil {
		return nil, fmt.Errorf("error parsing comments: %w", err)
	}

	return comments, nil
}

func IssueClosedHandler(repo string, issueNumber int, token string) {
	issue, err := FetchIssueDetails(repo, issueNumber)
	if err != nil {
		fmt.Printf("Error fetching issue details: %v\n", err)
		return
	}

	comments, err := FetchIssueComments(issue.CommentsURL)
	if err != nil {
		fmt.Printf("Error fetching issue comments: %v\n", err)
		return
	}

	kbArticle, err := GenerateKBArticle(issue, comments)
	if err != nil {
		fmt.Printf("Error generating KB article: %v\n", err)
		return
	}

	// Output the KB article
	fmt.Printf(kbArticle)

	err = PostDiscussionToRepo(repo, issue.Title, kbArticle, token)
    if err != nil {
        fmt.Printf("Error creating GitHub discussion: %v\n", err)
        return
    }

    fmt.Println("KB article successfully posted as a comment.")
}
