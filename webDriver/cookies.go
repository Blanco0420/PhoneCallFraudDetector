package webdriver

import (
	"PhoneNumberCheck/utils"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tebeka/selenium"
)

type WebScrapingProvider string

const (
	TelnaviWebScrapingProvider  WebScrapingProvider = "telnavi"
	JpNumberWebScrapingProvider WebScrapingProvider = "jpnumber"
)

func (d *WebDriverWrapper) LoadCookies(provider WebScrapingProvider) error {
	filePath := fmt.Sprintf(".cookies-%s.json", provider)
	if exists := utils.CheckIfFileExists(filePath); !exists {
		return fmt.Errorf("no cookies")
	}
	cookieFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer cookieFile.Close()

	var cookies []*selenium.Cookie
	decoder := json.NewDecoder(cookieFile)
	if err := decoder.Decode(&cookies); err != nil {
		return err
	}

	for _, cookie := range cookies {
		err := d.driver.AddCookie(cookie)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *WebDriverWrapper) SaveCookies(provider WebScrapingProvider) error {
	cookies, err := d.driver.GetCookies()
	if err != nil {
		return err
	}

	cookieFile, err := os.Create(fmt.Sprintf(".cookies-%s.json", provider))
	if err != nil {
		return err
	}
	defer cookieFile.Close()

	encoder := json.NewEncoder(cookieFile)
	if err := encoder.Encode(cookies); err != nil {
		return err
	}

	return nil
}
