package calculator

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/karl-gustav/power_price/common"
	"github.com/karl-gustav/power_price/currency"
)

const (
	priceURL         = "https://web-api.tp.entsoe.eu/api?documentType=A44&in_Domain=%s&out_Domain=%s&periodStart=%s&periodEnd=%s&securityToken=%s"
	entsoeDateFormat = "200601021504"
)

var ErrorPricesNotAvialableYet = errors.New(`The prices was not found on the transparency.entsoe.eu server.
Try again later or check https://transparency.entsoe.eu/news/widget if there are any delays.`)

type PricePoint struct {
	PriceKWhNOK      float64   `json:"NOK_per_kWh" firestore:"PriceKWhNOK"`
	PriceMWhEUR      float64   `json:"EUR_per_MWh" firestore:"PriceMWhEUR"`
	ExchangeRate     float64   `json:"exchange_rate" firestore:"ExchangeRate"`
	ExchangeRateDate string    `json:"exchange_rate_date" firestore:"ExchangeRateDate"`
	From             time.Time `json:"valid_from" firestore:"From"`
	To               time.Time `json:"valid_to" firestore:"To"`
}

// not using a pointer here because this is used as a value type in a map
func (p PricePoint) MarshalJSON() ([]byte, error) {
	type Alias PricePoint
	alias := Alias(p)
	alias.PriceKWhNOK = round(alias.PriceKWhNOK, 4)
	return json.Marshal(alias)
}

func round(number, decimalPlaces float64) float64 {
	return math.Round(number*math.Pow(10, decimalPlaces)) / math.Pow(10, decimalPlaces)
}

type Zone string

var Zones = map[string]Zone{
	"NO1": Zone("10YNO-1--------2"),
	"NO2": Zone("10YNO-2--------T"),
	"NO3": Zone("10YNO-3--------J"),
	"NO4": Zone("10YNO-4--------9"),
	"NO5": Zone("10Y1001A1001A48H"),
}

func CalculatePriceForcast(powerPrices PublicationMarketDocument, exchangeRate currency.ExchangeRate) map[string]PricePoint {
	priceForecast := map[string]PricePoint{}
	startDate := powerPrices.PeriodTimeInterval.Start.In(common.Loc)
	for _, price := range powerPrices.TimeSeries.Period.Point {
		priceMWhEUR := price.PriceAmount
		priceMWhNOK := priceMWhEUR * exchangeRate.Rate
		priceKWhNOK := priceMWhNOK / 1000

		startOfPeriod := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), price.Position-1, 0, 0, 0, common.Loc)
		endOfPeriod := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), price.Position, 0, 0, 0, common.Loc)
		priceForecast[startOfPeriod.Format(time.RFC3339)] = PricePoint{
			PriceKWhNOK:      priceKWhNOK,
			PriceMWhEUR:      priceMWhEUR,
			ExchangeRate:     exchangeRate.Rate,
			ExchangeRateDate: exchangeRate.Date,
			From:             startOfPeriod,
			To:               endOfPeriod,
		}
		if price.Position == 24 {
			startDate = startDate.AddDate(0, 0, 1)
		}
	}
	return priceForecast
}

func GetPrice(zone Zone, date time.Time, token string) (*PublicationMarketDocument, error) {
	endDate := date.Add(24 * time.Hour)
	url := fmt.Sprintf(
		priceURL,
		zone,
		zone,
		date.In(time.UTC).Format(entsoeDateFormat),
		endDate.In(time.UTC).Format(entsoeDateFormat),
		token,
	)
	priceBody, err := common.GetUrl(url, token)
	if err != nil {
		return nil, err
	}
	if bytes.Contains(priceBody, []byte("<Acknowledgement_MarketDocument")) {
		return nil, ErrorDayAheadPricesNotFound
	}

	var powerPrices PublicationMarketDocument
	err = xml.Unmarshal(priceBody, &powerPrices)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling price xml: %w\n%.4000s", err, priceBody)
	}
	return &powerPrices, nil
}
