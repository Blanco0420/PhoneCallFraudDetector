package providerdataprocessing

import (
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
)

type NumberRiskInput struct {
	SourceName  string
	GraphData   []providers.GraphData
	FraudScore  *int
	RecentAbuse *bool
	Comments    []providers.Comment
}

type FinalDisplayData struct {
	BusinessName         []ConfidenceResult
	BusinessNameSuffixes []string
	LineType             []ConfidenceResult
	Industry             []ConfidenceResult
	CompanyOverview      []ConfidenceResult
	FinalFraudScore      int
	FinalRecentAbuse     bool
}
