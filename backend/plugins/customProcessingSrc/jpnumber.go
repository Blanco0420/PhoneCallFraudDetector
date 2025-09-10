package customproviderprocessing

import (
	"errors"
	"fmt"
	"path"

	sourceconfig "github.com/Blanco0420/Phone-Number-Check/backend/plugins"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
)

type JpNumberProcessor struct{}

func (j *JpNumberProcessor) ProcessNumberPageSlug(sourceConfig *sourceconfig.SourceConfig, numberDetails *providers.NumberDetails) (processedSlug string, err error) {
	var slug string
	if numberDetails.VitalInfo.LineType == nil {
		return "", errors.New("line type is nil")
	}
	switch *numberDetails.VitalInfo.LineType {
	case providers.LineTypeMobile:
		slug = "mobile"
	case providers.LineTypeTollFree:
		slug = "freedial"
	case providers.LineTypeVOIP:
		slug = "ipphone"
	}
	if slug != "" {
		slug += "/"
	} else {
		slug = "/"
	}

	if numberDetails.Number == nil {
		return "", errors.New("phone number is nil")
	}
	numberParts, err := sourceConfig.GetNumberParts(numberDetails.Number)
	if err != nil {
		return processedSlug, err
	}

	return fmt.Sprintf("%snumberinfo_%s_%s_%s.html", slug, numberParts[0], numberParts[1], numberParts[2]), nil
}

func (j *JpNumberProcessor) ProcessUrl(sourceConfig *sourceconfig.SourceConfig, numberDetails *providers.NumberDetails) (string, error) {
	if sourceConfig == nil {
		return "", errors.New("sourceConfig is nil")
	}
	if numberDetails == nil {
		return "", errors.New("numberDetails is nil")
	}
	slug, err := j.ProcessNumberPageSlug(sourceConfig, numberDetails)
	if err != nil {
		return "", err
	}
	return path.Join(sourceConfig.BaseUrl, slug), nil
}

func init() {
	sourceconfig.RegisterProcessor("jpnumber", &JpNumberProcessor{})
}
