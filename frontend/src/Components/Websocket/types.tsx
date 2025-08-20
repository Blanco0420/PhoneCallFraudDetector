

export type RoiData = {
  x: number;
  y: number;
  width: number;
  height: number;
}


type FraudulentDetails = {
  fraudScore: number;
  recentAbuse: boolean;
}

export type CollatedVitalInfo = Record<string, VitalInfo>;

export type VitalInfo = {
  name: string;
  industry: string;
  companyOverview: string;
  lineType: string;
  fraudulentDetails: FraudulentDetails;
}

export type WebsocketCurrentData = {
  VideoFeed: string;
  VitalInfo: CollatedVitalInfo | undefined;
}

export type WebsocketMessage =
  | { Command: "start"; Payload: RoiData; }
  | { Command: "stop"; Payload: null; }
  | { Command: "currentData"; Payload: WebsocketCurrentData; }
