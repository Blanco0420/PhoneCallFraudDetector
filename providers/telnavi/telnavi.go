package telnavi

import (
	japaneseinfo "PhoneNumberCheck/japaneseInfo"
	"PhoneNumberCheck/logging"
	providerdataprocessing "PhoneNumberCheck/providerDataProcessing"
	"PhoneNumberCheck/providers"
	"PhoneNumberCheck/utils"
	webdriver "PhoneNumberCheck/webDriver"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
var callback = arguments[arguments.length - 1];
var interval = setInterval(() => {
  if (window.pageview_stat) {
    clearInterval(interval);
    callback(JSON.stringify(window.pageview_stat)); // return JSON string of the object
  }
}, 100);`

	rawData, err := t.driver.ExecuteScriptAsync(script)
	if err != nil {
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
	var data providers.NumberDetails
	// t.currentVitalInfo = &data.VitalInfo
	data.Number = phoneNumber
	phoneNumberInfoPageUrl := fmt.Sprintf("%s/%s", baseUrl, phoneNumber)
	t.driver.GotoUrl(phoneNumberInfoPageUrl)
	t.driver.LoadCookies(webdriver.TelnaviWebScrapingProvider)

	businessTableContainer, err := t.driver.FindElement("div.info_table:nth-child(1) > table > tbody:nth-child(1)")
	if err != nil {
		if strings.Contains(err.Error(), "no such element") {
		} else {
			return providers.NumberDetails{}, err
		}
	} else {
		businessTableEntries, err := webdriver.GetTableInformation(t.driver, businessTableContainer, "th", "td")
		if err != nil {
			return providers.NumberDetails{}, err
		}
		if err := t.getBusinessInfo(&data, businessTableEntries); err != nil {
			return providers.NumberDetails{}, err
		}
	}
	phoneNumberTableContainer, err := t.driver.FindElement("div.info_table:nth-child(2) > table > tbody")
	if err != nil {
		return providers.NumberDetails{}, err
	}

	phoneNumberTableEntries, err := webdriver.GetTableInformation(t.driver, phoneNumberTableContainer, "th", "td")
	if err != nil {
		return providers.NumberDetails{}, err
	}
	if err := t.getPhoneNumberInfo(&data, phoneNumberTableEntries); err != nil {
		return providers.NumberDetails{}, err
	}

	// businessInfoContainer, err = businessInfoContainer.FindElement(selenium.ByCSSSelector, "table:nth-child(1) > tbody:nth-child(1)")
	// if err != nil {
	// 	return err
	// }

	// rawUserRating, err := t.driver.GetInnerText(phoneNumberInfoContainer, "tr:nth-child(13) > td")
	// if err != nil {
	// 	return providers.NumberDetails{}, err
	// }
	userCommentsContainer, err := t.getUserCommentsContainer()
	if err != nil {
		return providers.NumberDetails{}, err
	}

	paginationControlElement, err := userCommentsContainer.FindElement(selenium.ByCSSSelector, "div.paginationControl")
	if err != nil {
		panic(fmt.Errorf("error comments have no pagination element: %v", err))
	}
	spans, err := paginationControlElement.FindElements(selenium.ByTagName, "span")
	if err != nil {
		panic(fmt.Errorf("error, pagination has no span elements"))
	}
	pages := []int{1}
	if len(spans) < 2 {
		links, err := paginationControlElement.FindElements(selenium.ByTagName, "a")
		if err != nil {
			panic(fmt.Errorf("can't find any link elements: %v", err))
		}
		for _, elem := range links {
			rawPageNumber, err := elem.Text()
			if err != nil {
				panic(fmt.Errorf("couldn't get page number text: %v", err))
			}
			parsedPageNumber, err := strconv.Atoi(rawPageNumber)
			if err != nil {
				continue
			}
			pages = append(pages, parsedPageNumber)
		}
	}
	reviewCount := 0

	for _, pageNumber := range pages {
		if pageNumber != 1 {
			t.driver.GotoUrl(fmt.Sprintf("%s?page=%d", phoneNumberInfoPageUrl, pageNumber))
			t.driver.LoadCookies(webdriver.TelnaviWebScrapingProvider)
		}
		//TODO: Make this into a function. Pretty much make everything comment wise into separated functions. And maybe later for re-usability on other providers
		userCommentsContainer, err = t.getUserCommentsContainer()
		if err != nil {
			return providers.NumberDetails{}, err
		}
		commentsElements, err := userCommentsContainer.FindElements(selenium.ByCSSSelector, "#thread")
		if err != nil {
			panic(fmt.Errorf("no comments?: %v", err))
		}

		reviewCount += len(commentsElements)
		for _, elem := range commentsElements {
			var comment providers.Comment

			tableBody, err := elem.FindElement(selenium.ByCSSSelector, "tbody")
			if err != nil {
				panic(fmt.Errorf("couldn't get comment table body? %v", err))
			}
			dateElement, err := tableBody.FindElement(selenium.ByCSSSelector, "tr:nth-child(1) > td:nth-child(1) > time:nth-child(1)")
			if err != nil {
				panic(fmt.Errorf("failed to get date element: %v", err))
			}
			dateString, err := dateElement.Text()
			if err != nil {
				panic(fmt.Errorf("failed to get content attr. from date elem: %v", err))
			}
			formattedDate, err := utils.ParseDate("2006年1月2日 15時4分", dateString)
			if err != nil {
				panic(fmt.Errorf("failed to parse date: %v", err))
			}
			comment.PostDate = formattedDate

			commentText, err := t.driver.GetInnerText(tableBody, "tr:nth-child(2) > td > div")
			if err != nil {
				panic(fmt.Errorf("failed to get comment text: %v", err))
			}
			comment.Text = commentText
			data.SiteInfo.Comments = append(data.SiteInfo.Comments, comment)
		}
	}
	data.SiteInfo.ReviewCount = reviewCount

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
	return data, nil
}
