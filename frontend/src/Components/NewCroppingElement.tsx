import React, { useRef, useState } from "react"
import { useEffect } from "react"
import ReactCrop, { PixelCrop } from "react-image-crop";
import useWebSocket from "react-use-websocket";

type WebsocketMessage = {
  command: string;
  payload: object | undefined;
}

const NewCroppingElement = () => {
  const [imageSrc, setImageSrc] = useState<string | undefined>(undefined)
  const [crop, setCrop] = useState<PixelCrop>()
  const imageRef = useRef<HTMLImageElement | null>(null)

  const [socketUrl, setSocketUrl] = useState<string>(`ws://${window.location.host}/ws`)
  const { sendMessage, lastMessage, readyState } = useWebSocket(socketUrl)

  useEffect(() => {
    if (lastMessage !== null) {
      const base64Img = lastMessage.data
      setImageSrc(`data:image/jpeg;base64,${base64Img}`)
    }
  }, [lastMessage])

  const sendCropData = () => {
    const image = imageRef.current;
        if (!image || !crop) {
      console.log("No crop or no image")
      return
    }
    const scaleX = image.naturalWidth / image.clientWidth
    const scaleY = image.naturalHeight / image.clientHeight
    const data = {
      x: crop.x * scaleX,
      y: crop.y * scaleY,
      width: crop.width * scaleX,
      height: crop.height * scaleY
    }
    const payload :WebsocketMessage = {
      command: "start",
      payload: data
    }
    const payloadString = JSON.stringify(payload)
    console.log(payloadString)
    console.log(crop)
    sendMessage(payloadString)
  }

  return (
    <div style={{ border: '1px solid #ccc', borderRadius: 8, padding: 24, maxWidth: 1200, margin: '32px auto', background: '#fafbfc', boxShadow: '0 2px 8px rgba(0,0,0,0.06)' }}>
      <div style={{ display: 'flex', gap: 12, marginBottom: 18 }}>
        <button style={{ padding: '8px 16px', borderRadius: 4, border: '1px solid #888', background: '#f0f0f0', cursor: 'pointer' }} onClick={sendCropData}>Send data</button>
      </div>
      <div style={{ border: '1px solid #eee', borderRadius: 6, padding: 8, background: '#fff' }}>
        {imageSrc &&
          <ReactCrop crop={crop} onChange={c => setCrop(c)}>
            <img ref={imageRef} style={{ height: 640, width: "100%", objectFit: 'contain', borderRadius: 4 }} src={imageSrc} />
          </ReactCrop>
        }
      </div>
      {/* <Cropper key={imageSrc} ref={cropperRef} viewMode={1} src={imageSrc} crop={onCrop} style={{ height: 400, width: "100%" }} guides={true} /> */}
    </div>
  )
}

export default NewCroppingElement
