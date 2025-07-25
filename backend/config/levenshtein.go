package config

import (
	"strconv"
)

var LevenshteinThreshold int

func initLevenshtein() {
	if levenshteinEnvValue, exists := GetEnvVariable("LEVENSHTEIN_THRESHOLD"); !exists {
		LevenshteinThreshold = 2
		return
	} else {
		parsed, err := strconv.Atoi(levenshteinEnvValue)
		if err != nil {
			LevenshteinThreshold = 2
			return
		}
		LevenshteinThreshold = parsed
	}

}
