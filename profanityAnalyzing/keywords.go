package profanityAnalyzing

import (
	"PhoneNumberCheck/config"
	"strings"

	"github.com/agnivade/levenshtein"
)

var badWords map[string]struct{}

func containsExactMatch(text string) bool {
	for word := range badWords {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}
func ContainsFuzzyMatch(text string) bool {
	for word := range badWords {
		if levenshtein.ComputeDistance(text, word) <= config.LevenshteinThreshold {
			return true
		}
	}
	return false
}

func isBadByKeyword(text string) bool {
	text = strings.TrimSpace(text)
	return containsExactMatch(text) || ContainsFuzzyMatch(text)
}

//
// func classifyComment(text string) (bool, error) {
// 	if isBadByKeyword(text) {
// 		return true, nil
// 	}
// 	return false, nil
// 	//TODO: Implement this function
// 	// return isBadByMl(text)
// }
