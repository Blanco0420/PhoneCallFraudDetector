package jpnumber

import (
	"encoding/json"
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

	"github.com/tebeka/selenium"
	"github.com/ttacon/libphonenumber"
)

const (
	baseUrl                        = "https://www.jpnumber.com"
	initialPhoneNumberInfoSelector = ".frame-728-orange-l > table:nth-child(2) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1) > div:nth-child(1)"
	lineTypeSelector               = ".frame-728-orange-l > table:nth-child(2) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1) > div:nth-child(1) > dt:nth-child(3)"
	phoneNumberInfoSelector        = "div.frame-728-green-l:nth-child(2) > div:nth-child(2) > table:nth-child(1) > tbody:nth-child(1)"
	searchSelector                 = "#number"
	commentDateSelector            = ".title-background-pink table tbody tr td:nth-child(2) table  tbody tr td:nth-child(1)"
	commentTextSelector            = "div:nth-child(2) > dt:nth-child(1)"
)

//TODO: Search for the city in the businessDetails Address (for each (rune?token?) check if exists in japaneseInfo)

type JpNumberSource struct {
	driver *webdriver.WebDriverWrapper
	// vitalInfoChannel chan providers.VitalInfo
	// currentVitalInfo *providers.VitalInfo
}

// func (s *JpNumberSource) VitalInfoChannel() <-chan providers.VitalInfo {
// 	return s.vitalInfoChannel
// }
//
// func (s *JpNumberSource) CloseVitalInfoChannel() {
// 	close(s.vitalInfoChannel)
// }

func Initialize() (*JpNumberSource, error) {
	driver, err := webdriver.InitializeDriver(webdriver.JpNumberWebScrapingProvider)
	if err != nil {
		return &JpNumberSource{}, err
	}
	return &JpNumberSource{
		driver: driver,
		// vitalInfoChannel: make(chan providers.VitalInfo),
	}, nil
}

func (s *JpNumberSource) Close() {
	s.driver.Close()
}

func (s *JpNumberSource) getGraphData(graphData *[]providers.GraphData) error {
	script := `
	var callback = arguments[arguments.length - 1];
var interval = setInterval(() => {
  var script = document.querySelector('script.code');
  if (script) {
    clearInterval(interval);
    callback(script.innerHTML); // this returns to Go
  }
}, 100);
	`
	rawScriptText, err := s.driver.ExecuteScriptAsync(script)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`(?s)var accessPoints=\[(.*?)\];`)
	scriptText := rawScriptText.(string)
	match := re.FindStringSubmatch(scriptText)

	if len(match) == 1 {
		return fmt.Errorf("no match found")
	}

	rawGraphDataString := match[1]
	rawGraphDataString = fmt.Sprintf("[%v]", strings.ReplaceAll(rawGraphDataString, "'", "\""))
	if err := utils.ParseGraphData(rawGraphDataString, graphData); err != nil {
		return err
	}
	return nil
}

func (s *JpNumberSource) getNumberParts(number string) ([]string, error) {
	num, err := libphonenumber.Parse(number, "JP")
	if err != nil {
		return []string{}, err
	}
	return strings.Split(libphonenumber.Format(num, libphonenumber.NATIONAL), "-"), nil
}

func (s *JpNumberSource) getDetailsPageURL(lineType providers.LineType, number string) (string, error) {
	// https://www.jpnumber.com/ipphone/numberinfo_050_5482_2807.html
	var url = baseUrl
	switch lineType {
	case providers.LineTypeMobile:
		url = url + "/mobile"
	case providers.LineTypeTollFree:
		url = url + "/freedial"
	case providers.LineTypeVOIP:
		url = url + "/ipphone"
	}
	parts, err := s.getNumberParts(number)
	if err != nil {
		return "", err
	}
	return url + fmt.Sprintf("/numberinfo_%s_%s_%s.html", parts[0], parts[1], parts[2]), nil
}

