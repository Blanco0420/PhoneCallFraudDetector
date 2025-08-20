import React from "react"
import { CollatedVitalInfo } from "./Websocket/types"
type Props = {
  currentData: CollatedVitalInfo | undefined;
}

const DataDisplay = ({ currentData }: Props) => {
  if (!currentData || Object.keys(currentData).length === 0) {
    return null
  }

  return (
    <div>
      <h1>Data Display</h1>
      <pre>{JSON.stringify(currentData, null, 2)}</pre>
    </div>
  );
}

export default DataDisplay
