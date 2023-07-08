package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"

	ghClient "github.com/ravilushqa/gpt-pullrequest-updater/github"
	oAIClient "github.com/ravilushqa/gpt-pullrequest-updater/openai"
	"github.com/ravilushqa/gpt-pullrequest-updater/review"
)

var opts struct {
	GithubToken string `long:"gh-token" env:"GITHUB_TOKEN" description:"GitHub token" required:"true"`
	OpenAIToken string `long:"openai-token" env:"OPENAI_TOKEN" description:"OpenAI token" required:"true"`
	Owner       string `long:"owner" env:"OWNER" description:"GitHub owner" required:"true"`
	Repo        string `long:"repo" env:"REPO" description:"GitHub repo" required:"true"`
	PRNumber    int    `long:"pr-number" env:"PR_NUMBER" description:"Pull request number" required:"true"`
	OpenAIModel string `long:"openai-model" env:"OPENAI_MODEL" description:"OpenAI model" default:"gpt-3.5-turbo"`
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
	openAIClient := oAIClient.NewClient(opts.OpenAIToken, opts.OpenAIModel)
	githubClient := ghClient.NewClient(ctx, opts.GithubToken)

	pr, err := githubClient.GetPullRequest(ctx, opts.Owner, opts.Repo, opts.PRNumber)
	if err != nil {
		return fmt.Errorf("error getting pull request: %w", err)
	}

	diff, err := githubClient.CompareCommits(ctx, opts.Owner, opts.Repo, pr.GetBase().GetSHA(), pr.GetHead().GetSHA())
	if err != nil {
		return fmt.Errorf("error getting commits: %w", err)
	}

	comments, err := review.GenerateCommentsFromDiff(ctx, openAIClient, diff)
	if err != nil {
		return err
	}

	if opts.Test {
		fmt.Printf("Comments: %v \n", comments)
		return nil
	}

	err = review.PushComments(ctx, githubClient, opts.Owner, opts.Repo, opts.PRNumber, comments)
	if err != nil {
		return fmt.Errorf("error creating comments: %w", err)
	}

	return nil
}
