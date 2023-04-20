package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v51/github"
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
	SkipFiles   string `long:"skip-files" env:"SKIP_FILES" description:"Skip files. Comma separated list" default:"go.mod,go.sum,.pb.go"`
}

// FileDiff represents a single file diff.
type FileDiff struct {
	Header string
	Diff   string
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("Error parsing flags: %v \n", err)
		}
		os.Exit(0)
	}
	openaiClient := NewOpenAIClient(opts.OpenAIToken)
	githubClient := NewGithubClient(context.Background(), opts.GithubToken)

	diff, err := githubClient.GetPullRequestDiff(context.Background(), opts.Owner, opts.Repo, opts.PRNumber)
	if err != nil {
		fmt.Printf("Error getting pull request diff: %v\n", err)
		return
	}
	filesDiff, err := parseGitDiffAndSplitPerFile(diff, strings.Split(opts.SkipFiles, ","))
	if err != nil {
		return
	}

	var messages []openai.ChatCompletionMessage
	prompt := fmt.Sprintf("Generate a GitHub pull request description based on the following changes without basic prefix in markdown format with ###Description and ###Changes blocks:\n")
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	for _, fileDiff := range filesDiff {
		fileName := getFilenameFromDiffHeader(fileDiff.Header)

		prompt := fmt.Sprintf("File %s:\n%s\n%s\n", fileName, fileDiff.Header, fileDiff.Diff)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		})
	}
	chatGPTDescription, err := openaiClient.ChatCompletion(context.Background(), messages)
	if err != nil {
		fmt.Printf("Error generating pull request description: %v\n", err)
		return
	}

	pr, err := githubClient.GetPullRequest(context.Background(), opts.Owner, opts.Repo, opts.PRNumber)
	if err != nil {
		fmt.Printf("Error getting pull request: %v\n", err)
		return
	}

	jiraLink := generateJiraLinkByTitle(*pr.Title)

	description := fmt.Sprintf("### Jira\n%s\n%s", jiraLink, chatGPTDescription)
	if opts.Test {
		fmt.Println(description)
		os.Exit(0)
	}
	// Update the pull request with the generated description
	_, err = githubClient.UpdatePullRequest(
		context.Background(), opts.Owner, opts.Repo, opts.PRNumber, &github.PullRequest{Body: &description},
	)
	if err != nil {
		fmt.Printf("Error updating pull request: %v\n", err)
		return
	}
}

// parseGitDiffAndSplitPerFile parses a git diff and splits it into a slice of FileDiff.
func parseGitDiffAndSplitPerFile(diff string, skipFiles []string) ([]FileDiff, error) {
	lines := strings.Split(diff, "\n")
	var fileDiffs []FileDiff

	inFileDiff, isSkipFile := false, false
	var currentFileDiff FileDiff
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if inFileDiff {
				fileDiffs = append(fileDiffs, currentFileDiff)
			}
			if len(skipFiles) > 0 {
				isSkipFile = false
				for _, skipFile := range skipFiles {
					if strings.Contains(line, skipFile) {
						isSkipFile = true
						break
					}
				}
			}
			if isSkipFile {
				continue
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

func generateJiraLinkByTitle(title string) string {
	//NCR-1234
	issueKey := strings.ToUpper(strings.Split(title, " ")[0])
	if !strings.HasPrefix(issueKey, "NCR-") {
		return ""
	}
	jiraBaseURL := "https://jira.deliveryhero.com/browse/"

	return fmt.Sprintf("[%s](%s%s)", issueKey, jiraBaseURL, issueKey)
}
