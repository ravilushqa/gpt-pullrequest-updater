package review

import (
	"reflect"
	"testing"
)

func Test_extractReviewFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Review
		wantErr bool
	}{
		{
			name:  "correctly parses a review",
			input: `{ "quality": "good", "issues": [] }`,
			want: &Review{
				Quality: Good,
				Issues:  []Issue{},
			},
			wantErr: false,
		},
		{
			name:  "correctly parses a review with issues",
			input: `{ "quality": "good", "issues": [{ "type": "typo", "line": 1, "description": "typo" }] }`,
			want: &Review{
				Quality: Good,
				Issues:  []Issue{{Type: "typo", Line: 1, Description: "typo"}},
			},
			wantErr: false,
		},
		{
			name:  "correctly parses a review with multiple issues",
			input: `{ "quality": "good", "issues": [{ "type": "typo", "line": 1, "description": "typo" }, { "type": "typo", "line": 2, "description": "typo" }] }`,
			want: &Review{
				Quality: Good,
				Issues:  []Issue{{Type: "typo", Line: 1, Description: "typo"}, {Type: "typo", Line: 2, Description: "typo"}},
			},
			wantErr: false,
		},
		{
			name:  "correctly parses a review with prefix and suffix",
			input: `Review: { "quality": "good", "issues": [] } Done`,
			want: &Review{
				Quality: Good,
				Issues:  []Issue{},
			},
		},
		{
			name:  "correctly parses a review with prefix and suffix and issues",
			input: `Review: { "quality": "good", "issues": [{ "type": "typo", "line": 1, "description": "typo" }] } Done`,
			want: &Review{
				Quality: Good,
				Issues:  []Issue{{Type: "typo", Line: 1, Description: "typo"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractReviewFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractReviewFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractReviewFromString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
