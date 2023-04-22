package github

import (
	"context"

	"github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
}

func NewClient(ctx context.Context, token string) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{github.NewClient(tc)}
}

func (c *Client) GetPullRequestFiles(ctx context.Context, owner, repo string, prNumber int) ([]*github.CommitFile, error) {
	files, _, err := c.client.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
	return files, err
}

func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, number)
	return pr, err
}

func (c *Client) GetPullRequestDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	diff, _, err := c.client.PullRequests.GetRaw(ctx, owner, repo, number, github.RawOptions{Type: github.Diff})
	return diff, err
}

func (c *Client) UpdatePullRequest(ctx context.Context, owner, repo string, number int, pr *github.PullRequest) (*github.PullRequest, error) {
	updatedPR, _, err := c.client.PullRequests.Edit(ctx, owner, repo, number, pr)
	return updatedPR, err
}

func (c *Client) CreatePullRequestComment(ctx context.Context, owner, repo string, number int, comment *github.PullRequestComment) (*github.PullRequestComment, error) {
	createdComment, _, err := c.client.PullRequests.CreateComment(ctx, owner, repo, number, comment)
	return createdComment, err
}

func (c *Client) CreateReview(ctx context.Context, owner, repo string, number int, comment *github.PullRequestReviewRequest) (*github.PullRequestReview, error) {
	createdReview, _, err := c.client.PullRequests.CreateReview(ctx, owner, repo, number, comment)
	return createdReview, err
}

func (c *Client) CompareCommits(ctx context.Context, owner, repo, base, head string) (*github.CommitsComparison, error) {
	comp, _, err := c.client.Repositories.CompareCommits(ctx, owner, repo, base, head, nil)
	return comp, err
}
