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

type Source interface {
	GetData(phoneNumber string) (NumberDetails, error)
	Close()
}
