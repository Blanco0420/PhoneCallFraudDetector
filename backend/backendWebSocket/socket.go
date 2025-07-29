package backendwebsocket

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/Blanco0420/Phone-Number-Check/backend/logging"
	webcamdetection "github.com/Blanco0420/Phone-Number-Check/backend/webcamDetection"
	"github.com/gorilla/websocket"
	"gocv.io/x/gocv"
)

var upgrader = websocket.Upgrader{}

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
	sendToClients(base64Img, ws)
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

		for {
			if err := sendFrames(ws, cs); err != nil {
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
	http.HandleFunc("/ping", ping)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		return err
	}
	logging.Info().Msg("Websocket server started on :8080")
	return nil
}
