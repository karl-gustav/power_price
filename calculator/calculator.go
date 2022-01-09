package calculator

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/karl-gustav/power_price/common"
)

const (
	priceURL         = "https://transparency.entsoe.eu/api?documentType=A44&in_Domain=%s&out_Domain=%s&periodStart=%s2300&periodEnd=%s2300&securityToken=%s"
	entsoeDateFormat = "20060102"
)

type MyFloat float64

func (mf MyFloat) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.4f", float64(mf))), nil
}

type PricePoint struct {
	PriceNOK MyFloat   `json:"NOK_per_kWh"`
	From     time.Time `json:"valid_from"`
	To       time.Time `json:"valid_to"`
}

type Zone string

var Zones = map[string]Zone{
	"NO1": Zone("10YNO-1--------2"),
	"NO2": Zone("10YNO-2--------T"),
	"NO3": Zone("10YNO-3--------J"),
	"NO4": Zone("10YNO-4--------9"),
	"NO5": Zone("10Y1001A1001A48H"),
}

var AvailableZones []string

func init() {
	for zone := range Zones {
		AvailableZones = append(AvailableZones, zone)
	}
}

func CalculatePriceForcast(powerPrices PublicationMarketDocument, exchangeRate float64, loc *time.Location) map[string]PricePoint {
	priceForecast := map[string]PricePoint{}
	for _, price := range powerPrices.TimeSeries.Period.Point {
		pricePerKWh := price.PriceAmount / 1000 // original price is in MWh
		startDate := powerPrices.PeriodTimeInterval.Start.In(loc)
		startOfPeriod := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), price.Position-1, 0, 0, 0, loc)
		endOfPeriod := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), price.Position, 0, 0, 0, loc)
		priceForecast[startOfPeriod.Format(time.RFC3339)] = PricePoint{
			MyFloat(pricePerKWh * exchangeRate),
			startOfPeriod,
			endOfPeriod,
		}
	}
	return priceForecast
}

func GetPrice(zone Zone, date time.Time, token string) (PublicationMarketDocument, error) {
	startDate := date.Add(-24 * time.Hour)
	url := fmt.Sprintf(
		priceURL,
		zone,
		zone,
		startDate.Format(entsoeDateFormat),
		date.Format(entsoeDateFormat),
		token,
	)
	priceBody, err := common.GetUrl(url, []string{token})
	if err != nil {
		return PublicationMarketDocument{}, err
	}
	var powerPrices PublicationMarketDocument
	err = xml.Unmarshal(priceBody, &powerPrices)
	if err != nil {
		return PublicationMarketDocument{}, fmt.Errorf("error unmarshaling price xml: %w\n%s", err, priceBody)
	}
	return powerPrices, nil
}
