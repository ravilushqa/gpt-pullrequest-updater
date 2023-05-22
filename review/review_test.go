package review

import (
	"context"
	"testing"

	"github.com/google/go-github/v51/github"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCompleter struct {
	mock.Mock
}

func (m *MockCompleter) ChatCompletion(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	args := m.Called(ctx, messages)
	return args.String(0), args.Error(1)
}

type MockPullRequestUpdater struct {
	mock.Mock
}

func (m *MockPullRequestUpdater) CreatePullRequestComment(ctx context.Context, owner, repo string, number int, comment *github.PullRequestComment) (*github.PullRequestComment, error) {
	args := m.Called(ctx, owner, repo, number, comment)
	return args.Get(0).(*github.PullRequestComment), args.Error(1)
}

func TestGenerateCommentsFromDiff(t *testing.T) {
	testCases := []struct {
		name           string
		mockResponse   string
		expectedResult int
	}{
		{
			name: "Single issue",
			mockResponse: `{
				"quality": "bad",
				"issues": [
					{
						"type": "coding-style",
						"line": 3,
						"description": "Inconsistent indentation"
					}
				]
			}`,
			expectedResult: 1,
		},
		{
			name: "No issues",
			mockResponse: `{
				"quality": "good",
				"issues": []
			}`,
			expectedResult: 0,
		},
		{
			name: "dirty response",
			mockResponse: `
			Review result is inconsistent
			{
				"quality": "bad",
				"issues": [
					{
						"type": "coding-style",
						"line": 3,
						"description": "Inconsistent indentation"
					},
					{
						"type": "bug",
						"line": 4,
						"description": "Missing semicolon"
					}
				]
			}
			Review result is inconsistent
			`,
			expectedResult: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCompleter := new(MockCompleter)
			mockDiff := &github.CommitsComparison{
				Files: []*github.CommitFile{
					{
						Filename: ptrOf("file1").(*string),
						Patch:    ptrOf("patch1").(*string),
						Status:   ptrOf("modified").(*string),
					},
				},
				Commits: []*github.RepositoryCommit{
					{
						Commit: &github.Commit{
							SHA: ptrOf("sha1").(*string),
						},
					},
				},
			}

			mockCompleter.On("ChatCompletion", mock.Anything, mock.Anything).Return(tc.mockResponse, nil)

			comments, err := GenerateCommentsFromDiff(context.Background(), mockCompleter, mockDiff)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, len(comments))
		})
	}
}

func TestPushComments(t *testing.T) {
	testCases := []struct {
		name     string
		comments []*github.PullRequestComment
	}{
		{
			name: "Single comment",
			comments: []*github.PullRequestComment{
				{
					Line: ptrOf(3).(*int),
					Body: ptrOf("Inconsistent indentation").(*string),
					Path: ptrOf("file1").(*string),
				},
			},
		},
		{
			name:     "No comments",
			comments: []*github.PullRequestComment{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPullRequestUpdater := new(MockPullRequestUpdater)
			owner := "owner"
			repo := "repo"
			number := 1

			for _, comment := range tc.comments {
				commentResponse := &github.PullRequestComment{
					Line: comment.Line,
					Body: comment.Body,
					Path: comment.Path,
				}

				mockPullRequestUpdater.On("CreatePullRequestComment", mock.Anything, owner, repo, number, comment).Return(commentResponse, nil)
			}

			err := PushComments(context.Background(), mockPullRequestUpdater, owner, repo, number, tc.comments)

			assert.NoError(t, err)
			mockPullRequestUpdater.AssertExpectations(t)
		})
	}
}

func ptrOf(i interface{}) interface{} {
	switch v := i.(type) {
	case int:
		return &v
	case string:
		return &v
	}
	return &i
}
