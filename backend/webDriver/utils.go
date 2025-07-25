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
		panic(fmt.Errorf("could not get phone number info table rows: %v", err))
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
		if slices.Contains(ignoredTableKeys, *key) {
			continue
		}
		value, err := d.GetInnerText(element, tableValueElementTagName)
		if err != nil {
			return tableEntries, err
		}
		//Clean text
		*key = utils.CleanText(*key)
		*value = utils.CleanText(*value)

		tableEntries = append(tableEntries, providers.TableEntry{Key: *key, Value: *value, Element: element})
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
	for i := 0; i < attempts; i++ {
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

// BatchExtractComments extracts all comment date and text pairs from a container element using a single JavaScript execution.
// commentSelector: CSS selector for each comment element (e.g., '#thread' or 'div.frame-728-gray-l')
// dateSelector: CSS selector for the date element, relative to each comment element
// textSelector: CSS selector for the text element, relative to each comment element
// Returns a slice of maps with keys 'date' and 'text'.
func BatchExtractComments(driver *WebDriverWrapper, container selenium.WebElement, commentSelector, dateSelector, textSelector string) ([]map[string]string, error) {
	script := `
	  const container = arguments[0];
	  const commentSelector = arguments[1];
	  const dateSelector = arguments[2];
	  const textSelector = arguments[3];
	  const results = [];
	  const comments = container.querySelectorAll(commentSelector);
	  for (let i = 0; i < comments.length; i++) {
	    const commentElem = comments[i];
	    const dateElem = commentElem.querySelector(dateSelector);
	    const textElem = commentElem.querySelector(textSelector);
	    results.push({
	      date: dateElem ? dateElem.textContent.trim() : '',
	      text: textElem ? textElem.textContent.trim() : ''
	    });
	  }
	  return results;
	`
	res, err := driver.driver.ExecuteScript(script, []interface{}{container, commentSelector, dateSelector, textSelector})
	if err != nil {
		return nil, err
	}
	// The result is []interface{} of map[string]interface{}; convert to []map[string]string
	out := []map[string]string{}
	if arr, ok := res.([]interface{}); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				entry := map[string]string{}
				for k, v := range m {
					if str, ok := v.(string); ok {
						entry[k] = str
					}
				}
				out = append(out, entry)
			}
		}
	}
	return out, nil
}
