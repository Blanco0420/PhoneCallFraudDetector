import React, { useEffect, useState } from "react"
import rawUseWebSocket, { ReadyState } from "react-use-websocket"
import { WebsocketMessage } from "./types"

const useWebSocket = (setCurrentData: Function) => {
  const [imageSrc, setImageSrc] = useState<string | undefined>(undefined)
  const [socketUrl, setSocketUrl] = useState<string>(`ws://${window.location.host}/ws`)
  const { sendMessage: rawSendMessage, lastMessage, readyState } = rawUseWebSocket(socketUrl)
  const connectionStatus = {
    [ReadyState.CONNECTING]: 'Connecting',
    [ReadyState.OPEN]: 'Open',
    [ReadyState.CLOSING]: 'Closing',
    [ReadyState.CLOSED]: 'Closed',
    [ReadyState.UNINSTANTIATED]: 'Uninstantiated',
  }[readyState];

  const sendMessage = (payload: WebsocketMessage) => {
    rawSendMessage(JSON.stringify(payload))

  }

  useEffect(() => {
    if (lastMessage !== null) {
      try {
        const data: WebsocketMessage = JSON.parse(lastMessage.data)
        switch (data.Command) {
          case "currentData":
            console.log(data.Payload.VitalInfo)
            const base64Img = data.Payload.VideoFeed
            setImageSrc(`data:image/jpeg;base64,${base64Img}`)
            setCurrentData(data.Payload.VitalInfo)
        }
      } catch (err) {
        console.error("failed to parse websocket message: ", err)
      }
    }
  }, [lastMessage])

  return { sendMessage, connectionStatus, imageSrc }

}

export default useWebSocket
