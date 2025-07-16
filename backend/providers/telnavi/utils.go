package telnavi

import (
	"regexp"
	"strconv"

	"github.com/Blanco0420/Phone-Number-Check/backend/utils"
)

func extractBusinessName(text *string) (cleaned string, suffixes []string) {
	re := regexp.MustCompile("[0-9]")
	cleaned = re.ReplaceAllString(*text, "")
	cleaned, suffixes = utils.GetSuffixesFromCompanyName(&cleaned)
	return
}

func getCleanRating(rawUserRating string) (float32, error) {
	var rating float32
	re := regexp.MustCompile(`[^0-9.]`)
	cleaned := re.ReplaceAllString(rawUserRating, "")
	if cleaned == "" {
		rating = 0
	} else {
		f64UserRating, err := strconv.ParseFloat(cleaned, 32)
		if err != nil {
			return 0, err
		}
		userRating := float32(f64UserRating)
		rating = userRating

	}
	return rating, nil
}
