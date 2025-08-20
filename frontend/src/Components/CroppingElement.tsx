import React, { useRef, useState } from "react"
import { useEffect } from "react"
import ReactCrop, { PixelCrop } from "react-image-crop";
import 'react-image-crop/dist/ReactCrop.css'
import { WebsocketMessage, RoiData } from "./Websocket/types";
import toast from "react-hot-toast";

type Props = {
  imageSrc: string | undefined;
  sendMessage: Function;
}

const CroppingElement = ({ imageSrc, sendMessage }: Props) => {
  const [crop, setCrop] = useState<PixelCrop>()
  const imageRef = useRef<HTMLImageElement | null>(null)


  const pauseSystem = () => {
    const payload: WebsocketMessage = {
      Command: "stop",
      Payload: null
    }
    sendMessage(payload)
  }

  const sendCropData = () => {
    const image = imageRef.current;
    if (!image || !crop) {
      toast.error("No crop data selected on video. Please select the RoI (Region of Interest) and submit again.")
      return
    }
    const scaleX = image.naturalWidth / image.clientWidth
    const scaleY = image.naturalHeight / image.clientHeight
    const data: RoiData = {
      x: crop.x * scaleX,
      y: crop.y * scaleY,
      width: crop.width * scaleX,
      height: crop.height * scaleY
    }
    const payload: WebsocketMessage = {
      Command: "start",
      Payload: data
    }
    sendMessage(payload)
  }

  return (
    <div style={{ border: '1px solid #ccc', borderRadius: 8, padding: 24, maxWidth: 1200, margin: '32px auto', background: '#fafbfc', boxShadow: '0 2px 8px rgba(0,0,0,0.06)' }}>
      <div style={{ display: 'flex', gap: 12, marginBottom: 18 }}>
        <button style={{ padding: '8px 16px', borderRadius: 4, border: '1px solid #888', background: '#f0f0f0', cursor: 'pointer' }} onClick={sendCropData}>Send data</button>
        <button style={{ padding: '8px 16px', borderRadius: 4, border: '1px solid #888', background: '#f0f0f0', cursor: 'pointer' }} onClick={pauseSystem}>Pause system</button>
      </div>
      <div style={{ border: '1px solid #eee', borderRadius: 6, padding: 8, background: '#fff' }}>
        {imageSrc &&
          <ReactCrop crop={crop} onChange={c => setCrop(c)}>
            <img ref={imageRef} style={{ height: 640, width: "100%", objectFit: 'contain', borderRadius: 4 }} src={imageSrc} />
          </ReactCrop>
        }
      </div>
    </div>
  )
}

export default CroppingElement
