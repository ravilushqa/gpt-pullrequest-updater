package description

import (
	"context"
	"fmt"

	"github.com/google/go-github/v51/github"
	"github.com/sashabaranov/go-openai"

	oAIClient "github.com/ravilushqa/gpt-pullrequest-updater/openai"
)

func GenerateCompletion(ctx context.Context, client *oAIClient.Client, diff *github.CommitsComparison, pr *github.PullRequest) (string, error) {
	sumDiffs := calculateSumDiffs(diff)

	var completion string
	var err error
	if sumDiffs < 4000 {
		completion, err = genCompletionOnce(ctx, client, diff)
	} else {
		completion, err = genCompletionPerFile(ctx, client, diff, pr)
	}

	return completion, err
}

func calculateSumDiffs(diff *github.CommitsComparison) int {
	sumDiffs := 0
	for _, file := range diff.Files {
		if file.Patch == nil {
			continue
		}
		sumDiffs += len(*file.Patch)
	}
	return sumDiffs
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
		if file.Patch == nil {
			continue
		}
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
			Content: fmt.Sprintf(oAIClient.PromptDescribeOverall, OverallDescribeCompletion),
		},
	})
	if err != nil {
		return "", fmt.Errorf("error completing final prompt: %w", err)
	}

	return overallCompletion, nil
}
