package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/go-github/v51/github"
	"github.com/jessevdk/go-flags"

	"github.com/ravilushqa/gpt-pullrequest-updater/description"
	ghClient "github.com/ravilushqa/gpt-pullrequest-updater/github"
	"github.com/ravilushqa/gpt-pullrequest-updater/jira"
	oAIClient "github.com/ravilushqa/gpt-pullrequest-updater/openai"
	"github.com/ravilushqa/gpt-pullrequest-updater/shortcut"
)

var opts struct {
	GithubToken     string `long:"gh-token" env:"GITHUB_TOKEN" description:"GitHub token" required:"true"`
	OpenAIToken     string `long:"openai-token" env:"OPENAI_TOKEN" description:"OpenAI token" required:"true"`
	Owner           string `long:"owner" env:"OWNER" description:"GitHub owner" required:"true"`
	Repo            string `long:"repo" env:"REPO" description:"GitHub repo" required:"true"`
	PRNumber        int    `long:"pr-number" env:"PR_NUMBER" description:"Pull request number" required:"true"`
	OpenAIModel     string `long:"openai-model" env:"OPENAI_MODEL" description:"OpenAI model" default:"gpt-3.5-turbo"`
	Test            bool   `long:"test" env:"TEST" description:"Test mode"`
	JiraURL         string `long:"jira-url" env:"JIRA_URL" description:"Jira URL. Example: https://jira.atlassian.com"`
	ShortcutBaseURL string `long:"shortcut-url" env:"SHORTCUT_URL" description:"Shortcut URL. Example: https://app.shortcut.com/foo"`
}

var descriptionInfo description.Info

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

	descriptionInfo.Completion, err = description.GenerateCompletion(ctx, openAIClient, diff, pr)
	if err != nil {
		return fmt.Errorf("error generating completion: %w", err)
	}

	if opts.JiraURL != "" {
		fmt.Println("Adding Jira ticket")
		id, err := jira.ExtractJiraTicketID(*pr.Title)
		if err != nil {
			fmt.Printf("Error extracting Jira ticket ID: %v \n", err)
		} else {
			descriptionInfo.JiraInfo = fmt.Sprintf("### JIRA ticket: [%s](%s)", id, jira.GenerateJiraTicketURL(opts.JiraURL, id))
		}
	}

	if opts.ShortcutBaseURL != "" {
		descriptionInfo.ShortcutInfo = buildShortcutContent(opts.ShortcutBaseURL, pr)
	}

	if opts.Test {
		return nil
	}

	// Update the pull request description
	fmt.Println("Updating pull request")
	updatedPr := description.BuildUpdatedPullRequest(*pr.Body, descriptionInfo)
	if _, err = githubClient.UpdatePullRequest(ctx, opts.Owner, opts.Repo, opts.PRNumber, updatedPr); err != nil {
		return fmt.Errorf("error updating pull request: %w", err)
	}

	return nil
}

func buildShortcutContent(shortcutBaseURL string, pr *github.PullRequest) string {
	fmt.Println("Adding Shortcut ticket")

	id, err := shortcut.ExtractShortcutStoryID(*pr.Title)

	if err != nil {
		// Extracting from the branch name
		id, err = shortcut.ExtractShortcutStoryID(*pr.Head.Ref)
	}

	if err != nil {
		fmt.Printf("There is no Shortcut story ID: %v \n", err)
		return ""
	}

	return fmt.Sprintf("### Shortcut story: [%s](%s)", id, shortcut.GenerateShortcutStoryURL(shortcutBaseURL, id))
}
