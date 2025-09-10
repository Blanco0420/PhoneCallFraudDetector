package plugins

import (
	"errors"
	"strings"

	"github.com/ttacon/libphonenumber"
)

func (source *SourceConfig) GetNumberParts(phoneNumber *string) ([]string, error) {
	if phoneNumber == nil {
		return nil, errors.New("phoneNumber is nil")
	}
	parsedNumber, err := libphonenumber.Parse(*phoneNumber, "JP")
	if err != nil {
		return []string{}, err
	}
	return strings.Split(libphonenumber.Format(parsedNumber, libphonenumber.NATIONAL), "-"), nil
}
