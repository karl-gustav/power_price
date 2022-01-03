package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	priceURL         = "https://transparency.entsoe.eu/api?documentType=A44&in_Domain=%s&out_Domain=%s&periodStart=%s2300&periodEnd=%s2300&securityToken=%s"
	currencyURL      = "https://data.norges-bank.no/api/data/EXR/M.%s.%s.SP?lastNObservations=1"
	stdDateFormat    = "2006-01-02"
	entsoeDateFormat = "20060102"
)

var zones = map[string]Zone{
	"NO1": Zone("10YNO-1--------2"),
	"NO2": Zone("10YNO-2--------T"),
	"NO3": Zone("10YNO-3--------J"),
	"NO4": Zone("10YNO-4--------9"),
	"NO5": Zone("10Y1001A1001A48H"),
}

var availableZones []string
var loc *time.Location

func init() {
	for zone := range zones {
		availableZones = append(availableZones, zone)
	}
	var err error
	loc, err = time.LoadLocation("Europe/Oslo")
	if err != nil {
		panic(err)
	}
}

var SECURITY_TOKEN = os.Getenv("SECURITY_TOKEN")

type Zone string

type MyFloat float64

func (mf MyFloat) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.4f", float64(mf))), nil
}

type PricePoint struct {
	PriceNOK MyFloat   `json:"NOK_per_kWh"`
	From     time.Time `json:"valid_from"`
	To       time.Time `json:"valid_to"`
}

func main() {
	if SECURITY_TOKEN == "" {
		panic("Envionment variable SECURITY_TOKEN is required!")
	}

	http.HandleFunc("/", powerPriceHandler)
	http.HandleFunc("/graph", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func powerPriceHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	res.Header().Set("Access-Control-Allow-Origin", "*")
	queryZones, ok := req.URL.Query()["zone"]
	if !ok || len(queryZones[0]) < 1 {
		http.Error(res, "\"zone\" query parameter is a required field", http.StatusBadRequest)
		return
	}
	queryZone := strings.ToUpper(queryZones[0])
	zone, ok := zones[queryZone]

	if !ok {
		http.Error(res, fmt.Sprintf(
			"%s is not a valid zone! Valid zones are %s",
			queryZone,
			strings.Join(availableZones, ", "),
		), http.StatusBadRequest)
		return
	}
	var date time.Time
	queryDates, ok := req.URL.Query()["date"]
	if !ok || len(queryDates) < 1 {
		http.Error(res, "\"date\" query parameter is a required field", http.StatusBadRequest)
		return
	}
	date, err := time.Parse(stdDateFormat, queryDates[0])
	if err != nil {
		http.Error(
			res,
			fmt.Sprintf("Could not parse %s, in the format %s", queryDates[0], stdDateFormat),
			http.StatusBadRequest,
		)
		return
	}

	var priceForecast map[string]PricePoint
	cache, err := GetCache(ctx, date, zone)
	if err != nil {
		fmt.Printf("got error when retreving cache: %v", err)
	}
	if len(cache) != 0 {
		// re-add timezone info because that is lost in firebase
		for key, pricePoint := range cache {
			pricePoint.From = pricePoint.From.In(loc)
			pricePoint.To = pricePoint.To.In(loc)
			cache[key] = pricePoint
		}
		priceForecast = cache
	} else {
		powerPrices, err := getPrice(Zone(zone), date)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		exchangeRate, err := getExchangeRate("EUR", "NOK")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		priceForecast = calculatePriceForcast(powerPrices, exchangeRate)

		err = StoreCache(ctx, date, zone, priceForecast)
		if err != nil {
			panic(err)
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", "public,max-age=31536000,immutable") // 31536000sec --> 1 year
	if err = json.NewEncoder(res).Encode(priceForecast); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

func calculatePriceForcast(powerPrices PublicationMarketDocument, exchangeRate float64) map[string]PricePoint {
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

func getExchangeRate(from, to string) (float64, error) {
	url := fmt.Sprintf(currencyURL, from, to)
	exchangeRateInfoBody, err := getUrl(url, []string{})
	if err != nil {
		return 0, err
	}
	var exchangeRateInfo ExchangeRateInfo
	err = xml.Unmarshal(exchangeRateInfoBody, &exchangeRateInfo)
	if err != nil {
		return 0, err
	}

	return exchangeRateInfo.DataSet.Series.Obs.OBSVALUE / math.Pow10(exchangeRateInfo.DataSet.Series.UNITMULT), nil
}

func getPrice(zone Zone, date time.Time) (PublicationMarketDocument, error) {
	startDate := date.Add(-24 * time.Hour)
	url := fmt.Sprintf(
		priceURL,
		zone,
		zone,
		startDate.Format(entsoeDateFormat),
		date.Format(entsoeDateFormat),
		SECURITY_TOKEN,
	)
	priceBody, err := getUrl(url, []string{SECURITY_TOKEN})
	if err != nil {
		return PublicationMarketDocument{}, err
	}
	var powerPrices PublicationMarketDocument
	err = xml.Unmarshal(priceBody, &powerPrices)
	if err != nil {
		return PublicationMarketDocument{}, fmt.Errorf("error unmarshaling price xml: %w", err)
	}
	return powerPrices, nil
}

func getUrl(url string, secrets []string) ([]byte, error) {
	resp, err := http.Get(url)
	for _, secret := range secrets {
		url = strings.ReplaceAll(url, secret, "***secret***")
	}
	if err != nil {
		return nil, fmt.Errorf("Couldn't make GET request to %s:\n%v", url, err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("None 200 response code %v from %s:\n%s", resp.StatusCode, url, body)
	}
	return body, nil
}
