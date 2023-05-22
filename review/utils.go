package review

import (
	"encoding/json"
	"errors"
)

func extractReviewFromString(input string) (*Review, error) {
	var jsonObj *Review

	// find the start and end positions of the JSON object
	start := 0
	end := len(input)
	for i, c := range input {
		if c == '{' {
			start = i
			break
		}
		if i == len(input)-1 {
			return nil, errors.New("invalid JSON object")
		}
	}
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] == '}' {
			end = i + 1
			break
		}

		if i == 0 {
			return nil, errors.New("invalid JSON object")
		}
	}

	// extract the JSON object from the input
	jsonStr := input[start:end]
	err := json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		return nil, errors.New("invalid JSON object")
	}

	return jsonObj, nil
}
