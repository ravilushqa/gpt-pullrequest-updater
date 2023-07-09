package shortcut

import (
	"fmt"
	"regexp"
)

const storyUrlFormat = "%s/story/%s"

func ExtractShortcutStoryId(title string) (string, error) {

	// This regular expression pattern matches a Shortcut story ID (e.g. sc-12345).
	pattern := `sc-([\d]+)`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("error compiling regex: %w", err)
	}

	matches := re.FindStringSubmatch(title)
	if len(matches) < 2 {
		return "", fmt.Errorf("no Shortcut story ID found in the input string")
	}

	return matches[1], nil
}

func GenerateShortcutStoryUrl(shortcutBaseUrl, ticketId string) string {
	return fmt.Sprintf(storyUrlFormat, shortcutBaseUrl, ticketId)
}
