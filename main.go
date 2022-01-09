package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/karl-gustav/power_price/calculator"
	"github.com/karl-gustav/power_price/common"
	"github.com/karl-gustav/power_price/currency"
	"github.com/karl-gustav/power_price/storage"
	"github.com/karl-gustav/runlogger"
)

const (
	missingKeyMessage = "send an email to ffaildotwin@gmail.com to get a free API key"
)

var (
	loc *time.Location
	log *runlogger.Logger
)

func init() {
	if os.Getenv("K_SERVICE") != "" { // Check if running in cloud run
		log = runlogger.StructuredLogger()
	} else {
		log = runlogger.PlainLogger()
	}
	var err error
	loc, err = time.LoadLocation("Europe/Oslo")
	if err != nil {
		panic(err)
	}
}

var SECURITY_TOKEN = os.Getenv("SECURITY_TOKEN")

func main() {
	if SECURITY_TOKEN == "" {
		panic("Envionment variable SECURITY_TOKEN is required!")
	}

	http.HandleFunc("/favicon.ico", notFound)
	http.HandleFunc("/", powerPriceHandler)
	http.HandleFunc("/graph", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Info("Serving http://localhost:" + port)
	log.Critical(http.ListenAndServe(":"+port, nil))
}

func powerPriceHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	res.Header().Set("Access-Control-Allow-Origin", "*")
	queryZone := req.URL.Query().Get("zone")
	if queryZone == "" {
		http.Error(res, "\"zone\" query parameter is a required field", http.StatusBadRequest)
		return
	}
	zone, ok := calculator.Zones[queryZone]
	if !ok {
		http.Error(res, fmt.Sprintf(
			"%s is not a valid zone! Valid zones are %s",
			queryZone,
			strings.Join(calculator.AvailableZones, ", "),
		), http.StatusBadRequest)
		return
	}

	key := req.URL.Query().Get("key")
	if key == "" {
		http.Error(res, "\"key\" query parameter is a required field\n"+missingKeyMessage, http.StatusUnauthorized)
		return
	}
	ok, apiKey, err := storage.GetApiKey(ctx, key)
	if err != nil {
		log.Errorf("got error when getting api key for key `%s`: %v", key, err)
		http.Error(res, "error when verifying api key: "+key, http.StatusInternalServerError)
		return
	} else if !ok {
		log.Warningf("denied %s access to server because of key was not found", key)
		m := fmt.Sprintf("the key you supplied is not in our systems: %s\n%s", key, missingKeyMessage)
		http.Error(res, m, http.StatusUnauthorized)
		return
	} else if apiKey.Blocked {
		log.Warningf("denied %s (%s) access to server because of %s", apiKey.Email, key, apiKey.Reason)
		http.Error(res, "You have lost access to server: "+apiKey.Reason, http.StatusForbidden)
		return
	}
	queryDate := req.URL.Query().Get("date")
	if queryDate == "" {
		http.Error(res, "\"date\" query parameter is a required field", http.StatusBadRequest)
		return
	}
	date, err := time.ParseInLocation(common.StdDateFormat, queryDate, loc)
	if err != nil {
		http.Error(
			res,
			fmt.Sprintf("Could not parse %s, in the format %s", queryDate, common.StdDateFormat),
			http.StatusBadRequest,
		)
		return
	}
	if !isValidTimePeriod(date) {
		http.Error(res, "price data only become available at 14:00 for the next day", http.StatusBadRequest)
		return
	}

	var priceForecast map[string]calculator.PricePoint
	cache, err := storage.GetCache(ctx, date, zone)
	if err != nil {
		log.Debugf("got error when retreving cache: %v", err)
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
		powerPrices, err := calculator.GetPrice(calculator.Zone(zone), date, SECURITY_TOKEN)
		if err != nil {
			log.Errorf("got error when running getPrice(`%s`, `%s`): %v", zone, date, err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		exchangeRate, err := currency.GetExchangeRate("EUR", "NOK", date)
		if err != nil {
			log.Errorf(`got error when running getExchangeRate("EUR", "NOK"): %v`, err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		priceForecast = calculator.CalculatePriceForcast(powerPrices, exchangeRate, loc)

		err = storage.StoreCache(ctx, date, zone, priceForecast)
		if err != nil {
			log.Errorf("got error when running StoreCache(): %v", err)
			panic(err)
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", "public,max-age=31536000,immutable") // 31536000sec --> 1 year
	if err = json.NewEncoder(res).Encode(priceForecast); err != nil {
		log.Errorf("got error when encoding priceForecast: %ov", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

func notFound(res http.ResponseWriter, req *http.Request) {
	http.Error(res, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func isValidTimePeriod(date time.Time) bool {
	now := time.Now()
	startOfDay := getStartOfDay(now)
	tomorrow := startOfDay.Add(time.Hour * 24)
	if date.Before(tomorrow) {
		return true
	} else if date.Equal(tomorrow) {
		return now.After(startOfDay.Add(time.Hour * 14))
	}
	return false
}

func getStartOfDay(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}
