package main

import (
	"context"

	"github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
)

const baseGithubURL = "https://api.github.com"

type GithubClient struct {
	client *github.Client
}

func NewGithubClient(ctx context.Context, token string) *GithubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &GithubClient{github.NewClient(tc)}
}

func (c *GithubClient) GetPullRequestFiles(ctx context.Context, owner, repo string, prNumber int) ([]*github.CommitFile, error) {
	files, _, err := c.client.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
	return files, err
}

func (c *GithubClient) GetPullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, number)
	return pr, err
}

func (c *GithubClient) GetPullRequestDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	diff, _, err := c.client.PullRequests.GetRaw(ctx, owner, repo, number, github.RawOptions{Type: github.Diff})
	return diff, err
}

func (c *GithubClient) UpdatePullRequest(ctx context.Context, owner, repo string, number int, pr *github.PullRequest) (*github.PullRequest, error) {
	updatedPR, _, err := c.client.PullRequests.Edit(ctx, owner, repo, number, pr)
	return updatedPR, err
}

func (c *GithubClient) CreateComment(ctx context.Context, owner, repo string, number int, comment *github.PullRequestComment) (*github.PullRequestComment, error) {
	createdComment, _, err := c.client.PullRequests.CreateComment(ctx, owner, repo, number, comment)
	return createdComment, err
}

func (c *GithubClient) CompareCommits(ctx context.Context, owner, repo, base, head string) (*github.CommitsComparison, error) {
	comp, _, err := c.client.Repositories.CompareCommits(ctx, owner, repo, base, head, nil)
	return comp, err
}