func (s *JpNumberSource) getComments() ([]providers.Comment, error) {
	var comments []providers.Comment
	commentsContainer, err := s.driver.FindElement("#result-main-right > span:nth-child(6)")
	if err != nil {
		return []providers.Comment{}, err
	}
	commentElements, err := commentsContainer.FindElements(selenium.ByCSSSelector, "div.frame-728-gray-l")
	commentElements = commentElements[:len(commentElements)-1]
	if err != nil {
		return []providers.Comment{}, err
	}

	for _, elem := range commentElements {
		dateElement, err := elem.FindElement(selenium.ByCSSSelector, commentDateSelector)

		if err != nil {
			elem, err := json.MarshalIndent(dateElement, "", "  ")
			fmt.Println(string(elem))
			return []providers.Comment{}, fmt.Errorf("error getting date Element: %v", err)
		}
		commentText, err := s.driver.GetInnerText(elem, "div:nth-child(2) > dt:nth-child(1)")
		if err != nil {
			return []providers.Comment{}, fmt.Errorf("comment text error!\n%v", err)
		}
		dateText, err := s.driver.GetInnerText(elem, ".title-background-pink table tbody tr td:nth-child(2) table  tbody tr td:nth-child(1)")
		if err != nil {
			return []providers.Comment{}, fmt.Errorf("comment date error!\n%v", err)
		}

		parsedDate, err := utils.ParseDate("2006/01/02 15:04:05", dateText)
		if err != nil {
			return []providers.Comment{}, fmt.Errorf("parsing date error:\n%v", err)
		}

		comments = append(comments, providers.Comment{Text: commentText, PostDate: parsedDate})

	}
	return comments, nil
}

// func getTextFromTd(row selenium.WebElement) (label string, value string, err error) {
// 	cols, err := row.FindElements(selenium.ByTagName, "td")
// 	if err != nil || len(cols) < 2 {
// 		return "", "", fmt.Errorf("Invalid table row format")
// 	}
// 	label, err = cols[0].Text()
// 	if err != nil {
// 		return "", "", err
// 	}
// 	value, err = cols[1].Text()
// 	if err != nil {
// 		return "", "", err
// 	}
// 	return label, value, nil
// }

