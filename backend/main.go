package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/Blanco0420/Phone-Number-Check/backend/config"
	japanesetokenizing "github.com/Blanco0420/Phone-Number-Check/backend/japaneseTokenizing"
	"github.com/Blanco0420/Phone-Number-Check/backend/logging"
	"github.com/Blanco0420/Phone-Number-Check/backend/profanityAnalyzing"
	providerdataprocessing "github.com/Blanco0420/Phone-Number-Check/backend/providerDataProcessing"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers/jpnumber"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers/telnavi"
	"github.com/Blanco0420/Phone-Number-Check/backend/utils"

	backendapi "github.com/Blanco0420/Phone-Number-Check/backend/backendApi"
	webcamdetection "github.com/Blanco0420/Phone-Number-Check/backend/webcamDetection"
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

func callProviders(number string, data *map[string]providers.NumberDetails, sources map[string]providers.Source) error {
	// Add a timeout context to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

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
			fmt.Printf("[%s] calling getData\n", srcName)

			// Create a channel to receive the result
			resultChan := make(chan struct {
				data providers.NumberDetails
				err  error
			}, 1)

			// Run the provider in a goroutine
			go func() {
				sourceData, err := src.GetData(number)
				resultChan <- struct {
					data providers.NumberDetails
					err  error
				}{sourceData, err}
			}()

			// Wait for result or timeout
			select {
			case result := <-resultChan:
				if result.err != nil {
					fmt.Printf("[%s] error: %v\n", srcName, result.err)
					return
				}
				mu.Lock()
				(*data)[srcName] = result.data
				mu.Unlock()
				fmt.Printf("[%s] finished GetData and closed channel\n", srcName)
			case <-ctx.Done():
				fmt.Printf("[%s] timeout after 5 minutes\n", srcName)
				return
			}
		}(localSourceName, localSource)
	}

	wg.Wait()
	return nil
}

// func startNumberProcessing(number) {
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

// }

func roiMonitorLoop(cs *webcamdetection.CameraService, numberChan chan string, roiChan chan webcamdetection.RoiData) error {
	for {
		roi := <-roiChan
		num, err := cs.MonitorCamera(roi)
		if err != nil {
			return err
		}
		numberChan <- num
		break
	}
	return nil
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	logging.Info().Msg("Starting camera service")
	cs, err := webcamdetection.NewCameraService(0)
	if err != nil {
		panic(err)
	}
	roiChan := make(chan webcamdetection.RoiData, 1)
	numberChan := make(chan string)
	go func() {
		if err := backendapi.StartBackendApi(roiChan, cs); err != nil {
			panic(err)
		}
	}()

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
	go roiMonitorLoop(cs, numberChan, roiChan)
	var num string
	for {
		num = <-numberChan
		if num != "" && num != "<nil>" {
			break
		}
	}

	err = callProviders(num, &data, sources)
	if err != nil {
		panic(err)
	}
	fmt.Println("Got here")
	printData, err := buildFinalDisplayData(data)
	if err != nil {
		logging.Fatal().Err(err)
	}
	fmt.Println("Got here too")
	printFinalDisplayData(printData)
	elapsed := time.Since(start)
	logging.Info().Msg(fmt.Sprintf("Finished. Time taken: %v", elapsed))
	// cameraConfig := webcamdetection.StartCameraWithControls()
	// jsonData, err := json.MarshalIndent(cameraConfig, "", "  ")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(jsonData))
}
