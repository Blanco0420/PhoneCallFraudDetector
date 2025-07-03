package main

import (
	"PhoneNumberCheck/config"
	japanesetokenizing "PhoneNumberCheck/japaneseTokenizing"
	"PhoneNumberCheck/logging"
	"PhoneNumberCheck/profanityAnalyzing"
	providerdataprocessing "PhoneNumberCheck/providerDataProcessing"
	"PhoneNumberCheck/providers"
	"PhoneNumberCheck/providers/jpnumber"
	"PhoneNumberCheck/providers/telnavi"
	"PhoneNumberCheck/utils"
	"fmt"
	"slices"
	"sync"
	"time"
)

var (
	testNums = []string{
		// "08003007299",
		// "08007770319",
		// "0366360855",
		// "05031075729",
		// "0648642471",
		// "07091762683",
		"05031595686",
		// "05811521308",
		// "0648642471",
		// "0368935962",
		// "0120830068",
		// "0752317111",
		// "0356641888",
		// "05054822807",
		// "0661675628",
		// "08005009120",
		// "09077097477",
		// "08087569409",
		// "09034998875",
		// "08020178530",
	}
)

func printFinalDisplayData(data providerdataprocessing.FinalDisplayData) {
	fmt.Println("Final Display Data:")

	printConfidenceResults := func(title string, results []providerdataprocessing.ConfidenceResult) {
		fmt.Printf("%s:\n", title)
		if len(results) == 0 {
			fmt.Println("  (no data)")
			return
		}
		for _, res := range results {
			fmt.Printf("  Value: %s, Confidence: %.2f, Sources: %v\n", res.NormalizedValue, res.Confidence, res.Supporters)
		}
	}

	printConfidenceResults("Business Names", data.BusinessName)
	fmt.Printf("All suffixes: %v\n", data.BusinessNameSuffixes)
	printConfidenceResults("Line Types", data.LineType)
	printConfidenceResults("Industries", data.Industry)
	printConfidenceResults("Company Overviews", data.CompanyOverview)

	fmt.Printf("Final Fraud Score: %d\n", data.FinalFraudScore)
	fmt.Printf("Final Recent Abuse: %v\n", data.FinalRecentAbuse)
}

func buildFinalDisplayData(data map[string]providers.NumberDetails) (providerdataprocessing.FinalDisplayData, error) {
	var businessNames, allSuffixes, lineTypes, industries, businessOverviews []string
	var businessSources, lineTypeSources, industrySources, overviewSources []string
	var fraudScores []int
	var recentAbuseCount int
	var abuseSeen int

	for sourceName, details := range data {
		if details.VitalInfo.Name != "" {
			businessNames = append(businessNames, details.VitalInfo.Name)
			businessSources = append(businessSources, sourceName)

			suffixes := details.BusinessDetails.NameSuffixes
			if len(suffixes) > 0 {
				for _, suffix := range suffixes {
					if !slices.Contains(suffixes, suffix) {
						allSuffixes = append(suffixes, suffix)
					}
				}
			}
		}
		if details.VitalInfo.LineType != "" {
			lineTypes = append(lineTypes, string(details.VitalInfo.LineType))
			lineTypeSources = append(lineTypeSources, sourceName)
		}
		if details.VitalInfo.Industry != "" {
			industries = append(industries, details.VitalInfo.Industry)
			industrySources = append(industrySources, sourceName)
		}
		if details.VitalInfo.OverallFraudScore != 0 {
			fraudScores = append(fraudScores, details.VitalInfo.OverallFraudScore)
		}
		if details.VitalInfo.FraudulentDetails.RecentAbuse {
			//TODO: Maybe fix:
			// if not nil: abuseSeen++ ; if is true: recentAbuseCount++
			recentAbuseCount++
			abuseSeen++
		}
	}

	tokenizer, err := japanesetokenizing.Initialize()
	if err != nil {
		return providerdataprocessing.FinalDisplayData{}, err
	}
	return providerdataprocessing.FinalDisplayData{
		BusinessName:         providerdataprocessing.CalculateFieldConfidence(tokenizer, businessNames, businessSources),
		BusinessNameSuffixes: allSuffixes,
		LineType:             providerdataprocessing.CalculateFieldConfidence(tokenizer, lineTypes, lineTypeSources),
		Industry:             providerdataprocessing.CalculateFieldConfidence(tokenizer, industries, industrySources),
		CompanyOverview:      providerdataprocessing.CalculateFieldConfidence(tokenizer, businessOverviews, overviewSources),
		FinalFraudScore:      utils.AverageIntSlice(fraudScores),
		FinalRecentAbuse: func() bool {
			if abuseSeen == 0 {
				return false
			}
			return recentAbuseCount >= (abuseSeen / 2)
		}(),
	}, nil
}