func (s *JpNumberSource) getBusinessInfo(data *providers.NumberDetails, businessDetails *providers.BusinessDetails) error {

	businessInfoTableContainer, err := s.driver.FindElement("div.frame-728-green-l:nth-child(4)")
	if err != nil {
		return fmt.Errorf("no business details available")
	}

	businessInfoTableElement, err := businessInfoTableContainer.FindElement(selenium.ByCSSSelector, "div.content > table > tbody")
	if err != nil {
		if strings.Contains(err.Error(), "no such element") {
		} else {
			return err
		}
	} else {
		tableEntries, err := webdriver.GetTableInformation(s.driver, businessInfoTableElement, "td", "td")
		if err != nil {
			return err
		}

		for _, entry := range tableEntries {

			key := entry.Key
			value := entry.Value

			switch key {
			case "Name", "事業者名称":
				cleanName, suffixes := utils.GetSuffixesFromCompanyName(&value)
				data.BusinessDetails.NameSuffixes = suffixes
				data.VitalInfo.Name = cleanName
				// s.currentVitalInfo.Name = cleanName
				// s.vitalInfoChannel <- *s.currentVitalInfo
			case "Industry", "業種":
				// s.currentVitalInfo.Industry = value
				// s.vitalInfoChannel <- *s.currentVitalInfo
				data.VitalInfo.Industry = value
			case "Address", "住所":
				japaneseinfo.GetAddressInfo(value, &businessDetails.LocationDetails)
			case "Official website", "公式サイト":
				businessDetails.Website = value
			case "Business", "事業紹介":
				// s.currentVitalInfo.CompanyOverview = value
				// s.vitalInfoChannel <- *s.currentVitalInfo
				data.VitalInfo.CompanyOverview = value
			}
		}
	}

	// rows, err := businessInfoElementContainer.FindElements(selenium.ByTagName, "tr")
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (s *JpNumberSource) GetData(number string) (providers.NumberDetails, error) {
	numberQuery := fmt.Sprintf("%s/searchnumber.do?number=%s", baseUrl, number)
	var data providers.NumberDetails
	// s.currentVitalInfo = &data.VitalInfo

	data.Number = number
	var siteInfo providers.SiteInfo

	s.driver.GotoUrl(numberQuery)
	s.driver.LoadCookies(webdriver.JpNumberWebScrapingProvider)

	//TODO: use the utils getTableInfo function eventually (jpnumber is difficult and doesn't split their table by tr > th,td . Instead, tr >td,td,td,td for like 3 different key:val pairs)

	// Get line type
	initialPhoneNumberInfoContainer, err := s.driver.FindElement(initialPhoneNumberInfoSelector)
	if err != nil {
		return data, err
	}
	text, err := s.driver.GetInnerText(initialPhoneNumberInfoContainer, "dt:nth-child(3)")
	if err != nil {
		return data, err
	}
	rawLineType := strings.ReplaceAll(strings.Split(text, ">")[0], " ", "")
	lineType, err := utils.GetLineType(rawLineType)
	if err != nil {
		return data, err
	}
	// s.currentVitalInfo.LineType = lineType
	// s.vitalInfoChannel <- *s.currentVitalInfo
	data.VitalInfo.LineType = lineType

	// goto detailed page
	detailesPagesUrl, err := s.getDetailsPageURL(lineType, number)
	if err != nil {
		return data, err
	}
	s.driver.GotoUrl(detailesPagesUrl)
	s.driver.LoadCookies(webdriver.JpNumberWebScrapingProvider)

	if err := s.getBusinessInfo(&data, &data.BusinessDetails); err != nil {
		if strings.Contains(err.Error(), "no business details available") {
		} else {
			return data, err
		}
	}

	//TODO: Move all of this to another function (getNumberMainInfo)
	phoneNumberInfoContainer, err := s.driver.FindElement(phoneNumberInfoSelector)
	if err != nil {
		return data, err
	}
	prefecture, _ := s.driver.GetInnerText(phoneNumberInfoContainer, "tr:nth-child(4)>td:nth-child(2)")
	data.BusinessDetails.LocationDetails.Prefecture = prefecture

	carrier, _ := s.driver.GetInnerText(phoneNumberInfoContainer, "tr:nth-child(3)>td:nth-child(4)")
	data.Carrier = carrier

	reviewCount, err := s.driver.GetInnerText(phoneNumberInfoContainer, "span.red")
	if err != nil {
		return data, err
	}
	if reviewCount != "" {
		i, err := strconv.Atoi(reviewCount)
		if err != nil {
			return data, err
		}
		siteInfo.ReviewCount = i
	}

	if siteInfo.ReviewCount >= 1 {
		comments, err := s.getComments()
		if err != nil {
			return data, err
		}
		siteInfo.Comments = comments
	}

	data.SiteInfo = siteInfo

	var graphData []providers.GraphData
	if err := s.getGraphData(&graphData); err != nil {
		return data, err
	}

	numberRiskInput := providerdataprocessing.NumberRiskInput{
		SourceName:  "jpnumber",
		GraphData:   graphData,
		Comments:    data.SiteInfo.Comments,
		RecentAbuse: &data.VitalInfo.FraudulentDetails.RecentAbuse,
	}
	go func() {
		if err := s.driver.SaveCookies(webdriver.JpNumberWebScrapingProvider); err != nil {
			logging.Error().Err(err).Msg("Failed to save cookies for jpnumber provider")
		}
	}()
	// s.currentVitalInfo.OverallFraudScore = providerdataprocessing.EvaluateSource(numberRiskInput)
	// s.vitalInfoChannel <- *s.currentVitalInfo
	data.VitalInfo.OverallFraudScore = providerdataprocessing.EvaluateSource(numberRiskInput)

	return data, nil
}
