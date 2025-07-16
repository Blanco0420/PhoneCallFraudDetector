package telnavi

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	japaneseinfo "github.com/Blanco0420/Phone-Number-Check/backend/japaneseInfo"
	"github.com/Blanco0420/Phone-Number-Check/backend/logging"
	providerdataprocessing "github.com/Blanco0420/Phone-Number-Check/backend/providerDataProcessing"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
	"github.com/Blanco0420/Phone-Number-Check/backend/utils"
	webdriver "github.com/Blanco0420/Phone-Number-Check/backend/webDriver"

	_ "net/http/pprof"

	"github.com/tebeka/selenium"
)

const (
	baseUrl = "https://www.telnavi.jp/phone"
	//NOTE: 推定発信地域 estimated outgoing area (maybe use later)
)

type TelnaviSource struct {
	driver *webdriver.WebDriverWrapper
	// vitalInfoChannel chan providers.VitalInfo
	// currentVitalInfo *providers.VitalInfo
}

// func (t *TelnaviSource) VitalInfoChannel() <-chan providers.VitalInfo {
// 	return t.vitalInfoChannel
// }
//
// func (t *TelnaviSource) CloseVitalInfoChannel() {
// 	close(t.vitalInfoChannel)
// }

func Initialize() (*TelnaviSource, error) {
	driver, err := webdriver.InitializeDriver(webdriver.TelnaviWebScrapingProvider)
	if err != nil {
		fmt.Println("here")
		return &TelnaviSource{}, err
	}
	return &TelnaviSource{
		driver: driver,
		// vitalInfoChannel: make(chan providers.VitalInfo),
	}, nil
}

func (t *TelnaviSource) Close() {
	t.driver.Close()
}

func (t *TelnaviSource) calculateFraudScore(ratingTableContainer selenium.WebElement) (int, error) {
	tableRows, err := ratingTableContainer.FindElement(selenium.ByCSSSelector, "td > table > tbody")
	if err != nil {
		panic(fmt.Errorf("row exists but not the info?? : %v", err))
	}
	values, err := tableRows.FindElements(selenium.ByCSSSelector, "tr")
	if err != nil {
		panic("IDEK")
	}
	percentageRegex := regexp.MustCompile(`(\d+)%`)
	var ratings []int
	for _, val := range values {
		rawString, err := t.driver.GetInnerText(val, "td:nth-child(1)")
		if err != nil {
			panic(fmt.Errorf("csould not get the inner text of fraud score? : %v", err))
		}
		matches := percentageRegex.FindStringSubmatch(rawString)
		if len(matches) < 2 {
			panic(fmt.Errorf("matches len is < 0 : %v", matches))
		} else if len(matches) > 2 {
			panic(fmt.Errorf("matches length is more than 2???: %v", matches))
		}
		score, err := strconv.Atoi(matches[1])
		if err != nil {
			panic(fmt.Errorf("error formatting decimal: %v", err))
		}
		ratings = append(ratings, score)
	}
	fraudScore := ratings[2] + ratings[1]/2
	return fraudScore, nil
}

func (t *TelnaviSource) getGraphData(graphData *[]providers.GraphData) error {
	script := `
  if (window.pageview_stat) {
    return JSON.stringify(window.pageview_stat); // return JSON string of the object
  }
`

	rawData, err := t.driver.ExecuteScript(script)
	if err != nil {
		fmt.Println("Error, failed to get graph data: ", err)
		return err
	}

	rawDataString, ok := rawData.(string)
	if !ok {
		return fmt.Errorf("unexpected graph data %v of type %T", rawData, rawData)
	}

	if err := utils.ParseGraphData(rawDataString, graphData); err != nil {
		return err
	}
	return nil
}