func testingProviders(data *map[string]providers.NumberDetails, sources map[string]providers.Source) error {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// UI goroutine to refresh screen
	// go func() {
	// 	for {
	// 		// time.Sleep(500 * time.Millisecond)
	// 		// fmt.Print("\033[H\033[2J") // Clear screen
	// 		// fmt.Println("Live Source Output:")
	// 		// fmt.Println("====================")
	// 		// outputsMu.Lock()
	// 		// for _, name := range orderedSourceNames {
	// 		// 	if out, ok := outputs[name]; ok {
	// 		// 		fmt.Printf("--- %s ---\n%s\n\n", name, out)
	// 		// 	} else {
	// 		// 		fmt.Printf("--- %s ---\n(waiting for data...)\n\n", name)
	// 		// 	}
	// 		// }
	// 		// outputsMu.Unlock()
	// 	}
	// }()
	for localSourceName, localSource := range sources {
		wg.Add(1)

		go func(srcName string, src providers.Source) {
			//TODO: Actually do something with the channel
			defer wg.Done()
			for _, number := range testNums {
				fmt.Printf("[%s] calling getData\n", srcName)
				sourceData, err := src.GetData(number)
				if err != nil {
					panic(err)
				}
				mu.Lock()
				(*data)[srcName] = sourceData
				mu.Unlock()
			}
			fmt.Printf("[%s] finished GetData and closed channel\n", srcName)
		}(localSourceName, localSource)
	}

	wg.Wait()
	return nil
}

func startNumberProcessing() {
	logging.Info().Msg("Loading environment file")
	config.LoadEnv()

	logging.Info().Msg("Initalizing profanity lists")
	profanityAnalyzing.Initialize()
	// TODO: Send error here
	// jpNumberProvider := jpnumber.Initialize(driver)

	// numverify, err := numverify.Initialize()
	// if err != nil {
	// 	panic(err)
	// }

	jpNumber, err := jpnumber.Initialize()
	if err != nil {
		panic(err)
	}
	defer func() {
		if r := recover(); r != nil {
			jpNumber.Close()
			panic(r)
		}
	}()

	// ipqsSource, err := ipqualityscore.Initialize()
	// if err != nil {
	// 	panic(err)
	// }

	telnavi, err := telnavi.Initialize()
	if err != nil {
		panic(err)
	}
	defer func() {
		if r := recover(); r != nil {
			telnavi.Close()
			panic(r)
		}
	}()

	sources := map[string]providers.Source{
		"jpNumber": jpNumber,
		// ipqsSource,
		// numverify,
		"telnavi": telnavi,
	}

	data := map[string]providers.NumberDetails{}
	start := time.Now()
	err = testingProviders(&data, sources)
	if err != nil {
		panic(err)
	}
	_, err = buildFinalDisplayData(data)
	if err != nil {
		logging.Fatal().Err(err)
	}
	elapsed := time.Since(start)
	logging.Info().Msg(fmt.Sprintf("Finished. Time taken: %v", elapsed))
	// Optionally print results for each run:
	// printFinalDisplayData(finalData)
	// // for key, val := range data {
	// // 	json, err := json.MarshalIndent(val, "", "  ")
	// // 	if err != nil {
	// // 		panic(err)
	// // 	}
	// // }
	// localData, err := json.MarshalIndent(data, "", "  ")
	// if err != nil {
	// 	panic(err)
	// }
	// file, err := os.OpenFile("output.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	panic(err)
	// }
	// defer file.Close()
	// _, err = file.WriteString("\n")
	// if err != nil {
	// 	panic(err)
	// }
	// _, err = file.Write(localData)
	// if err != nil {
	// 	panic(err)
	// }

	// numberChan := make(chan string)
	// stopChan := make(chan struct{})
	// go webcamdetection.StartOCRScanner(numberChan, stopChan)

	// ipqsSource, err := ipqualityscore.Initialize()
	// if err != nil {
	// 	panic(err)
	// }

	// pref, exists := japaneseinfo.FindPrefectureByCityName("台東区", 1)
	// fmt.Println(pref, exists)

}

func main() {
	fmt.Println("heelo")
	time.Sleep(5 * time.Second)
	// webcamdetection.StartWebcamCapture()
	// cameraConfig := webcamdetection.StartCameraWithControls()
	// jsonData, err := json.MarshalIndent(cameraConfig, "", "  ")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(jsonData))
}
