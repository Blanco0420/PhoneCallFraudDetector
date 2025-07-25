package profanityAnalyzing

import (
	"bufio"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/Blanco0420/Phone-Number-Check/backend/config"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
	"github.com/Blanco0420/Phone-Number-Check/backend/utils"

	"github.com/agnivade/levenshtein"
)

type commentResult struct {
	index int
	text  string
	score int
}

const (
	LDNOOBWFileName = "LDNOOBW-bad.txt"
	LDNOOBWRawUrl   = "https://raw.githubusercontent.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words/refs/heads/master/ja"
)

var badWords map[string]struct{}

func Initialize() error {
	if exists := utils.CheckIfFileExists(LDNOOBWFileName); !exists {
		if err := downloadFile(LDNOOBWRawUrl, LDNOOBWFileName); err != nil {
			return err
		}
	}

	file, err := os.Open(LDNOOBWFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	if badWords == nil {
		badWords = make(map[string]struct{})
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			badWords[line] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func scoreComment(comment string) int {
	score := 0

	// Tokenize and classify hits
	greenHits, yellowHits, redHits := checkCommentForHits(comment)
	score += len(greenHits) * 1
	score += len(yellowHits) * 3
	score += len(redHits) * 6

	// Also score based on exact match or fuzzy match with bad words
	for word := range badWords {
		if strings.Contains(comment, word) {
			score += 2
		} else if levenshtein.ComputeDistance(comment, word) <= config.LevenshteinThreshold {
			score += 1
		}
	}

	return score
}

func ScoreComments(comments []providers.Comment, concurrency int) int {
	in := make(chan int, len(comments))
	out := make(chan int, len(comments))

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for range concurrency {
		go func() {
			defer wg.Done()
			for commentIndex := range in {
				text := comments[commentIndex].Text
				score := scoreComment(text)
				out <- score
			}
		}()
	}

	go func() {
		for i := range comments {
			in <- i
		}
		close(in)
	}()

	go func() {
		wg.Wait()
		close(out)
	}()

	total := 0
	for s := range out {
		total += s
	}

	// Cap score to 20 max
	return int(math.Min(float64(total), 20))
}
