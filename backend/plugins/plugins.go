package plugins

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/Blanco0420/Phone-Number-Check/backend/logging"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
)

var processors = make(map[string]PluginProcessor)

var sourceConfigDir string

func init() {
	sourceConfigDir = os.Getenv("PHRAUD__SOURCE_CONFIG_DIR")
	if sourceConfigDir == "" {
		currentPath, err := os.Getwd()
		if err != nil {
			logging.Fatal().Err(err).Msgf("failed to get current working directory")
		}
		sourceConfigDir = path.Join(currentPath, "sourceConfig", "configFiles")
	}
}

func RegisterProcessor(name string, p PluginProcessor) {
	processors[name] = p
}

func GetProcessor(name string) (PluginProcessor, error) {
	p, ok := processors[name]
	if !ok {
		return nil, errors.New("custom processor not found")
	}
	return p, nil
}

func SetupSources() (map[string]SourceConfig, error) {
	var sourceConfigs = make(map[string]SourceConfig)
	err := filepath.Walk(sourceConfigDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".json" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			fileReadError := errors.New("failed to read file")
			return errors.Join(fileReadError, err)
		}

		var sourceConfig SourceConfig
		if err := json.Unmarshal(data, &sourceConfig); err != nil {
			return err
		}

		sourceConfigs[sourceConfig.Name] = sourceConfig

		return nil
	})
	if err != nil {
		return nil, err
	}
	return sourceConfigs, nil
}

func LoadCustomProcessingBinFiles() (map[string][]byte, error) {
	customProcessingBinDir := path.Join(sourceConfigDir, "..", "customProcessingBin")

	// Check if directory exists
	if _, err := os.Stat(customProcessingBinDir); os.IsNotExist(err) {
		return make(map[string][]byte), nil // Return empty map if directory doesn't exist
	}

	var files = make(map[string][]byte)
	err := filepath.Walk(customProcessingBinDir, func(filePath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Read file contents
		data, err := os.ReadFile(filePath)
		if err != nil {
			fileReadError := errors.New("failed to read custom processing bin file")
			return errors.Join(fileReadError, err)
		}

		// Get just the filename without the full path
		filename := filepath.Base(filePath)
		files[filename] = data

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func TestProcessors() error {

	providerName := "jpnumber"

	jpnumberProcessor, err := GetProcessor(providerName)
	if err != nil {
		return err
	}

	sources, err := SetupSources()
	if err != nil {
		return err
	}

	strPtr := func(s string) *string {
		return &s
	}

	testNumberDetails := providers.NumberDetails{
		Number: strPtr("07091762683"),
	}
	lt := providers.LineTypeMobile
	testNumberDetails.VitalInfo.LineType = &lt

	currentSourceConfig := sources[providerName]

	currentSourceUrl, err := jpnumberProcessor.ProcessUrl(&currentSourceConfig, &testNumberDetails)
	if err != nil {
		return err
	}
	logging.Debug().Msgf("%s url: %s", providerName, currentSourceUrl)

	return nil
}
