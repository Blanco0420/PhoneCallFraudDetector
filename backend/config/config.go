package config

import (
	"os"
)

var (
	IsDev bool = false
)

func LoadEnv() {
	// devVar := os.Getenv("PHRAUD__APP_ENV")
	if val, exists := os.LookupEnv("PHRAUD__APP_ENV"); exists {
		if val == "dev" {
			IsDev = true
		}
	}
	initLevenshtein()
}

func GetEnvVariable(variableToCheck string) (string, bool) {
	envVar := os.Getenv(variableToCheck)

	if envVar != "" {
		return envVar, true
	}

	return envVar, false

}
