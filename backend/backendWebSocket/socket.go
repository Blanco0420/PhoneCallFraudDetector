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

type clientType struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if config.IsDev {
				return true
			} else {
				return false
			}
		},
	}
	clients      = make(map[*clientType]struct{})
	clientsMutex sync.RWMutex
)

func SendToClients(data WebsocketMessage) error {
	clientsMutex.Lock()
	for client := range clients {
		if err := client.sendMessage(data); err != nil {
			client.conn.Close()
			delete(clients, client)
		}
	}
	clientsMutex.Unlock()
	return nil
}

func (client *clientType) sendMessage(data any) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return client.conn.WriteJSON(data)
}

func sendFrames(cs *webcamdetection.CameraService) error {
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
	websocketMessage := WebsocketMessage{
		Command: "currentData",
		Payload: WebsocketCurrentData{
			VideoFeed: base64Img,
		},
	}
	return SendToClients(websocketMessage)
}

type WebsocketCurrentData struct {
	VideoFeed string
	VitalInfo providers.VitalInfo
}

type WebsocketMessage struct {
	Command string
	Payload interface{}
}

func wsHandler(websocketMessageChannel chan WebsocketMessage, cs *webcamdetection.CameraService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.Error().Err(err).Msg("failed to upgrade on websocket")
			return
		}
		logging.Info().Msg("Client connected")
		client := &clientType{
			conn:  ws,
			mutex: sync.Mutex{},
		}
		clientsMutex.Lock()
		clients[client] = struct{}{}
		clientsMutex.Unlock()

		done := make(chan struct{})

		go func() {
			defer close(done)
			for {
				messageType, data, err := ws.ReadMessage()
				if err != nil {
					logging.Error().Err(err).Msg("websocket read failed (client likely disconnected)")
					break
				}
				if messageType != websocket.TextMessage {
					logging.Warn().Msgf("Received unknown message type from websocket client: %d", messageType)
					continue
				}
				var websocketMessage WebsocketMessage
				if err := json.Unmarshal(data, &websocketMessage); err != nil {
					logging.Error().Err(err).Msg("Failed to unmarshal json data from client")
				}
				websocketMessageChannel <- websocketMessage
			}
		}()

		<-done
		ws.Close()
		clientsMutex.Lock()
		delete(clients, client)
		clientsMutex.Unlock()

	}
}

// func ping(w http.ResponseWriter, r *http.Request) {
// 	ws, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		logging.Error().Err(err).Msg("failed to upgrade websocket")
// 	}
// }

func SetupWebsocket(websocketMessageChannel chan WebsocketMessage, cs *webcamdetection.CameraService) error {
	go func() {

		for {
			if err := sendFrames(cs); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					break
				}
				logging.Error().Err(err).Msg("Failed to send frames to client")
				break
			}
			time.Sleep(33 * time.Millisecond)
		}
	}()
	http.HandleFunc("/ws", wsHandler(websocketMessageChannel, cs))
	// http.HandleFunc("/ping", ping)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		return err
	}
	return nil
}
