package main

import (
	"context"
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

	var OverallDescribeCompletion string
	OverallDescribeCompletion += fmt.Sprintf("Pull request title: %s, body: %s\n\n", pr.GetTitle(), pr.GetBody())
	for _, file := range diff.Files {

		if file.Patch == nil {
			continue
		}

		prompt := fmt.Sprintf(oAIClient.PromptDescribeChanges, *file.Patch)

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
		OverallDescribeCompletion += fmt.Sprintf("File: %s \nDescription: %s \n\n", file.GetFilename(), completion)
	}

	overallCompletion, err := openAIClient.ChatCompletion(ctx, []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf(oAIClient.PromptOverallDescribe, OverallDescribeCompletion),
		},
	})
	if err != nil {
		return fmt.Errorf("error getting overall review: %w", err)
	}

	if opts.Test {
		fmt.Println(OverallDescribeCompletion)
		fmt.Println("=====================================")
		fmt.Println(overallCompletion)

		return nil
	}

	// Update the pull request description
	updatePr := &github.PullRequest{Body: github.String(overallCompletion)}
	if _, err = githubClient.UpdatePullRequest(ctx, opts.Owner, opts.Repo, opts.PRNumber, updatePr); err != nil {
		return fmt.Errorf("error updating pull request: %w", err)
	}

	return nil
}
