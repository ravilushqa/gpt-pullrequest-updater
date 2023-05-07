package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/go-github/v51/github"
	"github.com/jessevdk/go-flags"
	"github.com/sashabaranov/go-openai"

	ghClient "github.com/ravilushqa/gpt-pullrequest-updater/github"
	oAIClient "github.com/ravilushqa/gpt-pullrequest-updater/openai"
)

var opts struct {
	GithubToken string `long:"gh-token" env:"GITHUB_TOKEN" description:"GitHub token" required:"true"`
	OpenAIToken string `long:"openai-token" env:"OPENAI_TOKEN" description:"OpenAI token" required:"true"`
	Owner       string `long:"owner" env:"OWNER" description:"GitHub owner" required:"true"`
	Repo        string `long:"repo" env:"REPO" description:"GitHub repo" required:"true"`
	PRNumber    int    `long:"pr-number" env:"PR_NUMBER" description:"Pull request number" required:"true"`
	Test        bool   `long:"test" env:"TEST" description:"Test mode"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if _, err := flags.Parse(&opts); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("Error parsing flags: %v \n", err)
		}
		os.Exit(0)
	}

	if err := run(ctx); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	openAIClient := oAIClient.NewClient(opts.OpenAIToken)
	githubClient := ghClient.NewClient(ctx, opts.GithubToken)

	pr, err := githubClient.GetPullRequest(ctx, opts.Owner, opts.Repo, opts.PRNumber)
	if err != nil {
		return fmt.Errorf("error getting pull request: %w", err)
	}

	diff, err := githubClient.CompareCommits(ctx, opts.Owner, opts.Repo, pr.GetBase().GetSHA(), pr.GetHead().GetSHA())
	if err != nil {
		return fmt.Errorf("error getting commits: %w", err)
	}

	var comments []*github.PullRequestComment

	for i, file := range diff.Files {
		patch := file.GetPatch()
		fmt.Printf("processing file: %s %d/%d\n", file.GetFilename(), i+1, len(diff.Files))
		if patch == "" || file.GetStatus() == "removed" || file.GetStatus() == "renamed" {
			continue
		}

		if len(patch) > 3000 {
			fmt.Println("Patch is too long, truncating")
			patch = fmt.Sprintf("%s...", patch[:3000])
		}
		completion, err := openAIClient.ChatCompletion(ctx, []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: oAIClient.PromptReview,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: patch,
			},
		})

		if err != nil {
			return fmt.Errorf("error getting completion: %w", err)
		}

		if opts.Test {
			fmt.Println("Completion:", completion)
		}

		review, err := extractJSON(completion)
		if err != nil {
			fmt.Println("Error extracting JSON:", err)
			continue
		}

		if review.Quality == Good {
			fmt.Println("Review is good")
			continue
		}
		for _, issue := range review.Issues {
			body := fmt.Sprintf("[%s] %s", issue.Type, issue.Description)
			comment := &github.PullRequestComment{
				CommitID: diff.Commits[len(diff.Commits)-1].SHA,
				Path:     file.Filename,
				Body:     &body,
				Position: &issue.Line,
			}
			comments = append(comments, comment)
		}

		if opts.Test {
			continue
		}

		for i, c := range comments {
			fmt.Printf("creating comment: %s %d/%d\n", *c.Path, i+1, len(comments))
			if _, err := githubClient.CreatePullRequestComment(ctx, opts.Owner, opts.Repo, opts.PRNumber, c); err != nil {
				return fmt.Errorf("error creating comment: %w", err)
			}
		}
	}
	return nil
}

type Review struct {
	Quality Quality `json:"quality"`
	Issues  []struct {
		Type        string `json:"type"`
		Line        int    `json:"line"`
		Description string `json:"description"`
	} `json:"issues"`
}

type Quality string

const (
	Good    Quality = "good"
	Bad     Quality = "bad"
	Neutral Quality = "neutral"
)

func extractJSON(input string) (*Review, error) {
	var jsonObj *Review

	// find the start and end positions of the JSON object
	start := 0
	end := len(input)
	for i, c := range input {
		if c == '{' {
			start = i
			break
		}
		if i == len(input)-1 {
			return nil, errors.New("invalid JSON object")
		}
	}
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] == '}' {
			end = i + 1
			break
		}

		if i == 0 {
			return nil, errors.New("invalid JSON object")
		}
	}

	// extract the JSON object from the input
	jsonStr := input[start:end]
	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return nil, errors.New("invalid JSON object")
	}

	return jsonObj, nil
}
