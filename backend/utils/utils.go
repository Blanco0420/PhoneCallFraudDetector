package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
)

var commonSuffixes = []string{
	"株式会社",
	"有限会社",
	"合同会社",
	"球場",
	"支店",
	"センター",
	"本社",
	"営業所",
	"店",
}

/*
01 - month
02 - day
03 - hour (12h)
04 - minute
05 - second
06 - year
07 - time zone offset
*/
func ParseDate(layout, date string) (time.Time, error) {
	parsedDate, err := time.Parse(layout, date)
	if err != nil {
		return time.Time{}, err
	}
	return parsedDate, nil
}

func GetSuffixesFromCompanyName(original *string) (cleaned string, foundSuffixes []string) {
	cleaned = *original
	foundSuffixes = []string{}
	for _, suffix := range commonSuffixes {
		if strings.Contains(cleaned, suffix) {
			cleaned = strings.ReplaceAll(cleaned, suffix, "")
			foundSuffixes = append(foundSuffixes, suffix)
		}
	}
	return
}

func CheckIfFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func CleanText(text string) string {
	re := regexp.MustCompile(`[^\p{L}\p{N}]`)
	return re.ReplaceAllString(text, "")
}

var lineTypeMap = map[string]providers.LineType{
	"landline": providers.LineTypeLandline,
	"固定電話":     providers.LineTypeLandline,

	"mobile":   providers.LineTypeMobile,
	"携帯電話":     providers.LineTypeMobile,
	"携帯電話/PHS": providers.LineTypeMobile,

	"tollfree": providers.LineTypeTollFree,
	"freedial": providers.LineTypeTollFree,
	"フリーダイヤル":  providers.LineTypeTollFree,

	"voip": providers.LineTypeVOIP,
	"ip電話": providers.LineTypeVOIP,
}

var lineTypeOtherMap = map[string]struct{}{
	"premiumrate": {},
	"paging":      {},
	"satellite":   {},
}

func GetLineType(rawLineType string) (providers.LineType, error) {
	cleaned := strings.TrimSpace(rawLineType)
	cleaned = strings.ToLower(strings.Trim(rawLineType, "-_"))
	if val, ok := lineTypeMap[cleaned]; ok {
		return val, nil
	}

	if _, ok := lineTypeOtherMap[cleaned]; ok {
		return providers.LineTypeOther, fmt.Errorf("Line type is of type other. Actual text: %s", rawLineType)
	}
	return providers.LineTypeOther, nil
}

func AverageIntSlice(vals []int) int {
	if len(vals) == 0 {
		return 0
	}
	total := 0
	for _, v := range vals {
		total += v
	}
	return total / len(vals)
}

func ParseGraphData(rawDataString string, graphData *[]providers.GraphData) error {
	var rawData [][]any
	if err := json.Unmarshal([]byte(rawDataString), &rawData); err != nil {
		return err
	}
	for _, row := range rawData {
		if len(row) == 2 {
			date, _ := row[0].(string)
			parsedDate, err := ParseDate("2006-01-02", date)
			if err != nil {
				return err
			}

			var accesses int
			switch v := row[1].(type) {
			case float64:
				accesses = int(v)
			case string:
				accesses, err = strconv.Atoi(v)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("Unexpected type for %q: %T", v, v)
			}

			if err != nil {
				return err
			}
			*graphData = append(*graphData, providers.GraphData{Date: parsedDate, Accesses: accesses})
		}
	}
	return nil
}
