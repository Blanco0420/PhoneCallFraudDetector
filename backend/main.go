package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	backendwebsocket "github.com/Blanco0420/Phone-Number-Check/backend/backendWebSocket"
	"github.com/Blanco0420/Phone-Number-Check/backend/config"
	databasedriver "github.com/Blanco0420/Phone-Number-Check/backend/databaseDriver"
	japanesetokenizing "github.com/Blanco0420/Phone-Number-Check/backend/japaneseTokenizing"
	"github.com/Blanco0420/Phone-Number-Check/backend/logging"
	"github.com/Blanco0420/Phone-Number-Check/backend/profanityAnalyzing"
	providerdataprocessing "github.com/Blanco0420/Phone-Number-Check/backend/providerDataProcessing"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers/jpnumber"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers/telnavi"
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

type Services struct {
	CameraService    *webcamdetection.CameraService
	WebsocketChannel chan backendwebsocket.WebsocketMessage
	Sources          map[string]providers.Source
	DatabaseDriver   *databasedriver.DatabaseDriver
}

// initializeServices sets up all required services and providers
func initializeServices() (*Services, error) {
	logging.Info().Msg("Loading environment variables")
	config.LoadEnv()

	logging.Info().Msg("Starting camera service")
	cs, err := webcamdetection.NewCameraService(0)
	if err != nil {
		return nil, err
	}

	// go func() {
	// 	if err := backendapi.StartBackendApi(roiChan, cs); err != nil {
	// 		logging.Fatal().Err(err).Msg("Failed to start backend api service")
	// 		os.Exit(1)
	// 	}
	// }()
	logging.Info().Msg("Initating websocket")
	websocketMessageChannel := make(chan backendwebsocket.WebsocketMessage)
	go func() {
		err := backendwebsocket.SetupWebsocket(websocketMessageChannel, cs)
		if err != nil {
			logging.Fatal().Err(err).Msg("Failed to start websocket")
		}
	}()

	logging.Info().Msg("Initializing database")
	databaseDriver, err := databasedriver.InitializeDriver()
	if err != nil {
		return nil, err
	}

	logging.Info().Msg("Initalizing profanity lists")
	if err := profanityAnalyzing.Initialize(); err != nil {
		logging.Fatal().Err(err).Msg("Failed to initialize profanity lists")
		os.Exit(2)
	}
	// TODO: Send error here
	// jpNumberProvider := jpnumber.Initialize(driver)

	// numverify, err := numverify.Initialize()
	// if err != nil {
	// 	panic(err)
	// }

	jpNumber, err := jpnumber.Initialize()
	if err != nil {
		return nil, err
	}
	// Clean up jpNumber on panic
	go func() {
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
		return nil, err
	}
	go func() {
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

	return &Services{
		CameraService:    cs,
		WebsocketChannel: websocketMessageChannel,
		Sources:          sources,
		DatabaseDriver:   databaseDriver,
	}, nil
}

// monitorAndParseNumber continuously monitors the ROI and returns a valid number
// func monitorAndParseNumber(cs *webcamdetection.CameraService, roiChan chan webcamdetection.RoiData) (string, error) {
// 	for {
// 		roi := <-roiChan
// 		numberChan := make(chan struct {
// 			string
// 			error
// 		})
// 		go func() {
// 			for {

// 				num, err := cs.MonitorCamera(roi)
// 				if err != nil {
// 					if strings.Contains(err.Error(), "phone number supplied is not a number") {
// 						logging.Error().Err(err).Msgf("Failed to read phone number. Text: %s", num)
// 						continue
// 					} else {
// 						logging.Error().Err(err).Msgf("Error monitoring camera")
// 						continue // Try again
// 					}
// 				}
// 				if num != "" && num != "<nil>" {
// 					numberChan <- struct {
// 						string
// 						error
// 					}{num, nil}
// 				}
// 			}
// 		}()
// 		select {
// 		case res := <-numberChan:
// 			return res.string, res.error
// 		case <-time.After(6 * time.Second):
// 			return "", fmt.Errorf("timed out reading number")
// 		}
// 	}
// }

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
func mainLoop(services *Services) {
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
				logging.Info().Msg("Received 'start' command")
				select {
				case startChan <- msg.Payload:
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

		fmt.Println("Valid number detected:", num)

		data := make(map[string]providers.NumberDetails)
		start := time.Now()
		finalFraudScore, err := processNumber(num, &data, services.Sources)
		if err != nil {
			logging.Error().Err(err).Msg("Error processing number")
			continue
		}
		if err := services.DatabaseDriver.InsertNumberIntoDatabase(ctx, data, finalFraudScore); err != nil {
			fmt.Println(err)
		}

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
	services, err := initializeServices()
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
