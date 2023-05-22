package review

import (
	"context"
	"fmt"

	"github.com/google/go-github/v51/github"
	"github.com/sashabaranov/go-openai"

	oAIClient "github.com/ravilushqa/gpt-pullrequest-updater/openai"
)

type PullRequestUpdater interface {
	CreatePullRequestComment(ctx context.Context, owner, repo string, number int, comment *github.PullRequestComment) (*github.PullRequestComment, error)
}

type Completer interface {
	ChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error)
}

type Review struct {
	Quality Quality `json:"quality"`
	Issues  []Issue `json:"issues"`
}

type Issue struct {
	Type        string `json:"type"`
	Line        int    `json:"line"`
	Description string `json:"description"`
}

type Quality string

const (
	Good    Quality = "good"
	Bad     Quality = "bad"
	Neutral Quality = "neutral"
)

func GenerateCommentsFromDiff(ctx context.Context, openAIClient Completer, diff *github.CommitsComparison) ([]*github.PullRequestComment, error) {
	var comments []*github.PullRequestComment

	for i, file := range diff.Files {
		patch := file.GetPatch()
		fmt.Printf("processing file: %s %d/%d\n", file.GetFilename(), i+1, len(diff.Files))
		if patch == "" || file.GetStatus() == "removed" || file.GetStatus() == "renamed" {
			continue
		}

		maxLength := 4096 - len(oAIClient.PromptReview)
		if len(patch) > maxLength {
			fmt.Println("Patch is too long, truncating")
			patch = fmt.Sprintf("%s...", patch[:maxLength])
		}
		completion, err := openAIClient.ChatCompletion(ctx, []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: oAIClient.PromptReview,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: patch,
			},
		})

		if err != nil {
			return nil, fmt.Errorf("error getting completion: %w", err)
		}

		fmt.Println("Completion:", completion)

		review, err := extractReviewFromString(completion)
		if err != nil {
			fmt.Println("Error extracting JSON:", err)
			continue
		}

		if review.Quality == Good {
			fmt.Println("Review is good")
			continue
		}
		for _, issue := range review.Issues {
			if issue.Line == 0 {
				fmt.Printf("Skipping file-level issue: %v\n", issue)
				continue // TODO: add support for file-level issues
			}
			body := fmt.Sprintf("[%s] %s", issue.Type, issue.Description)
			comment := &github.PullRequestComment{
				CommitID: diff.Commits[len(diff.Commits)-1].SHA,
				Path:     file.Filename,
				Body:     &body,
				Position: &issue.Line,
			}
			comments = append(comments, comment)
		}
	}

	return comments, nil
}

func PushComments(ctx context.Context, prUpdated PullRequestUpdater, owner, repo string, number int, comments []*github.PullRequestComment) error {
	for i, c := range comments {
		fmt.Printf("creating comment: %s %d/%d\n", *c.Path, i+1, len(comments))
		if _, err := prUpdated.CreatePullRequestComment(ctx, owner, repo, number, c); err != nil {
			fmt.Printf("error creating comment: %s\n%+v", err, *c) // TODO: return error instead of printing
		}
	}
	return nil
}
