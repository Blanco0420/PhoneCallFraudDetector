package providers

import (
	"net/http"
	"time"
)

func NewApiConfig(apiKey, baseUrl string) *APIConfig {
	return &APIConfig{
		APIKey:     apiKey,
		BaseUrl:    baseUrl,
		Timeout:    10 * time.Second,
		HttpClient: &http.Client{Timeout: 10 * time.Second},
		Headers:    make(map[string]string),
	}
}

func CreateVitalInfoChannel() chan VitalInfo {
	return make(chan VitalInfo)
}

func CloseVitalInfoChannel(channel chan VitalInfo) {
	close(channel)
}

func VitalInfoEqual(a, b VitalInfo) bool {
	if !stringPtrEqual(a.Name, b.Name) ||
		!stringPtrEqual(a.Industry, b.Industry) ||
		!stringPtrEqual(a.CompanyOverview, b.CompanyOverview) ||
		a.LineType != b.LineType ||
		a.OverallFraudScore != b.OverallFraudScore {
		return false
	}
	return a.FraudulentDetails == b.FraudulentDetails
}

func stringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

type Source interface {
	GetData(phoneNumber string) (NumberDetails, error)
	Close()
}
