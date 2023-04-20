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

func (g *GithubClient) GetPullRequestFiles(ctx context.Context, owner, repo string, prNumber int) ([]*github.CommitFile, error) {
	files, _, err := g.client.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
	return files, err
}

func (g *GithubClient) GetPullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := g.client.PullRequests.Get(ctx, owner, repo, number)
	return pr, err
}

func (g *GithubClient) GetPullRequestDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	diff, _, err := g.client.PullRequests.GetRaw(ctx, owner, repo, number, github.RawOptions{Type: github.Diff})
	return diff, err
}

func (g *GithubClient) UpdatePullRequest(ctx context.Context, owner, repo string, number int, pr *github.PullRequest) (*github.PullRequest, error) {
	updatedPR, _, err := g.client.PullRequests.Edit(ctx, owner, repo, number, pr)
	return updatedPR, err
}
