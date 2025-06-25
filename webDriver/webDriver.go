package webdriver

import (
	"PhoneNumberCheck/logging"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

type WebDriverWrapper struct {
	driver  selenium.WebDriver
	service *selenium.Service
}

var portMutex sync.Mutex

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
	var err error
	for range attempts {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(sleep)
	}
	return err
}

func InitializeDriver(providerName WebScrapingProvider) (*WebDriverWrapper, error) {
	port, err := getFreePort()
	if err != nil {
		return &WebDriverWrapper{}, err
	}
	service, err := selenium.NewGeckoDriverService("geckodriver", port)
	if err != nil {
		return &WebDriverWrapper{}, fmt.Errorf("error starting geckodriver service: %v", err)
	}

	profilePath, err := createProfile(providerName)
	if err != nil {
		return &WebDriverWrapper{}, err
	}

	caps := selenium.Capabilities{
		"browserName": "firefox",
	}
	firefoxCaps := firefox.Capabilities{
		Args: []string{
			// "--headless",
			"--profile", profilePath,
			"--noremote",
		},
		Prefs: map[string]any{
			"general.useragent.override": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
		},
	}
	caps.AddFirefox(firefoxCaps)
	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		return &WebDriverWrapper{}, fmt.Errorf("error connecting to remote server: %v", err)
	}

	return &WebDriverWrapper{
		driver:  driver,
		service: service,
	}, nil
}

func (w *WebDriverWrapper) GotoUrl(url string) error {
	status, err := getStatusCode(url)
	if err != nil {
		return err
	} else if status != 200 {
		fmt.Println("not 200")
		fmt.Println(status)
		switch status {
		case 403:
			logging.Warn().Msg(fmt.Sprintf("rate limited. Waiting until page is ready again"))
			time.Sleep(60 * time.Second)
			w.GotoUrl(url)
		}
	} else {
		err := retry(2, 2*time.Second, func() error {
			return w.driver.Get(url)
		})
		return err
	}
	// return retry(2, (2 * time.Second), func() error {
	// 	return w.driver.Get(url)
	//
	// })
	return nil
}

func (w *WebDriverWrapper) EnterText(selector, text string) error {
	elem, err := w.driver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return err
	}
	return elem.SendKeys(text)
}

func (w *WebDriverWrapper) FindElement(selector string) (selenium.WebElement, error) {
	var elem selenium.WebElement
	err := retry(3, 1*time.Second, func() error {
		var innerErr error
		elem, innerErr = w.driver.FindElement(selenium.ByCSSSelector, selector)
		return innerErr
	})
	return elem, err
}

func (w *WebDriverWrapper) FindElements(selector string) ([]selenium.WebElement, error) {
	var elems []selenium.WebElement
	err := retry(2, 1*time.Second, func() error {
		var innerErr error
		elems, innerErr = w.driver.FindElements(selenium.ByCSSSelector, selector)
		return innerErr
	})
	return elems, err
}

func (w *WebDriverWrapper) CheckElementExists(selector string) bool {
	_, err := w.driver.FindElement(selenium.ByCSSSelector, selector)
	return err == nil
}

func (w *WebDriverWrapper) GetInnerText(containerElement selenium.WebElement, selector string) (string, error) {
	elem, err := containerElement.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return "", err
	}
	text, err := elem.Text()
	if err != nil {
		return "", err
	}
	return text, nil
}

func (w *WebDriverWrapper) ExecuteScript(script string) (any, error) {
	res, err := w.driver.ExecuteScript(script, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (w *WebDriverWrapper) ExecuteScriptAsync(script string) (any, error) {
	res, err := w.driver.ExecuteScriptAsync(script, nil)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// func (w *WebDriverWrapper) GetInnerText(selector string) (string, error) {
// 	elem, err := w.driver.FindElement(selenium.ByCSSSelector, selector)
// 	if err != nil {
// 		return "", fmt.Errorf("Error finding element with selector %s: %v", selector, err)
// 	}
// 	text, err := w.GetInnerTextFromElement(elem)
// 	if err != nil {
// 		return "", err
// 	}
// 	return text, nil
// }
//
// func (w *WebDriverWrapper) GetInnerTextFromElement(elem selenium.WebElement) (string, error) {
// 	text, err := elem.Text()
// 	if err != nil {
// 		return "", fmt.Errorf("Error getting text on element %v: %v", elem, err)
// 	}
// 	return text, nil
// }

func (w *WebDriverWrapper) Close() {
	if w.driver != nil {
		w.driver.Quit()
	}
	if w.service != nil {
		w.service.Stop()
	}
}

// func Main() {
// 	const (
// 		serverUrl = "http://localhost:4444"
// 	)
//
// 	caps := selenium.Capabilities{"browserName": "firefox"}
//
// 	wd, err := selenium.NewRemote(caps, serverUrl)
// 	if err != nil {
// 		log.Fatal("Error starting remote: ", err)
// 	}
//
// 	defer wd.Quit()
//
// 	err = wd.Get("https://google.com")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	title, err := wd.Title()
// 	if err != nil {
// 		log.Fatal(err)
//
// 	}
//
// 	fmt.Println(title)
// }
