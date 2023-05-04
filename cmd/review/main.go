package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
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

	var OverallReviewCompletion string
	for _, file := range diff.Files {
		if file.Patch == nil || file.GetStatus() == "removed" || file.GetStatus() == "renamed" {
			continue
		}

		prompt := fmt.Sprintf(oAIClient.PromptReview, *file.Patch)

		if len(prompt) > 4096 {
			prompt = fmt.Sprintf("%s...", prompt[:4093])
		}

		completion, err := openAIClient.ChatCompletion(ctx, []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		})
		if err != nil {
			return fmt.Errorf("error getting review: %w", err)
		}
		OverallReviewCompletion += fmt.Sprintf("File: %s \nReview: %s \n\n", file.GetFilename(), completion)

		position := len(strings.Split(*file.Patch, "\n")) - 1

		comment := &github.PullRequestComment{
			CommitID: diff.Commits[len(diff.Commits)-1].SHA,
			Path:     file.Filename,
			Body:     &completion,
			Position: &position,
		}

		if opts.Test {
			continue
		}

		if _, err := githubClient.CreatePullRequestComment(ctx, opts.Owner, opts.Repo, opts.PRNumber, comment); err != nil {
			return fmt.Errorf("error creating comment: %w", err)
		}
	}

	overallCompletion, err := openAIClient.ChatCompletion(ctx, []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf(oAIClient.PromptOverallReview, OverallReviewCompletion),
		},
	})
	if err != nil {
		return fmt.Errorf("error getting overall review: %w", err)
	}

	if opts.Test {
		fmt.Println(OverallReviewCompletion)
		fmt.Println("=====================================")
		fmt.Println(overallCompletion)

		return nil
	}

	comment := &github.PullRequestReviewRequest{Body: &overallCompletion}
	if _, err = githubClient.CreateReview(ctx, opts.Owner, opts.Repo, opts.PRNumber, comment); err != nil {
		return fmt.Errorf("error creating comment: %w", err)
	}

	return nil
}
