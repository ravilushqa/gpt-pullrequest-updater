package main

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

const (
	PromptUpdateDescription = "\"Generate a GitHub pull request description based on the following changes without basic prefix in markdown format with ###Description and ###Changes blocks:\\n\""
	PromptDescribeChanges   = "Bellow is the code patch, please describe the changes in the following format: **file name**:\\n###Description\\n###Changes\\n"
)

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(token string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(token),
	}
}

func (o *OpenAIClient) ChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	return "", nil
	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
