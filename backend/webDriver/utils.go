package webdriver

import (
	"fmt"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
	"github.com/Blanco0420/Phone-Number-Check/backend/utils"

	"github.com/tebeka/selenium"
)

func GetTableInformation(d *WebDriverWrapper, tableBodyElement selenium.WebElement, tableKeyElementTagName string, tableValueElementTagName string) ([]providers.TableEntry, error) {
	var tableEntries []providers.TableEntry
	ignoredTableKeys := []string{"初回クチコミユーザー", "FAX番号", "市外局番", "市内局番", "加入者番号", "電話番号", "推定発信地域"}
	phoneNumberTableContainerRowElements, err := tableBodyElement.FindElements(selenium.ByCSSSelector, "tr")
	if err != nil {
		panic(fmt.Errorf("Could not get phone number info table rows: %v", err))
	}

	if tableKeyElementTagName == tableValueElementTagName {
		tableKeyElementTagName = tableKeyElementTagName + ":nth-child(1)"
		tableValueElementTagName = tableValueElementTagName + ":nth-child(2)"
	}

	for _, element := range phoneNumberTableContainerRowElements {
		key, err := d.GetInnerText(element, tableKeyElementTagName)
		if err != nil {
			continue
			//TODO: Fix this?
		}
		if slices.Contains(ignoredTableKeys, key) {
			continue
		}
		value, err := d.GetInnerText(element, tableValueElementTagName)
		if err != nil {
			return tableEntries, err
		}
		//Clean text
		key = utils.CleanText(key)
		value = utils.CleanText(value)

		tableEntries = append(tableEntries, providers.TableEntry{Key: key, Value: value, Element: element})
	}
	return tableEntries, nil
}

func getStatusCode(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func getFreePort() (int, error) {
	portMutex.Lock()
	defer portMutex.Unlock()
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return -1, err
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}

func retry(attempts int, sleep time.Duration, fn func() error) error {
	var lastErr error
	for i := 0; i < attempts; {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if i < attempts-1 {
			time.Sleep(sleep)
		}
	}
	return lastErr
}
