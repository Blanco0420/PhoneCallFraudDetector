package profanityAnalyzing

import (
	"bufio"
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
	for word := range badWords {
		if strings.Contains(comment, word) {
			score += 2
		} else if levenshtein.ComputeDistance(comment, word) <= config.LevenshteinThreshold {
			score += 1
		}
	}
	return score
}

// TODO: Maybe remove concurrency as paramater and make an env var
func ScoreComments(comments []providers.Comment, concurrency int) int {
	in := make(chan int, len(comments))
	out := make(chan int, len(comments))

	var wg sync.WaitGroup
	wg.Add(concurrency)
	//TODO: Add the comment score to final data for ml use
	for range concurrency {
		go func() {
			defer wg.Done()
			for comment := range in {
				score := scoreComment(comments[comment].Text)
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
	return total
}
