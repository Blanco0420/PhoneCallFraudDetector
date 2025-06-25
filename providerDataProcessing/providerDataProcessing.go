package providerdataprocessing

import (
	"PhoneNumberCheck/config"
	japanesetokenizing "PhoneNumberCheck/japaneseTokenizing"
	"PhoneNumberCheck/profanityAnalyzing"
	"PhoneNumberCheck/providers"
	"PhoneNumberCheck/utils"
	"math"

	"github.com/agnivade/levenshtein"
)

func calculateGraphScore(data []providers.GraphData, recentAbuse *bool) int {
	score := 0

	n := len(data)
	if n < 3 {
		return score
	}

	todayAccesses := data[n-1].Accesses
	last3AvgAccesses := (todayAccesses + data[n-2].Accesses + data[n-3].Accesses) / 3

	if todayAccesses > 10 {
		score += 2
	}

	if last3AvgAccesses > 30 {
		score += 2
	}

	if last3AvgAccesses > 0 && todayAccesses > last3AvgAccesses*3 {
		score += 3
	}

	if last3AvgAccesses > 10 {
		*recentAbuse = true
	}

	nonZeroDayAccesses := 0
	for _, d := range data {
		if d.Accesses > 0 {
			nonZeroDayAccesses++
		}
	}
	if nonZeroDayAccesses > 7 {
		score += 1
	}
	return score
}

func EvaluateSource(input NumberRiskInput) int {
	score := 0

	if len(input.GraphData) > 0 {
		graphScore := calculateGraphScore(input.GraphData, input.RecentAbuse)
		score += int(math.Min(float64(graphScore)/10.0*30.0, 30))
	}

	if input.FraudScore != nil {
		score += int(math.Min(float64(*input.FraudScore)/100.0*25.0, 25))
	}

	if input.RecentAbuse != nil && *input.RecentAbuse {
		score += 15
	}

	if len(input.Comments) > 0 {
		commentScore := profanityAnalyzing.ScoreComments(input.Comments, 4)
		score += int(math.Min(float64(commentScore)/20.0*30.0, 30))
	}
	if score > 100 {
		return 100
	}

	return score
}

type FieldComparison struct {
	Values      []string
	SourceNames []string
}

// TODO: Instead of continuing when value has already been used, keep it in the loop and check if it get's a better score in another group
// func CalculateFieldConfidence(tokenizer *japanesetokenizing.Tokenizer, values []string, sourceNames []string) []ConfidenceResult {
//
//		threshold := config.LevenshteinThreshold
//		fmt.Println("Thresh: ", threshold)
//		groups := make([]ConfidenceResult, 0)
//		used := make([]bool, len(values))
//
//		for i, value := range values {
//			if used[i] {
//				continue
//			}
//			used[i] = true
//			group := ConfidenceResult{
//				NormalizedValue: value,
//				Supporters:      []string{sourceNames[i]},
//			}
//			for j := i + 1; j < len(values); j++ {
//				if used[j] {
//					continue
//				}
//				levDist := levenshtein.ComputeDistance(value, values[j])
//				commonToken := tokenizer.SharesCommonToken(value, values[j])
//				fmt.Printf("  Comparing to: '%s' → lev: %d, token match: %v\n", values[j], levDist, commonToken)
//				if levDist <= threshold || commonToken {
//					used[j] = true
//					group.Supporters = append(group.Supporters, sourceNames[j])
//				}
//			}
//			group.Confidence = float32(len(group.Supporters)) / float32(len(values))
//			groups = append(groups, group)
//		}
//		return groups
//	}
type ConfidenceResult struct {
	NormalizedValue string   // Representative name
	Supporters      []string // Source providers
	SimilarNames    []string // All names in this group
	Confidence      float32  // Group size / total input

	Suffixes   []string
	CoreTokens []string
}

type NameEntry struct {
	Name   string
	Source string
	Tokens map[string]struct{}
}

func preprocess(t *japanesetokenizing.Tokenizer, name string) (cleaned string, tokenSet map[string]struct{}) {
	cleaned, _ = utils.GetSuffixesFromCompanyName(&name)

	tokens := t.TokenizeJapanese(cleaned)

	tokenSet = make(map[string]struct{}, len(tokens))
	for _, tk := range tokens {
		tokenSet[tk] = struct{}{}
	}

	return
}

func weightedTokenOverlap(a, b map[string]struct{}) float32 {
	weight := func(token string) float32 {
		switch token {
		case "メルカリ", "楽天", "アマゾン", "ヤマト":
			return 2.0
		case "詐欺", "偽", "なりすまし":
			return 1.5
		case "確認", "番号", "認証", "電話":
			return 0.5
		default:
			return 1.0
		}
	}

	var sumWeights float32
	var matchedWeights float32

	allTokens := make(map[string]struct{})
	for t := range a {
		allTokens[t] = struct{}{}
	}
	for t := range b {
		allTokens[t] = struct{}{}
	}

	for token := range allTokens {
		w := weight(token)
		sumWeights += w
		if _, inA := a[token]; inA {
			if _, inB := b[token]; inB {
				matchedWeights += w
			}
		}
	}

	if sumWeights == 0 {
		return 0
	}
	return matchedWeights / sumWeights
}

