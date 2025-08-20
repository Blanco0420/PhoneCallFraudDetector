import React from 'react'
import './App.css'
import CroppingElement from './Components/CroppingElement'
import { Toaster } from 'react-hot-toast'
import { useState } from 'react'
import { CollatedVitalInfo, VitalInfo } from './Components/Websocket/types'
import useWebSocket from './Components/Websocket/useWebsocket'
import DataDisplay from './Components/DataDisplay'

function App() {
  const [currentData, setCurrentData] = useState<CollatedVitalInfo | undefined>()
  const { sendMessage, connectionStatus, imageSrc } = useWebSocket(setCurrentData)
  return (
    <>
      <Toaster />
      <CroppingElement imageSrc={imageSrc} sendMessage={sendMessage} />
      <DataDisplay currentData={currentData} />
    </>
  )
}

export default App
