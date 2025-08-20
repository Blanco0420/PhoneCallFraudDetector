package backendwebsocket

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/Blanco0420/Phone-Number-Check/backend/config"
	"github.com/Blanco0420/Phone-Number-Check/backend/logging"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
	webcamdetection "github.com/Blanco0420/Phone-Number-Check/backend/webcamDetection"
	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return config.IsDev // allow all in dev, block in prod
		},
	}
	clients      = make(map[*clientType]bool)
	clientsMutex sync.RWMutex
	sharedData   = WebsocketCurrentData{
		VideoFeed: "",
		VitalInfo: make(map[string]providers.VitalInfo),
	}
	sharedDataMutex sync.Mutex
)

// sendMessage is thread-safe sending for a single client
func (client *clientType) sendMessage(data any) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return client.conn.WriteJSON(data)
}

// sendStateToClients broadcasts the full shared state to all connected clients
func sendStateToClients() {
	sharedDataMutex.Lock()
	message := WebsocketMessage{
		Command: "currentData",
		Payload: sharedData,
	}
	sharedDataMutex.Unlock()

	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for client := range clients {
		if err := client.sendMessage(message); err != nil {
			client.conn.Close()
			delete(clients, client)
		}
	}
}

// SendVideoFeed grabs a new webcam frame and broadcasts it
func sendVideoFeed(cs *webcamdetection.CameraService) error {
	frame := gocv.NewMat()
	if err := cs.GetFrame(&frame); err != nil {
		return err
	}
	buf, err := gocv.IMEncode(gocv.JPEGFileExt, frame)
	if err != nil {
		return err
	}
	frame.Close()

	base64Img := base64.StdEncoding.EncodeToString(buf.GetBytes())
	buf.Close()

	sharedDataMutex.Lock()
	sharedData.VideoFeed = base64Img
	sharedDataMutex.Unlock()

	sendStateToClients()
	return nil
}

// StartProviderWatchers listens to provider channels and sends updates when data changes
func StartProviderWatchers(providerChannels map[string]<-chan providers.VitalInfo) {
	for providerName, ch := range providerChannels {
		go func(name string, c <-chan providers.VitalInfo) {
			for info := range c {
				sharedDataMutex.Lock()
				old := sharedData.VitalInfo[name]
				if !providers.VitalInfoEqual(old, info) {
					if sharedData.VitalInfo == nil {
						sharedData.VitalInfo = make(map[string]providers.VitalInfo)
					}
					sharedData.VitalInfo[name] = info
					sharedDataMutex.Unlock()
					sendStateToClients()
				} else {
					sharedDataMutex.Unlock()
				}
			}
		}(providerName, ch)
	}
}

// wsHandler handles a single WebSocket connection
func wsHandler(websocketMessageChannel chan WebsocketMessage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.Error().Err(err).Msg("failed to upgrade on websocket")
			return
		}
		logging.Info().Msg("Client connected")
		client := &clientType{conn: ws}

		clientsMutex.Lock()
		clients[client] = true
		clientsMutex.Unlock()

		defer func() {
			ws.Close()
			clientsMutex.Lock()
			delete(clients, client)
			clientsMutex.Unlock()
			logging.Info().Msg("Client disconnected")
		}()

		for {
			messageType, data, err := ws.ReadMessage()
			if err != nil {
				logging.Error().Err(err).Msg("websocket read failed (client likely disconnected)")
				return
			}
			if messageType != websocket.TextMessage {
				logging.Warn().Msgf("Received unknown message type from websocket client: %d", messageType)
				continue
			}
			var msg WebsocketMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				logging.Error().Err(err).Msg("Failed to unmarshal json data from client")
				continue
			}
			websocketMessageChannel <- msg
		}
	}
}

// SetupWebsocket starts everything
func SetupWebsocket(
	websocketMessageChannel chan WebsocketMessage,
	cs *webcamdetection.CameraService,
	providerChannels map[string]<-chan providers.VitalInfo,
) error {
	// Start provider watchers
	StartProviderWatchers(providerChannels)

	// Start webcam feed loop
	go func() {
		for {
			if err := sendVideoFeed(cs); err != nil {
				logging.Error().Err(err).Msg("Failed to send video feed")
				break
			}
			time.Sleep(33 * time.Millisecond) // ~30fps
		}
	}()

	http.HandleFunc("/ws", wsHandler(websocketMessageChannel))
	return http.ListenAndServe(":8080", nil)
}
