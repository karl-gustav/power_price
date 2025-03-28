package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/go-chi/chi/v5"
	"github.com/karl-gustav/power_price/calculator"
	"github.com/karl-gustav/power_price/common"
	"github.com/karl-gustav/power_price/currency"
	"github.com/karl-gustav/power_price/storage"
	"github.com/karl-gustav/slogdriver"
)

const (
	missingKeyMessage = "send an email to power@ffail.win to get a free API key"
)

var firstDayInDataset = time.Date(2014, 12, 12, 0, 0, 0, 0, common.Loc)

var SECURITY_TOKEN = os.Getenv("SECURITY_TOKEN")

func init() {
	if slogdriver.OnGCP() {
		projectID, err := metadata.ProjectID()
		if err != nil {
			panic(err)
		}
		slog.SetDefault(slog.New(slogdriver.NewCloudHandler(projectID)))
	} else {
		slog.SetDefault(slog.New(slogdriver.NewLocalHandler()))
	}
}

func main() {
	if SECURITY_TOKEN == "" {
		panic("Envionment variable SECURITY_TOKEN is required!")
	}

	r := chi.NewRouter()
	r.Use(slogdriver.WithTraceContext)
	r.Get("/favicon.ico", notFound)
	r.Get("/", powerPriceHandler)
	r.Get("/graph", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	slog.Info("Serving http://localhost:" + port)
	slog.Error(http.ListenAndServe(":"+port, r).Error())
}

func powerPriceHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	res.Header().Set("Access-Control-Allow-Origin", "*")
	queryZone := req.URL.Query().Get("zone")
	if queryZone == "" {
		m := "\"zone\" query parameter is a required field. Valid zones are NO1, NO2, NO3, NO4 and NO5"
		http.Error(res, m, http.StatusBadRequest)
		return
	}
	zone, ok := calculator.Zones[queryZone]
	if !ok {
		http.Error(
			res,
			queryZone+" is not a valid zone! Valid zones are NO1, NO2, NO3, NO4 and NO5",
			http.StatusBadRequest,
		)
		return
	}

	queryDate := req.URL.Query().Get("date")
	if queryDate == "" {
		m := fmt.Sprintf(
			"\"date\" query parameter is a required field. Date uses this format %s",
			common.StdDateFormat,
		)
		http.Error(res, m, http.StatusBadRequest)
		return
	}
	date, err := time.ParseInLocation(common.StdDateFormat, queryDate, common.Loc)
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
	if date.Before(firstDayInDataset) {
		http.Error(res, "there isn't any price data from before 2014-12-12", http.StatusBadRequest)
		return
	}

	key := req.URL.Query().Get("key")
	if key == "" {
		http.Error(res, "\"key\" query parameter is a required field\n"+missingKeyMessage, http.StatusUnauthorized)
		return
	}
	ok, apiKey, err := storage.GetApiKey(ctx, key)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("got error when getting API key for key `%s`: %v", key, err))
		http.Error(res, "error when verifying api key: "+key, http.StatusInternalServerError)
		return
	} else if !ok {
		slog.WarnContext(ctx, fmt.Sprintf("denied %s access to server because of key was not found", key))
		m := fmt.Sprintf("the key you supplied is not in our systems: %s\n%s", key, missingKeyMessage)
		http.Error(res, m, http.StatusUnauthorized)
		return
	} else if apiKey.Blocked {
		slog.WarnContext(ctx, fmt.Sprintf("denied %s (%s) access to server because of %s", apiKey.Email, key, apiKey.Reason))
		http.Error(res, "You have lost access to server: "+apiKey.Reason, http.StatusForbidden)
		return
	}
	usage, err := storage.GetKeyUsage(ctx, key)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("got error when getting usage for key `%s`: %v", key, err))
		http.Error(res, "error when getting usage for api key: "+key, http.StatusInternalServerError)
		return
	} else if usage.GetZoneCount(queryZone) >= apiKey.Quota {
		slog.WarnContext(ctx, fmt.Sprintf(
			"blocked access for %s because too many requests over quota(%d) in zone %s: %d",
			apiKey.Email,
			apiKey.Quota,
			queryZone,
			usage.GetZoneCount(queryZone),
		),
			slog.String("email", apiKey.Email),
			slog.String("key", key),
		)
		m := fmt.Sprintf(
			"you have exceeded your daily quota of %d requests for zone %s\n"+
				"use https://playground-norway-power.ffail.win for testing your code (unlimited use)",
			apiKey.Quota,
			queryZone,
		)
		http.Error(res, m, http.StatusTooManyRequests)
		err = storage.IncrementKeyUsage(ctx, key, queryZone)
		if err != nil {
			slog.ErrorContext(ctx, "got error when running IncrementKeyUsage():", slog.Any("error", err))
		}
		return
	}

	var priceForecast map[string]calculator.PricePoint
	ok, cache, err := storage.GetCache(ctx, date, zone)
	if !ok {
		slog.DebugContext(ctx, fmt.Sprintf(
			"date/zone %s/%s not found in cache, getting from source",
			date.Format(common.StdDateFormat),
			zone,
		))
	}
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("got error when retreving cache: %v", err))
	}
	if len(cache) != 0 {
		// re-add timezone info because that is lost in firebase
		for key, pricePoint := range cache {
			pricePoint.From = pricePoint.From.In(common.Loc)
			pricePoint.To = pricePoint.To.In(common.Loc)
			cache[key] = pricePoint
		}
		priceForecast = cache
	} else {
		powerPrices, err := calculator.GetPrice(ctx, zone, date, SECURITY_TOKEN)
		if err != nil {
			if errors.Is(calculator.ErrorPricesNotAvialableYet, err) {
				slog.WarnContext(ctx, fmt.Sprintf("got Acknowledgement_MarketDocument for zone %s and date %s", zone, date))
				http.Error(res, err.Error(), http.StatusTooEarly)
				return
			}
			slog.ErrorContext(ctx, fmt.Sprintf("got error when running getPrice(`%s`, `%s`): %v", zone, date, err))
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		exchangeRate, err := currency.GetExchangeRate(ctx, "EUR", "NOK", date)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf(`got error when running getExchangeRate("EUR", "NOK"): %v`, err))
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		priceForecast = calculator.CalculatePriceForcast(ctx, *powerPrices, *exchangeRate)

		err = storage.StoreCache(ctx, date, zone, priceForecast)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("got error when running StoreCache(): %v", err))
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Cache-Control", "public,max-age=31536000,immutable") // 31536000sec --> 1 year
	if err = json.NewEncoder(res).Encode(&priceForecast); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("got error when encoding priceForecast: %ov", err))
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	err = storage.IncrementKeyUsage(ctx, key, queryZone)
	if err != nil {
		slog.ErrorContext(ctx, "got error when running IncrementKeyUsage():", slog.Any("error", err))
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
	return time.Date(year, month, day, 0, 0, 0, 0, common.Loc)
}
