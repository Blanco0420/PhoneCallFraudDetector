package plugins

import "github.com/Blanco0420/Phone-Number-Check/backend/providers"

type NumberPageConfig struct {
	SlugTemplate string
	Processor    string
}

type SourceConfig struct {
	Enabled                  bool     `json:"enabled"`
	Name                     string   `json:"name"`
	BaseUrl                  string   `json:"baseUrl"`
	SupportedCountryCodes    []string `json:"supportedCountryCodes"`
	NumberPageSlugProcessing string   `json:"numberPageSlugProcessing"`
	NumberPageSlug           string   `json:"numberPageSlug"`
}

type PluginProcessor interface {
	ProcessUrl(sourceConfig *SourceConfig, numberDetails *providers.NumberDetails) (string, error)
}
