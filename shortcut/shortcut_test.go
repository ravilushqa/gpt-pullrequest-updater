package shortcut

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractShortcutStoryID(t *testing.T) {
	testCases := []struct {
		title       string
		expectedID  string
		expectedErr bool
	}{
		{
			title:       "This is a sample title with a valid story ID sc-12345",
			expectedID:  "12345",
			expectedErr: false,
		},
		{
			title:       "No story ID in this title",
			expectedID:  "",
			expectedErr: true,
		},
		{
			title:       "Invalid story ID format sc-abcde",
			expectedID:  "",
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		id, err := ExtractShortcutStoryID(tc.title)

		if tc.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedID, id)
		}
	}
}

func TestGenerateShortcutStoryURL(t *testing.T) {
	baseURL := "https://app.shortcut.com/foo"
	ticketID := "12345"
	expectedURL := "https://app.shortcut.com/foo/story/12345"

	url := GenerateShortcutStoryURL(baseURL, ticketID)
	assert.Equal(t, expectedURL, url)
}
