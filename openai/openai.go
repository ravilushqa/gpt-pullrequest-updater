package openai

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

//go:embed prompts/review
var PromptReview string

//go:embed prompts/describe_changes
var PromptDescribeChanges string

//go:embed prompts/describe_overall
var PromptDescribeOverall string

type Client struct {
	client *openai.Client
	model  string
}

func NewClient(token, model string) *Client {
	return &Client{
		client: openai.NewClient(token),
		model:  model,
	}
}

func (c *Client) ChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       c.model,
			Messages:    messages,
			Temperature: 0.1,
		},
	)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			return "", err
		}
		fmt.Println("Error completing prompt:", err)
		fmt.Println("Retrying after 1 minute")
		// retry once after 1 minute
		time.Sleep(time.Minute)
		resp, err = c.client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:       c.model,
				Messages:    messages,
				Temperature: 0.1,
			},
		)
		if err != nil {
			return "", fmt.Errorf("error completing prompt: %w", err)
		}
	}

	return resp.Choices[0].Message.Content, nil
}
