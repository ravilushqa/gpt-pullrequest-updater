package jira

import "testing"

func TestExtractJiraTicketID(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "Valid ticket ID",
			input:       "This is a sample text with a JIRA ticket ID: PROJ-123, let's extract it.",
			expected:    "PROJ-123",
			expectError: false,
		},
		{
			name:        "No ticket ID",
			input:       "This is a sample text without a JIRA ticket ID.",
			expectError: true,
		},
		{
			name:        "Multiple ticket IDs",
			input:       "This text has multiple JIRA ticket IDs: PROJ-123, TASK-456, and BUG-789.",
			expected:    "PROJ-123",
			expectError: false,
		},
		{
			name:        "Valid ticket ID. Lowercase.",
			input:       "This is an invalid JIRA ticket ID: Proj-123.",
			expected:    "Proj-123",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ExtractJiraTicketID(tc.input)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if result != tc.expected {
					t.Errorf("expected result '%s', but got '%s'", tc.expected, result)
				}
			}
		})
	}
}
