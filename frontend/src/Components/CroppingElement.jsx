import axios from "axios"
import { useEffect, useRef, useState } from "react"
import ReactCrop from "react-image-crop"
import 'react-image-crop/dist/ReactCrop.css'
const CroppingElement = () => {
  const imageRef = useRef(null)
  const [imageSrc, setImageSrc] = useState(null)
  const [crop, setCrop] = useState()

  const api = axios.create({
    baseURL: "/api",
    timeout: 10000
  })

  api.interceptors.request.use(
    config => {
      return config;
    },
    error => Promise.reject(error)
  )
  api.interceptors.response.use(
    config => {
      return config
    },
    error => {
      if (!error.response) {
        console.error("Failed to communicate with backend server")
      }
      const status = error.response.status
      switch (status) {
        case 404:
          console.error("Error, content not found")
          break
        default:
          console.error("Unexpected error: ", error.message)
      }

      return Promise.reject(error)
    })

  useEffect(() => {
    const image = new Image()
    image.src = imageSrc
    console.log(image)
    console.log(crop)
  }, [crop])
  const postRoiCrop = async () => {
    const image = imageRef.current
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
    try {
      const res = await axios.post("/api/setROIData", data)
      console.log("roi sent: ", res.data)
    }
    catch (e) {
      console.error("Error sending crop data: ", e)
    }
  }
  const fetchImage = async () => {
    try {
      const res = await axios.get("/api/getCurrentImage")
      if (!res.data || !res.data.image) {
        throw new Error("No data or image data in response")
      }
      setImageSrc(`data:image/jpeg;base64,${res.data.image}`)
    }
    catch (error) {
      console.error("Error getting snapshot: ", error.response)
    }
  }
  return (
    <div style={{ border: '1px solid #ccc', borderRadius: 8, padding: 24, maxWidth: 1200, margin: '32px auto', background: '#fafbfc', boxShadow: '0 2px 8px rgba(0,0,0,0.06)' }}>
      <div style={{ display: 'flex', gap: 12, marginBottom: 18 }}>
        <button style={{ padding: '8px 16px', borderRadius: 4, border: '1px solid #888', background: '#f0f0f0', cursor: 'pointer' }} onClick={fetchImage}>Get snapshot</button>
        <button style={{ padding: '8px 16px', borderRadius: 4, border: '1px solid #888', background: '#f0f0f0', cursor: 'pointer' }} onClick={postRoiCrop}>Send data</button>
      </div>
      <div style={{ border: '1px solid #eee', borderRadius: 6, padding: 8, background: '#fff' }}>
        <ReactCrop crop={crop} onChange={c => setCrop(c)}>
          <img ref={imageRef} style={{ height: 640, width: "100%", objectFit: 'contain', borderRadius: 4 }} src={imageSrc} />
        </ReactCrop >
      </div>
      {/* <Cropper key={imageSrc} ref={cropperRef} viewMode={1} src={imageSrc} crop={onCrop} style={{ height: 400, width: "100%" }} guides={true} /> */}
    </div>
  )
}

export default CroppingElement
