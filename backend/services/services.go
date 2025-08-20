package services

import (
	"os"

	backendwebsocket "github.com/Blanco0420/Phone-Number-Check/backend/backendWebSocket"
	"github.com/Blanco0420/Phone-Number-Check/backend/config"
	databasedriver "github.com/Blanco0420/Phone-Number-Check/backend/databaseDriver"
	"github.com/Blanco0420/Phone-Number-Check/backend/logging"
	"github.com/Blanco0420/Phone-Number-Check/backend/profanityAnalyzing"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers/jpnumber"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers/telnavi"
	webcamdetection "github.com/Blanco0420/Phone-Number-Check/backend/webcamDetection"
)

type Services struct {
	CameraService             *webcamdetection.CameraService
	WebsocketChannel          chan backendwebsocket.WebsocketMessage
	Sources                   map[string]providers.Source
	DatabaseDriver            *databasedriver.DatabaseDriver
	CollatedVitalInfoChannels map[string]<-chan providers.VitalInfo
}

func InitializeServices() (*Services, error) {
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
	collatedProviderChannels := make(map[string]<-chan providers.VitalInfo)

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

	collatedProviderChannels["jpnumber"] = jpNumber.VitalInfoChannel

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
	collatedProviderChannels["telnavi"] = telnavi.VitalInfoChannel
	sources := map[string]providers.Source{
		"jpNumber": jpNumber,
		// ipqsSource,
		// numverify,
		"telnavi": telnavi,
	}

	websocketMessageChannel := make(chan backendwebsocket.WebsocketMessage)
	go func() {
		err := backendwebsocket.SetupWebsocket(websocketMessageChannel, cs, collatedProviderChannels)
		if err != nil {
			logging.Fatal().Err(err).Msg("Failed to start websocket")
		}
	}()
	return &Services{
		CameraService:             cs,
		WebsocketChannel:          websocketMessageChannel,
		Sources:                   sources,
		DatabaseDriver:            databaseDriver,
		CollatedVitalInfoChannels: collatedProviderChannels,
	}, nil
}
