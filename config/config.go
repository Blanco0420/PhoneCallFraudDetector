package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func LoadEnv() {
	env := os.Getenv("APP_ENV")

	if env == "development" {
		if err := godotenv.Load(); err != nil {
			log.Warn().Msg("Failed to load .env file. Continuing without.")
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
