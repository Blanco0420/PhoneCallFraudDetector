package backendwebsocket

import (
	"sync"

	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
	"github.com/gorilla/websocket"
)

type clientType struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

type WebsocketCurrentData struct {
	VideoFeed string
	VitalInfo map[string]providers.VitalInfo // key = provider name
}

type WebsocketMessage struct {
	Command string
	Payload any
}
