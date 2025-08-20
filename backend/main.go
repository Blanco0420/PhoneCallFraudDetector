package main

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	japanesetokenizing "github.com/Blanco0420/Phone-Number-Check/backend/japaneseTokenizing"
	"github.com/Blanco0420/Phone-Number-Check/backend/logging"
	providerdataprocessing "github.com/Blanco0420/Phone-Number-Check/backend/providerDataProcessing"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
	"github.com/Blanco0420/Phone-Number-Check/backend/services"
	"github.com/Blanco0420/Phone-Number-Check/backend/utils"

	webcamdetection "github.com/Blanco0420/Phone-Number-Check/backend/webcamDetection"

	_ "net/http/pprof"
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

func buildFinalDisplayData(data *map[string]providers.NumberDetails) (providerdataprocessing.FinalDisplayData, error) {
	var businessNames, allSuffixes, lineTypes, industries, businessOverviews []string
	var businessSources, lineTypeSources, industrySources, overviewSources []string
	var fraudScores []int
	var recentAbuseCount int
	var abuseSeen int

	for sourceName, details := range *data {
		if details.VitalInfo.Name != nil {
			businessNames = append(businessNames, *details.VitalInfo.Name)
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
		if details.VitalInfo.Industry != nil {
			industries = append(industries, *details.VitalInfo.Industry)
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
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	// defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex

	for localSourceName, localSource := range sources {
		wg.Add(1)

		go func(srcName string, src providers.Source) {
			defer wg.Done()

			logging.Info().Msgf("[%s] starting search", srcName)

			//TODO: Actually do something with the channel
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
			result := <-resultChan
			if result.err != nil {
				logging.Error().Err(result.err).Msgf("[%s] error while searching for data", srcName)
				return
			}
			mu.Lock()
			(*data)[srcName] = result.data
			mu.Unlock()
			logging.Info().Msgf("[%s] finished searching for data", srcName)
		}(localSourceName, localSource)
	}

	wg.Wait()
	return nil
}

type numberResult struct {
	Number string
	Err    error
}

func monitorAndParseNumber(cs *webcamdetection.CameraService, roiData webcamdetection.RoiData) (string, error) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()

		numberChan := make(chan numberResult, 1)

		go func() {
			num, err := cs.MonitorCamera(ctx, roiData)
			if err != nil {
				logging.Error().Err(err).Msg("Error monitoring camera")
				return
			}
			if num != "" {
				numberChan <- numberResult{Number: num, Err: nil}
			}
		}()

		select {
		case res := <-numberChan:
			return res.Number, res.Err
		case <-ctx.Done():
			return "", fmt.Errorf("timed out reading number")
		}
	}
}

func processNumber(num string, data *map[string]providers.NumberDetails, sources map[string]providers.Source) (fraudScore int, err error) {

	if err = callProviders(num, data, sources); err != nil {
		return 0, fmt.Errorf("callProviders failed: %w", err)
	}

	printData, err := buildFinalDisplayData(data)
	if err != nil {
		return 0, fmt.Errorf("failed to build display data: %w", err)
	}

	printFinalDisplayData(printData)
	return printData.FinalFraudScore, nil
}

// processNumber calls providers, builds and prints results
// func processNumber(num string, sources map[string]providers.Source) error {
// 	data := map[string]providers.NumberDetails{}
// 	err := callProviders(num, &data, sources)
// 	if err != nil {
// 		return err
// 	}
// 	printData, err := buildFinalDisplayData(data)
// 	if err != nil {
// 		return err
// 	}
// 	printFinalDisplayData(printData)
// 	return nil
// }

// Add resource monitoring
func startResourceMonitor(interval time.Duration) {
	go func() {
		var m runtime.MemStats
		for {
			runtime.ReadMemStats(&m)
			numGoroutine := runtime.NumGoroutine()
			logging.Info().Msgf(
				"MEM: Alloc = %v MiB, TotalAlloc = %v MiB, Sys = %v MiB, NumGC = %v, Goroutines = %v",
				bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC, numGoroutine,
			)
			time.Sleep(interval)
		}
	}()
}

// mainLoop orchestrates the monitoring and processing in a loop
func mainLoop(services *services.Services) {
	ctx := context.Background()

	var (
		startChan = make(chan webcamdetection.RoiData, 1)
		stopChan  = make(chan struct{}, 1) // Channel to signal stop/pause
	)

	// Single goroutine that listens to WebSocket messages
	go func() {
		for msg := range services.WebsocketChannel {
			switch msg.Command {
			case "start":
				rawMap, ok := msg.Payload.(map[string]interface{})
				if !ok {
					logging.Error().Msgf("Error, invalid ROI data received: %v", msg.Payload)
					continue
				}
				rawJson, err := json.Marshal(rawMap)
				if err != nil {
					logging.Error().Err(err).Msg("failed to marshall data")
				}
				var roiData webcamdetection.RoiData
				if err := json.Unmarshal(rawJson, &roiData); err != nil {
					logging.Error().Err(err).Msg("failed to unmarshal websocket data")
					continue
				}

				select {
				case startChan <- roiData:
					logging.Info().Msg("Sent start payload to main loop.")
				default:
					logging.Warn().Msg("Start channel is busy or already running, skipping start command.")
				}
			case "stop":
				logging.Info().Msg("Received 'stop' command")
				select {
				case stopChan <- struct{}{}:
					logging.Info().Msg("Sent stop signal to main loop.")
				default:
					logging.Warn().Msg("Stop channel is busy, already in a stopped state or stop signal pending.")
				}
			default:
				logging.Error().Msgf("error on websocket command, unknown command: %s", msg.Command)
			}
		}
	}()

	// Variable to track the running state
	isRunning := false
	var currentPayload webcamdetection.RoiData // To store the last received payload
	var lastNumber string

	for {
		if !isRunning {
			logging.Info().Msg("Waiting for 'start' command from websocket to begin continuous monitoring...")
			// Wait here until we receive a payload from the 'start' command
			payload := <-startChan
			currentPayload = payload // Store the payload
			isRunning = true
			logging.Info().Msg("Received initial start payload. Starting continuous number monitoring...")
		}

		// Use a select with a default case to allow continuous processing
		// while still listening for stop commands.
		select {
		case <-stopChan:
			// If a stop command is received, set isRunning to false and pause
			isRunning = false
			logging.Info().Msg("Operation paused. Will wait for a 'start' command to resume continuous monitoring.")
			continue // Skip the rest of the loop iteration and go back to waiting for 'start'
		default:
			// If no stop command is received, continue with the monitoring process
			// This block only executes if isRunning is true and no stop signal is present.
			if !isRunning {
				// This should ideally not be reached if the `if !isRunning` block handles the initial wait.
				// Added as a safeguard.
				continue
			}
		}

		// This part only runs if `isRunning` is true and no `stop` command was received.
		num, err := monitorAndParseNumber(services.CameraService, currentPayload) // Use the stored payload
		if err != nil {
			if strings.Contains(err.Error(), "timed out reading number") {
				logging.Warn().Msg("Timed out waiting for a valid number, retrying...")
			} else {
				logging.Error().Err(err).Msg("Error in monitoring/parsing number")
			}
			// In a continuous loop, you might want to consider a small delay here
			// to prevent busy-looping on persistent errors.
			time.Sleep(500 * time.Millisecond) // Example: wait half a second before retrying
			continue
		}

		data := make(map[string]providers.NumberDetails)
		//TODO: Add logic to check if x amount of time has passed. If so, check the number again. (Maybe separate functions so that only the database is checked as it will be in there as fraud or not fraud.)
		if num == lastNumber {
			continue
		}
		start := time.Now()
		finalFraudScore, err := processNumber(num, &data, services.Sources)
		logging.Info().Msgf("Final fraud score: %s", finalFraudScore)
		// finalFraudScore, err := processNumber(num, &data, services.Sources)
		if err != nil {
			logging.Error().Err(err).Msg("Error processing number")
			continue
		}
		// if err := services.DatabaseDriver.InsertNumberIntoDatabase(ctx, data, finalFraudScore); err != nil {
		// 	logging.Error().Err(err).Msgf("failed to insert number %s into database", num)
		// }

		elapsed := time.Since(start)
		logging.Info().Msgf("Finished one cycle. Time taken: %v", elapsed)

		// Small delay to prevent the loop from consuming 100% CPU if processing is very fast
		time.Sleep(100 * time.Millisecond)
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func main() {
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	startResourceMonitor(10 * time.Second) // logs every 10 seconds
	services, err := services.InitializeServices()
	if err != nil {
		panic(err)
	}
	defer services.DatabaseDriver.Close()
	// testingDatabase(services.DatabaseDriver)
	mainLoop(services)
	// parsedAddress := parser.ParseAddress("神奈川県横浜市西区高島2514リバース横浜403")
	// for _, val := range parsedAddress {
	// 	fmt.Println(val)
	// }
}
