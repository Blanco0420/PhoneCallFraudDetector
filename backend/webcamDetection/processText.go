package webcamdetection

import (
	"fmt"

	"github.com/otiai10/gosseract/v2"
)

func ProcessText(bytes []byte) (string, error) {

	client := gosseract.NewClient()
	client.SetPageSegMode(gosseract.PSM_SINGLE_LINE)
	client.SetLanguage("eng")
	// client.SetWhitelist("0123456789")
	defer client.Close()
	client.SetImageFromBytes(bytes)
	text, err := client.Text()
	if err != nil {
		return text, fmt.Errorf("error getting text: %w", err)
	}
	return text, nil
}
