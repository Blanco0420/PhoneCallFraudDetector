package backendwebsocket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Blanco0420/Phone-Number-Check/backend/config"
	"github.com/Blanco0420/Phone-Number-Check/backend/logging"
	webcamdetection "github.com/Blanco0420/Phone-Number-Check/backend/webcamDetection"
	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if config.IsDev {
			return true
		} else {
			return false
		}
	},
}

func sendToClients(data string, ws *websocket.Conn) error {
	if err := ws.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
		return fmt.Errorf("error sending message to client/s: %v", err)
	}

	return nil
}

func sendFrames(ws *websocket.Conn, cs *webcamdetection.CameraService) error {
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
	if err := sendToClients(base64Img, ws); err != nil {
		return err
	}
	return nil
}

type websocketMessage struct {
	command string
	payload json.RawMessage
}

func handleCommand(messageType int, data []byte) error {
	return nil
}

func wsHandler(cs *webcamdetection.CameraService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.Error().Err(err).Msg("failed to upgrade on websocket")
			return
		}
		defer ws.Close()
		logging.Info().Msg("Client connected")

		go func() {
			for {
				messageType, data, err := ws.ReadMessage()
				if err != nil {
					logging.Error().Err(err).Msg("websocket read failed (client likely disconnected)")
					break
				}
				logging.Debug().Msgf("Message from client:\n%s", string(data))
			}
		}()

		for {
			// ws.WriteMessage(websocket.TextMessage, []byte("Yeet"))
			if err := sendFrames(ws, cs); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					break
				}
				logging.Error().Err(err).Msg("Failed to send frames to client")
				break
			}
			time.Sleep(33 * time.Millisecond)
		}
	}
}

// func ping(w http.ResponseWriter, r *http.Request) {
// 	ws, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		logging.Error().Err(err).Msg("failed to upgrade websocket")
// 	}
// }

func SetupWebsocket(cs *webcamdetection.CameraService) error {
	http.HandleFunc("/ws", wsHandler(cs))
	// http.HandleFunc("/ping", ping)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		return err
	}
	logging.Info().Msg("Websocket server started on :8080")
	return nil
}
