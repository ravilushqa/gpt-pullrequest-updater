package description

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v51/github"
	"github.com/sashabaranov/go-openai"

	oAIClient "github.com/ravilushqa/gpt-pullrequest-updater/openai"
)

const placeholder = "gpt-updater:description"
const placeholderFinished = "<!-- gpt-updater:description -->"

type Info struct {
	Completion   string
	JiraInfo     string
	ShortcutInfo string
}

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

func BuildUpdatedPullRequest(existingDescription *string, info Info) *github.PullRequest {

	desc := ""

	if info.JiraInfo != "" {
		desc = info.JiraInfo + "\n\n" + desc
	}

	if info.ShortcutInfo != "" {
		desc = info.ShortcutInfo + "\n\n" + desc
	}

	if info.Completion != "" {
		desc += info.Completion
	}

	builtBody := fmt.Sprintf("%s\n## 🤖 gpt-updater description\n%s", placeholderFinished, desc)

	if existingDescription != nil && needToUpdateByPlaceholder(*existingDescription) {
		builtBody = strings.Replace(*existingDescription, placeholder, builtBody, 1)
	}

	return &github.PullRequest{Body: github.String(builtBody)}
}

func IsDescriptionFinished(existingDescription string) bool {
	return strings.Contains(existingDescription, placeholderFinished)
}

func needToUpdateByPlaceholder(existingDescription string) bool {
	return !strings.Contains(existingDescription, placeholderFinished) &&
		strings.Contains(existingDescription, placeholder)
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

	fmt.Println("Completion:", completion)

	return completion, nil
}

func genCompletionPerFile(ctx context.Context, client *oAIClient.Client, diff *github.CommitsComparison, pr *github.PullRequest) (string, error) {
	fmt.Println("Generating completion per file")
	OverallDescribeCompletion := fmt.Sprintf("Pull request title: %s, body: %s\n\n", pr.GetTitle(), pr.GetBody())

	for i, file := range diff.Files {
		patch := file.GetPatch()
		if patch == "" {
			continue
		}
		maxLength := 4096 - len(oAIClient.PromptDescribeChanges)
		if len(*file.Patch) > maxLength {
			fmt.Println("Patch is too long, truncating")
			patch = fmt.Sprintf("%s...", patch[:maxLength])
		}

		fmt.Printf("processing file: %s %d/%d\n", file.GetFilename(), i+1, len(diff.Files))
		completion, err := client.ChatCompletion(ctx, []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: oAIClient.PromptDescribeChanges,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: patch,
			},
		})
		if err != nil {
			return "", fmt.Errorf("error getting review: %w", err)
		}
		fmt.Println("Completion:", completion)

		OverallDescribeCompletion += fmt.Sprintf("File: %s \nDescription: %s \n\n", file.GetFilename(), completion)
	}

	fmt.Println("Summarizing overall completion")
	overallCompletion, err := client.ChatCompletion(ctx, []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: oAIClient.PromptDescribeOverall,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: OverallDescribeCompletion,
		},
	})
	if err != nil {
		return "", fmt.Errorf("error completing final prompt: %w", err)
	}

	fmt.Println("Overall completion:", overallCompletion)

	return overallCompletion, nil
}