func (t *TelnaviSource) getPhoneNumberInfo(data *providers.NumberDetails, tableEntries []providers.TableEntry) error {
	for _, entry := range tableEntries {
		key := entry.Key
		val := entry.Value
		switch key {
		case "事業者名":
			// if t.currentVitalInfo.Name == "" {
			if data.VitalInfo.Name == "" {
				cleanName, suffixes := extractBusinessName(&val)

				data.BusinessDetails.NameSuffixes = suffixes
				// t.currentVitalInfo.Name = cleanName
				// t.vitalInfoChannel <- *t.currentVitalInfo
				data.VitalInfo.Name = cleanName
			}
		case "住所":
			if data.BusinessDetails.LocationDetails == (providers.LocationDetails{}) {
				if err := japaneseinfo.GetAddressInfo(val, &data.BusinessDetails.LocationDetails); err != nil {
					return err
				}
			}
		case "回線種別":

			lineType, err := utils.GetLineType(val)
			if err != nil {
				fmt.Println("Error, failed to get line type: ", err)
			}
			// t.currentVitalInfo.LineType = lineType
			// t.vitalInfoChannel <- *t.currentVitalInfo
			data.VitalInfo.LineType = lineType
		case "業種タグ":
			// t.currentVitalInfo.Industry = val
			// t.vitalInfoChannel <- *t.currentVitalInfo
			data.VitalInfo.Industry = val
		case "ユーザー評価":
			rating, err := getCleanRating(val)
			if err != nil {
				fmt.Println("Error, failed to get clean rating: ", err)
			}
			data.SiteInfo.UserRating = rating
		case "アクセス数":
			val = strings.TrimSpace(val)
			if val == "10回未満" {

			}
			re := regexp.MustCompile(`[^0-9]`)
			cleanedAccessCount := re.ReplaceAllString(val, "")
			accessCount, err := strconv.Atoi(cleanedAccessCount)
			if err != nil {
				fmt.Printf("CleanedAccessCount: %s\naccessedCount: %d", cleanedAccessCount, accessCount)
				panic(fmt.Errorf("failed to parse access count: %v", err))
			}
			data.SiteInfo.AccessCount = accessCount
		case "迷惑電話度":
			//TODO: Channel
			fraudScore, err := t.calculateFraudScore(entry.Element)
			if err != nil {
				if strings.Contains(err.Error(), "no such element") {
					// t.currentVitalInfo.FraudulentDetails.FraudScore = 0
					// t.vitalInfoChannel <- *t.currentVitalInfo
					data.VitalInfo.FraudulentDetails.FraudScore = 0
				} else {
					return err
				}
			} else {
				// t.currentVitalInfo.FraudulentDetails.FraudScore = fraudScore
				// t.vitalInfoChannel <- *t.currentVitalInfo
				data.VitalInfo.FraudulentDetails.FraudScore = fraudScore
			}
		default:
			continue

		}
	}
	return nil
}

