package currency

import (
	"context"
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

type ExchangeRate struct {
	Rate float64
	Date string
}

func GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string, date time.Time) (*ExchangeRate, error) {
	// always use previous days exchange rate
	date = date.AddDate(0, 0, -1)
	// get exchange rage 7 days back in time to make sure we get even though there
	// might be bank holidays, wekends and so on where there are no new exchange rates
	url := fmt.Sprintf(
		currencyURL,
		fromCurrency,
		toCurrency,
		date.AddDate(0, 0, -7).Format(common.StdDateFormat),
		date.Format(common.StdDateFormat),
	)
	exchangeRateInfoBody, err := common.GetUrl(ctx, url)
	if err != nil {
		return nil, err
	}
	var exchangeRateInfo ExchangeRateResponse
	err = xml.Unmarshal(exchangeRateInfoBody, &exchangeRateInfo)
	if err != nil {
		return nil, err
	}

	var multiplicator int
	for _, attr := range exchangeRateInfo.DataSet.Series.Attributes.Value {
		if attr.ID == "UNIT_MULT" {
			multiplicator, _ = strconv.Atoi(attr.Value)
		}
	}
	obs := exchangeRateInfo.DataSet.Series.Obs

	return &ExchangeRate{
		Rate: obs[len(obs)-1].ObsValue.Value / math.Pow10(multiplicator),
		Date: obs[len(obs)-1].ObsDimension.Value,
	}, nil
}

type ExchangeRateResponse struct {
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