func CalculateFieldConfidence(tokenizer *japanesetokenizing.Tokenizer, values []string, sources []string) []ConfidenceResult {
	threshold := config.LevenshteinThreshold
	overlapThreshold := float32(0.15)

	type NameEntry struct {
		Name   string
		Source string
		Tokens map[string]struct{}
	}

	entries := make([]NameEntry, len(values))
	for i, val := range values {
		clean, tokens := preprocess(tokenizer, val)
		entries[i] = NameEntry{
			Name:   clean,
			Source: sources[i],
			Tokens: tokens,
		}
	}

	used := make([]bool, len(entries))
	var results []ConfidenceResult

	for i := range entries {
		if used[i] {
			continue
		}
		used[i] = true

		cluster := []NameEntry{entries[i]}
		queue := []int{i}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]

			for j := range entries {
				if used[j] {
					continue
				}
				lev := levenshtein.ComputeDistance(entries[cur].Name, entries[j].Name)
				tokOverlap := weightedTokenOverlap(entries[cur].Tokens, entries[j].Tokens)
				if lev <= threshold || tokOverlap > overlapThreshold {
					used[j] = true
					cluster = append(cluster, entries[j])
					queue = append(queue, j)
				}
			}
		}

		// Compute representative, coreTokens, supporters, similarNames
		rep := cluster[0].Name
		tokenFreq := make(map[string]int)
		for _, e := range cluster {
			if len(e.Name) > len(rep) {
				rep = e.Name
			}
			for tk := range e.Tokens {
				tokenFreq[tk]++
			}
		}
		var coreTokens []string
		for token, freq := range tokenFreq {
			if freq > 1 {
				coreTokens = append(coreTokens, token)
			}
		}
		var supporters []string
		var similarNames []string
		for _, e := range cluster {
			supporters = append(supporters, e.Source)
			similarNames = append(similarNames, e.Name)
		}

		results = append(results, ConfidenceResult{
			NormalizedValue: rep,
			Supporters:      supporters,
			SimilarNames:    similarNames,
			Confidence:      float32(len(cluster)) / float32(len(values)),
			CoreTokens:      coreTokens,
		})
	}
	return results
}

// func getSourceList(entries []NameEntry) []string {
// 	sources := make([]string, len(entries))
// 	for i, e := range entries {
// 		sources[i] = e.Source
// 	}
// 	return sources
// }
//
// func getNameList(entries []NameEntry) []string {
// 	names := make([]string, len(entries))
// 	for i, e := range entries {
// 		names[i] = e.Name
// 	}
// 	return names
// }
//
// func tokenOverlap(a, b map[string]struct{}) float32 {
// 	matches := 0
// 	for token := range a {
// 		if _, ok := b[token]; ok {
// 			matches++
// 		}
// 	}
// 	union := len(a) + len(b) - matches
// 	if union == 0 {
// 		return 0
// 	}
// 	return float32(matches) / float32(union)
// }
//
// func preprocess(t *japanesetokenizing.Tokenizer, name string) (string, map[string]struct{}) {
// 	clean := utils.CleanText(name)
// 	tokens := t.TokenizeJapanese(clean)
// 	tokenSet := make(map[string]struct{}, len(tokens))
// 	for _, tk := range tokens {
// 		tokenSet[tk] = struct{}{}
// 	}
// 	return clean, tokenSet
// }
//
// func CalculateFieldConfidence(tokenizer *japanesetokenizing.Tokenizer, values []string, sources []string) []ConfidenceResult {
// 	threshold := config.LevenshteinThreshold
// 	overlapThreshold := float32(0.15)
//
// 	entries := make([]NameEntry, len(values))
// 	for i := range values {
// 		clean, tokens := preprocess(tokenizer, values[i])
// 		entries[i] = NameEntry{
// 			Name:   clean,
// 			Source: sources[i],
// 			Tokens: tokens,
// 		}
// 	}
//
// 	used := make([]bool, len(entries))
// 	var results []ConfidenceResult
//
// 	for i := range entries {
// 		if used[i] {
// 			continue
// 		}
// 		used[i] = true
//
// 		cluster := []NameEntry{entries[i]}
// 		queue := []int{i}
//
// 		for len(queue) > 0 {
// 			cur := queue[0]
// 			queue = queue[1:]
//
// 			for j := range entries {
// 				if used[j] {
// 					continue
// 				}
// 				lev := levenshtein.ComputeDistance(entries[cur].Name, entries[j].Name)
// 				tokOverlap := tokenOverlap(entries[cur].Tokens, entries[j].Tokens)
// 				if lev <= threshold || tokOverlap >= overlapThreshold {
// 					used[j] = true
// 					cluster = append(cluster, entries[j])
// 					queue = append(queue, j)
// 				}
// 				fmt.Printf("Comparing '%s' and '%s' → lev: %d, tokOverlap: %.2f\n", entries[cur].Name, entries[j].Name, lev, tokOverlap)
// 			}
// 		}
//
// 		// Choose representative (longest name)
// 		rep := cluster[0].Name
// 		for _, e := range cluster {
// 			if len(e.Name) > len(rep) {
// 				rep = e.Name
// 			}
// 		}
//
// 		results = append(results, ConfidenceResult{
// 			NormalizedValue: rep,
// 			Supporters:      getSourceList(cluster),
// 			SimilarNames:    getNameList(cluster),
// 			Confidence:      float32(len(cluster)) / float32(len(values)),
// 		})
// 	}
//
// 	return results
// }