func (t *TelnaviSource) getBusinessInfo(data *providers.NumberDetails, businessTableEntries []providers.TableEntry) error {
	//TODO: Check if doesn't exist
	for _, entry := range businessTableEntries {
		key := entry.Key
		val := entry.Value

		switch key {
		case "事業者名":
			cleanedName, suffixes := extractBusinessName(&val)
			data.BusinessDetails.NameSuffixes = suffixes
			// t.currentVitalInfo.Name = cleanedName
			// t.vitalInfoChannel <- *t.currentVitalInfo
			data.VitalInfo.Name = cleanedName
		case "住所":
			if err := japaneseinfo.GetAddressInfo(val, &data.BusinessDetails.LocationDetails); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *TelnaviSource) getUserCommentsContainer() (selenium.WebElement, error) {

	userCommentsContainer, err := t.driver.FindElement("div.kuchikomi_thread_content")
	if err != nil {
		return userCommentsContainer, err
	}
	return userCommentsContainer, nil
}

func (t *TelnaviSource) GetData(phoneNumber string) (providers.NumberDetails, error) {
	fmt.Printf("[telnavi] Starting GetData for number: %s\n", phoneNumber)
	var data providers.NumberDetails
	// t.currentVitalInfo = &data.VitalInfo
	data.Number = phoneNumber
	phoneNumberInfoPageUrl := fmt.Sprintf("%s/%s", baseUrl, phoneNumber)
	fmt.Printf("[telnavi] Navigating to: %s\n", phoneNumberInfoPageUrl)

	// Add timeout for page loading
	pageCtx, pageCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer pageCancel()

	pageChan := make(chan error, 1)
	go func() {
		t.driver.GotoUrl(phoneNumberInfoPageUrl)
		t.driver.LoadCookies(webdriver.TelnaviWebScrapingProvider)
		pageChan <- nil
	}()

	select {
	case err := <-pageChan:
		if err != nil {
			return providers.NumberDetails{}, fmt.Errorf("failed to load page: %v", err)
		}
	case <-pageCtx.Done():
		fmt.Printf("[telnavi] Timeout loading page, continuing anyway...\n")
	}

	fmt.Printf("[telnavi] Loaded page and cookies\n")

	// Add timeout for business table processing
	businessCtx, businessCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer businessCancel()

	businessChan := make(chan error, 1)
	go func() {
		businessTableContainer, err := t.driver.FindElement("div.info_table:nth-child(1) > table > tbody:nth-child(1)")
		if err != nil {
			if strings.Contains(err.Error(), "no such element") {
				fmt.Printf("[telnavi] No business table found, continuing...\n")
				businessChan <- nil
				return
			} else {
				businessChan <- err
				return
			}
		} else {
			fmt.Printf("[telnavi] Found business table, processing...\n")
			businessTableEntries, err := webdriver.GetTableInformation(t.driver, businessTableContainer, "th", "td")
			if err != nil {
				businessChan <- err
				return
			}
			if err := t.getBusinessInfo(&data, businessTableEntries); err != nil {
				businessChan <- err
				return
			}
			fmt.Printf("[telnavi] Business info processed\n")
			businessChan <- nil
		}
	}()

	select {
	case err := <-businessChan:
		if err != nil {
			return providers.NumberDetails{}, err
		}
	case <-businessCtx.Done():
		fmt.Printf("[telnavi] Timeout processing business info, skipping...\n")
	}

	fmt.Printf("[telnavi] Looking for phone number table...\n")

	// Add timeout for phone number table processing
	phoneCtx, phoneCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer phoneCancel()

	phoneChan := make(chan error, 1)
	go func() {
		phoneNumberTableContainer, err := t.driver.FindElement("div.info_table:nth-child(2) > table > tbody")
		if err != nil {
			phoneChan <- err
			return
		}

		phoneNumberTableEntries, err := webdriver.GetTableInformation(t.driver, phoneNumberTableContainer, "th", "td")
		if err != nil {
			phoneChan <- err
			return
		}
		if err := t.getPhoneNumberInfo(&data, phoneNumberTableEntries); err != nil {
			phoneChan <- err
			return
		}
		fmt.Printf("[telnavi] Phone number info processed\n")
		phoneChan <- nil
	}()

	select {
	case err := <-phoneChan:
		if err != nil {
			return providers.NumberDetails{}, err
		}
	case <-phoneCtx.Done():
		fmt.Printf("[telnavi] Timeout processing phone number info, skipping...\n")
	}

	// Skip comment processing entirely to avoid hanging on heavy JavaScript
	fmt.Printf("[telnavi] Skipping comment processing to avoid hanging on JavaScript\n")
	data.SiteInfo.ReviewCount = 0

	fmt.Printf("[telnavi] Getting graph data...\n")
	var graphData []providers.GraphData
	if err := t.getGraphData(&graphData); err != nil {
		return data, err
	}
	numberRiskInput := providerdataprocessing.NumberRiskInput{
		SourceName:  "telnavi",
		GraphData:   graphData,
		RecentAbuse: &data.VitalInfo.FraudulentDetails.RecentAbuse,
		FraudScore:  &data.VitalInfo.FraudulentDetails.FraudScore,
		Comments:    data.SiteInfo.Comments,
	}
	go func() {
		if err := t.driver.SaveCookies(webdriver.TelnaviWebScrapingProvider); err != nil {
			logging.Error().Err(err).Msg("Failed to save cookies for telnavi provider")
		}
	}()
	overallFraudScore := providerdataprocessing.EvaluateSource(numberRiskInput)
	// t.currentVitalInfo.OverallFraudScore = overallFraudScore
	// t.vitalInfoChannel <- *t.currentVitalInfo
	data.VitalInfo.OverallFraudScore = overallFraudScore
	fmt.Printf("[telnavi] Finished GetData successfully\n")
	return data, nil
}
