package currency

import (
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/karl-gustav/power_price/common"
)

const (
	currencyURL = "https://data.norges-bank.no/api/data/EXR/B.%s.%s.SP?format=sdmx-generic-2.1&startPeriod=%s&endPeriod=%s&locale=en"
)

func GetExchangeRate(fromCurrency, toCurrency string, date time.Time) (float64, error) {
	date = findValidDate(date) //can't use weekends or future dates
	url := fmt.Sprintf(
		currencyURL,
		fromCurrency,
		toCurrency,
		date.Format(common.StdDateFormat),
		date.Format(common.StdDateFormat),
	)
	exchangeRateInfoBody, err := common.GetUrl(url, []string{})
	if err != nil {
		return 0, err
	}
	var exchangeRateInfo ExchangeRateInfo
	err = xml.Unmarshal(exchangeRateInfoBody, &exchangeRateInfo)
	if err != nil {
		return 0, err
	}

	var multiplicator int
	for _, attr := range exchangeRateInfo.DataSet.Series.Attributes.Value {
		if attr.ID == "UNIT_MULT" {
			multiplicator, _ = strconv.Atoi(attr.Value)
		}
	}

	return exchangeRateInfo.DataSet.Series.Obs.ObsValue.Value / math.Pow10(multiplicator), nil
}

// findValidDate changes a date to make sure it is on weekdays and not in
// the future
func findValidDate(date time.Time) time.Time {
	now := time.Now()
	if date.After(now) {
		date = now
	}
	switch date.Weekday() {
	case time.Saturday:
		date = date.AddDate(0, 0, -1)
	case time.Sunday:
		date = date.AddDate(0, 0, -2)
	}
	return date
}

type ExchangeRateInfo struct {
	XMLName xml.Name `xml:"GenericData"`
	Text    string   `xml:",chardata"`
	Footer  string   `xml:"footer,attr"`
	Generic string   `xml:"generic,attr"`
	Common  string   `xml:"common,attr"`
	Message string   `xml:"message,attr"`
	Xsi     string   `xml:"xsi,attr"`
	XML     string   `xml:"xml,attr"`
	/*	Header  struct {
		Text     string `xml:",chardata"`
		ID       string `xml:"ID"`
		Test     string `xml:"Test"`
		Prepared string `xml:"Prepared"`
		Sender   struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
		} `xml:"Sender"`
		Receiver struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
		} `xml:"Receiver"`
		Structure struct {
			Text                   string `xml:",chardata"`
			StructureID            string `xml:"structureID,attr"`
			DimensionAtObservation string `xml:"dimensionAtObservation,attr"`
			StructureUsage         struct {
				Text string `xml:",chardata"`
				Ref  struct {
					Text     string `xml:",chardata"`
					AgencyID string `xml:"agencyID,attr"`
					ID       string `xml:"id,attr"`
					Version  string `xml:"version,attr"`
				} `xml:"Ref"`
			} `xml:"StructureUsage"`
		} `xml:"Structure"`
		DataSetAction  string `xml:"DataSetAction"`
		Extracted      string `xml:"Extracted"`
		ReportingBegin string `xml:"ReportingBegin"`
		ReportingEnd   string `xml:"ReportingEnd"`
	} `xml:"Header"`*/
	DataSet struct {
		Text         string `xml:",chardata"`
		StructureRef string `xml:"structureRef,attr"`
		Series       struct {
			Text      string `xml:",chardata"`
			SeriesKey struct {
				Text  string `xml:",chardata"`
				Value []struct {
					Text  string `xml:",chardata"`
					ID    string `xml:"id,attr"`
					Value string `xml:"value,attr"`
				} `xml:"Value"`
			} `xml:"SeriesKey"`
			Attributes struct {
				Text  string `xml:",chardata"`
				Value []struct {
					Text  string `xml:",chardata"`
					ID    string `xml:"id,attr"`
					Value string `xml:"value,attr"`
				} `xml:"Value"`
			} `xml:"Attributes"`
			Obs struct {
				Text         string `xml:",chardata"`
				ObsDimension struct {
					Text  string `xml:",chardata"`
					Value string `xml:"value,attr"`
				} `xml:"ObsDimension"`
				ObsValue struct {
					Text  string  `xml:",chardata"`
					Value float64 `xml:"value,attr"`
				} `xml:"ObsValue"`
			} `xml:"Obs"`
		} `xml:"Series"`
	} `xml:"DataSet"`
}
