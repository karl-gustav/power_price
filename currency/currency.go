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
	if date.After(time.Now()) {
		date = date.AddDate(0, 0, -1)
	}
	url := fmt.Sprintf(
		currencyURL,
		fromCurrency,
		toCurrency,
		date.AddDate(0, 0, -7).Format(common.StdDateFormat),
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
	obs := exchangeRateInfo.DataSet.Series.Obs

	return obs[len(obs)-1].ObsValue.Value / math.Pow10(multiplicator), nil
}

type ExchangeRateInfo struct {
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
			Obs []struct {
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
