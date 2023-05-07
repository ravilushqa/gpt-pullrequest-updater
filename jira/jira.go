package jira

import (
	"fmt"
	"regexp"
)

const ticketURLFormat = "%s/browse/%s"

// ExtractJiraTicketID returns the first JIRA ticket ID found in the input string.
func ExtractJiraTicketID(s string) (string, error) {
	// This regular expression pattern matches a JIRA ticket ID (e.g. PROJ-123).
	pattern := `([aA-zZ]+-\d+)`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("error compiling regex: %w", err)
	}

	matches := re.FindStringSubmatch(s)
	if len(matches) == 0 {
		return "", fmt.Errorf("no JIRA ticket ID found in the input string")
	}

	return matches[0], nil
}

func GenerateJiraTicketURL(jiraURL, ticketID string) string {
	return fmt.Sprintf(ticketURLFormat, jiraURL, ticketID)
}
