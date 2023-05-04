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
	"github.com/ravilushqa/gpt-pullrequest-updater/jira"
	oAIClient "github.com/ravilushqa/gpt-pullrequest-updater/openai"
)

var opts struct {
	GithubToken string `long:"gh-token" env:"GITHUB_TOKEN" description:"GitHub token" required:"true"`
	OpenAIToken string `long:"openai-token" env:"OPENAI_TOKEN" description:"OpenAI token" required:"true"`
	Owner       string `long:"owner" env:"OWNER" description:"GitHub owner" required:"true"`
	Repo        string `long:"repo" env:"REPO" description:"GitHub repo" required:"true"`
	PRNumber    int    `long:"pr-number" env:"PR_NUMBER" description:"Pull request number" required:"true"`
	Test        bool   `long:"test" env:"TEST" description:"Test mode"`
	JiraURL     string `long:"jira-url" env:"JIRA_URL" description:"Jira URL"`
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

	if opts.Test {
		fmt.Println("Test mode")
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

	var sumDiffs int
	for _, file := range diff.Files {
		sumDiffs += len(*file.Patch)
	}

	var completion string
	if sumDiffs < 4000 {
		completion, err = genCompletionOnce(ctx, openAIClient, diff)
		if err != nil {
			return fmt.Errorf("error generating completition once: %w", err)
		}
	} else {
		completion, err = genCompletionPerFile(ctx, openAIClient, diff, pr)
		if err != nil {
			return fmt.Errorf("error generating completition twice: %w", err)
		}
	}

	if opts.JiraURL != "" {
		fmt.Println("Adding Jira ticket")
		id, err := jira.ExtractJiraTicketID(*pr.Title)
		if err != nil {
			fmt.Printf("Error extracting Jira ticket ID: %v \n", err)
		} else {
			completion = fmt.Sprintf("### JIRA ticket: [%s](%s) \n\n%s", id, jira.GenerateJiraTicketURL(opts.JiraURL, id), completion)
		}
	}

	if opts.Test {
		fmt.Println(completion)
		return nil
	}

	// Update the pull request description
	fmt.Println("Updating pull request")
	updatePr := &github.PullRequest{Body: github.String(completion)}
	if _, err = githubClient.UpdatePullRequest(ctx, opts.Owner, opts.Repo, opts.PRNumber, updatePr); err != nil {
		return fmt.Errorf("error updating pull request: %w", err)
	}

	return nil
}

func genCompletionOnce(ctx context.Context, client *oAIClient.Client, diff *github.CommitsComparison) (string, error) {
	fmt.Println("Generating completion once")
	messages := make([]openai.ChatCompletionMessage, 0, len(diff.Files))
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: oAIClient.PromptDescribeChanges,
	})
	for _, file := range diff.Files {
		if file.Patch == nil {
			continue
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: *file.Patch,
		})
	}

	fmt.Println("Sending prompt to OpenAI")
	completion, err := client.ChatCompletion(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("error completing prompt: %w", err)
	}

	return completion, nil
}

func genCompletionPerFile(ctx context.Context, client *oAIClient.Client, diff *github.CommitsComparison, pr *github.PullRequest) (string, error) {
	fmt.Println("Generating completion per file")
	OverallDescribeCompletion := fmt.Sprintf("Pull request title: %s, body: %s\n\n", pr.GetTitle(), pr.GetBody())

	for i, file := range diff.Files {
		prompt := fmt.Sprintf(oAIClient.PromptDescribeChanges, *file.Patch)

		if len(prompt) > 4096 {
			prompt = fmt.Sprintf("%s...", prompt[:4093])
		}

		fmt.Printf("Sending prompt to OpenAI for file %d/%d\n", i+1, len(diff.Files))
		completion, err := client.ChatCompletion(ctx, []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		})
		if err != nil {
			return "", fmt.Errorf("error getting review: %w", err)
		}
		OverallDescribeCompletion += fmt.Sprintf("File: %s \nDescription: %s \n\n", file.GetFilename(), completion)
	}

	fmt.Println("Sending final prompt to OpenAI")
	overallCompletion, err := client.ChatCompletion(ctx, []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf(oAIClient.PromptOverallDescribe, OverallDescribeCompletion),
		},
	})
	if err != nil {
		return "", fmt.Errorf("error getting overall review: %w", err)
	}

	return overallCompletion, nil
}
