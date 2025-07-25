package telnavi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
		matches := percentageRegex.FindStringSubmatch(*rawString)
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
			if data.VitalInfo.Name == nil {
				cleanName, suffixes := extractBusinessName(&val)

				*data.BusinessDetails.NameSuffixes = suffixes
				// t.currentVitalInfo.Name = cleanName
				// t.vitalInfoChannel <- *t.currentVitalInfo
				*data.VitalInfo.Name = cleanName
			}
		case "住所":
			if data.BusinessDetails.LocationDetails == (providers.LocationDetails{}) {
				if err := japaneseinfo.GetAddressInfo(val, &data.BusinessDetails.LocationDetails); err != nil {
					data.BusinessDetails.LocationDetails = providers.LocationDetails{
						Prefecture: nil,
					}
					continue
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
			*data.VitalInfo.Industry = val
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
			*data.BusinessDetails.NameSuffixes = suffixes
			// t.currentVitalInfo.Name = cleanedName
			// t.vitalInfoChannel <- *t.currentVitalInfo
			*data.VitalInfo.Name = cleanedName
		case "住所":
			if err := japaneseinfo.GetAddressInfo(val, &data.BusinessDetails.LocationDetails); err != nil {
				continue
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

	t.driver.GotoUrl(phoneNumberInfoPageUrl)
	t.driver.LoadCookies(webdriver.TelnaviWebScrapingProvider)

	fmt.Printf("[telnavi] Loaded page and cookies\n")
	// FIXME: Hung here \/\/\/\/
	businessTableContainer, err := t.driver.FindElement("div.info_table:nth-child(1) > table > tbody:nth-child(1)")
	if err != nil {
		if strings.Contains(err.Error(), "no such element") {
			fmt.Printf("[telnavi] No business table found, continuing...\n")
		} else {
			return data, err
		}
	} else {
		fmt.Printf("[telnavi] Found business table, processing...\n")
		businessTableEntries, err := webdriver.GetTableInformation(t.driver, businessTableContainer, "th", "td")
		if err != nil {
			return data, err
		}
		if err := t.getBusinessInfo(&data, businessTableEntries); err != nil {
			return data, err
		}
		fmt.Printf("[telnavi] Business info processed\n")
	}

	fmt.Printf("[telnavi] Looking for phone number table...\n")

	phoneNumberTableContainer, err := t.driver.FindElement("div.info_table:nth-child(2) > table > tbody")
	if err == nil {
		phoneNumberTableEntries, err := webdriver.GetTableInformation(t.driver, phoneNumberTableContainer, "th", "td")
		if err != nil {
			return data, err
		}
		if err := t.getPhoneNumberInfo(&data, phoneNumberTableEntries); err != nil {
			return data, err
		}
		fmt.Printf("[telnavi] Phone number info processed\n")
	} else {
		fmt.Printf("[telnavi] Could not find phone number table: %v\n", err)
	}

	comments, reviewCount := extractAllComments(t.driver, phoneNumberInfoPageUrl)
	data.SiteInfo.ReviewCount = reviewCount
	data.SiteInfo.Comments = comments

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

// extractAllComments extracts all comments for the given phone number page(s) using BatchExtractComments.
func extractAllComments(driver *webdriver.WebDriverWrapper, phoneNumberInfoPageUrl string) ([]providers.Comment, int) {
	var allComments []providers.Comment
	reviewCount := 0
	processedPages := map[int]bool{}

	// Always start on the first page
	driver.GotoUrl(phoneNumberInfoPageUrl)
	driver.LoadCookies(webdriver.TelnaviWebScrapingProvider)

	userCommentsContainer, err := driver.FindElement("div.kuchikomi_thread_content")
	if err != nil {
		logging.Warn().Err(err).Msg("Failed to find user comments container")
		return nil, 0
	}

	// Always try to extract comments from the current page (page 1)
	batchComments, err := webdriver.BatchExtractComments(
		driver,
		userCommentsContainer,
		"#thread",
		"tbody tr:nth-child(1) > td:nth-child(1) > time:nth-child(1)",
		"tbody tr:nth-child(2) > td > div",
	)
	if err != nil {
		logging.Warn().Err(err).Msg("Failed to extract comments on page 1")
	} else {
		reviewCount += len(batchComments)
		for _, entry := range batchComments {
			parsedDate, err := utils.ParseDate("2006年1月2日 15時4分", entry["date"])
			if err != nil {
				logging.Warn().Err(err).Msg("Failed to parse comment date")
				continue
			}
			comment := providers.Comment{
				Text:     entry["text"],
				PostDate: parsedDate,
			}
			allComments = append(allComments, comment)
		}
	}
	processedPages[1] = true

	// Now check for pagination
	paginationControlElement, err := userCommentsContainer.FindElement(selenium.ByCSSSelector, "div.paginationControl")
	pages := []int{}
	if err == nil {
		spans, err := paginationControlElement.FindElements(selenium.ByTagName, "span")
		if err != nil {
			logging.Warn().Err(err).Msg("Failed to find span elements in pagination")
		} else if len(spans) < 2 {
			links, err := paginationControlElement.FindElements(selenium.ByTagName, "a")
			if err != nil {
				logging.Warn().Err(err).Msg("Failed to find link elements in pagination")
			} else {
				for _, elem := range links {
					rawPageNumber, err := elem.Text()
					if err != nil {
						logging.Warn().Err(err).Msg("Couldn't get page number text")
						continue
					}
					parsedPageNumber, err := strconv.Atoi(rawPageNumber)
					if err != nil {
						continue
					}
					if !processedPages[parsedPageNumber] {
						pages = append(pages, parsedPageNumber)
					}
				}
			}
		}
	}

	// Process each additional page
	for _, pageNumber := range pages {
		driver.GotoUrl(fmt.Sprintf("%s?page=%d", phoneNumberInfoPageUrl, pageNumber))
		driver.LoadCookies(webdriver.TelnaviWebScrapingProvider)
		userCommentsContainer, err := driver.FindElement("div.kuchikomi_thread_content")
		if err != nil {
			logging.Warn().Err(err).Msg("Failed to find user comments container on page")
			continue
		}
		batchComments, err := webdriver.BatchExtractComments(
			driver,
			userCommentsContainer,
			"#thread",
			"tbody tr:nth-child(1) > td:nth-child(1) > time:nth-child(1)",
			"tbody tr:nth-child(2) > td > div",
		)
		if err != nil {
			logging.Warn().Err(err).Msgf("Failed to extract comments on page %d", pageNumber)
			continue
		}
		reviewCount += len(batchComments)
		for _, entry := range batchComments {
			parsedDate, err := utils.ParseDate("2006年1月2日 15時4分", entry["date"])
			if err != nil {
				logging.Warn().Err(err).Msg("Failed to parse comment date")
				continue
			}
			comment := providers.Comment{
				Text:     entry["text"],
				PostDate: parsedDate,
			}
			allComments = append(allComments, comment)
		}
		processedPages[pageNumber] = true
	}
	return allComments, reviewCount
}
