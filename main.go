package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/sashabaranov/go-openai"
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
	if _, err := flags.Parse(&opts); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("Error parsing flags: %v \n", err)
		}
		os.Exit(0)
	}
	openaiClient := openai.NewClient(opts.OpenAIToken)

	title, err := getPullRequestTitle(opts.GithubToken, opts.Owner, opts.Repo, opts.PRNumber)
	if err != nil {
		return
	}

	jiraLink := generateJiraLinkByTitle(title)

	diff, err := getDiffContent(opts.GithubToken, opts.Owner, opts.Repo, opts.PRNumber)
	if err != nil {
		fmt.Printf("Error fetching diff content: %v\n", err)
		return
	}
	filesDiff, err := ParseGitDiffAndSplitPerFile(diff)
	if err != nil {
		return
	}

	var messages []openai.ChatCompletionMessage
	prompt := fmt.Sprintf(
		"Generate a GitHub pull request description based on the following changes " +
			"without basic prefix in markdown format with ###Description and ###Changes blocks:\n",
	)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	for _, fileDiff := range filesDiff {
		prompt := fmt.Sprintf("File %s:\n%s\n%s\n", getFilenameFromDiffHeader(fileDiff.Header), fileDiff.Header, fileDiff.Diff)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		})
	}
	chatGPTDescription, err := generatePRDescription(openaiClient, messages)
	if err != nil {
		fmt.Printf("Error generating pull request description: %v\n", err)
		return
	}

	description := fmt.Sprintf("## Jira\n%s\n%s", jiraLink, chatGPTDescription)
	if opts.Test {
		fmt.Println(description)
		os.Exit(0)
	}
	// Update the pull request with the generated description
	err = updatePullRequestDescription(opts.GithubToken, opts.Owner, opts.Repo, opts.PRNumber, description)
	if err != nil {
		fmt.Printf("Error updating pull request description: %v\n", err)
		return
	}

	fmt.Println("Pull request description updated successfully")
}

func updatePullRequestDescription(token string, o string, r string, number int, description string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", o, r, number)

	data := map[string]string{
		"body": description,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update pull request description. Status code: %d", resp.StatusCode)
	}

	return nil
}

func getDiffContent(token, owner, repo string, prNumber int) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNumber)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Accept", "application/vnd.github.v3.diff")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
	//req.Header.Add("Cookie", "logged_in=no")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func generatePRDescription(client *openai.Client, messages []openai.ChatCompletionMessage) (string, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
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

// FileDiff represents a single file diff.
type FileDiff struct {
	Header string
	Diff   string
}

// ParseGitDiffAndSplitPerFile parses a git diff and splits it into a slice of FileDiff.
func ParseGitDiffAndSplitPerFile(diff string) ([]FileDiff, error) {
	lines := strings.Split(diff, "\n")
	var fileDiffs []FileDiff

	inFileDiff := false
	var currentFileDiff FileDiff
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if inFileDiff {
				fileDiffs = append(fileDiffs, currentFileDiff)
			}
			currentFileDiff = FileDiff{Header: line}
			inFileDiff = true
		} else if inFileDiff {
			currentFileDiff.Diff += line + "\n"
		}
	}
	if inFileDiff {
		fileDiffs = append(fileDiffs, currentFileDiff)
	}

	return fileDiffs, nil
}

func getFilenameFromDiffHeader(diffHeader string) string {
	// Split the diff header into individual lines
	lines := strings.Split(diffHeader, "\n")

	// Extract the filename from the "diff --git" line
	gitDiffLine := lines[0]
	parts := strings.Split(gitDiffLine, " ")
	oldFileName := strings.TrimPrefix(parts[2], "a/")
	newFileName := strings.TrimPrefix(parts[3], "b/")

	// Return the new filename if it exists, otherwise return the old filename
	if newFileName != "/dev/null" {
		return newFileName
	} else {
		return oldFileName
	}
}

func getPullRequestTitle(token, owner, repo string, prNumber int) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to fetch pull request details. Status code: %d", resp.StatusCode)
	}

	var pr struct {
		Title string `json:"title"`
	}

	err = json.NewDecoder(resp.Body).Decode(&pr)
	if err != nil {
		return "", err
	}

	return pr.Title, nil
}

func generateJiraLinkByTitle(title string) string {
	//NCR-1234
	issueKey := strings.ToUpper(strings.Split(title, " ")[0])
	if !strings.HasPrefix(issueKey, "NCR-") {
		return ""
	}
	jiraBaseURL := "https://jira.deliveryhero.com/browse/"

	return fmt.Sprintf("[%s](%s%s)", issueKey, jiraBaseURL, issueKey)
}
