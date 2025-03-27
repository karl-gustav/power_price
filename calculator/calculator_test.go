package calculator

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/karl-gustav/power_price/currency"
)

var exchangeRate = currency.ExchangeRate{
	Rate: 1,
}

func Test60m(t *testing.T) {
	prices := []float64{47.14, 40.6, 40.64, 40.75, 41.16, 50.12, 122.94, 200.98, 224, 193.92, 173.51, 167.58, 160.22, 165.07, 174.7, 189.99, 193.27, 202.93, 175.19, 162.23, 129.99, 123.82, 103.58, 58.51}
	xmlData, err := os.ReadFile("./testdata/60m.xml")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	var powerPricesXML PublicationMarketDocument
	err = xml.Unmarshal(xmlData, &powerPricesXML)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	powerPrices := CalculatePriceForcast(context.Background(), powerPricesXML, exchangeRate)

	for hour := range 24 {
		tsString := fmt.Sprintf("2025-01-22T%02d:00:00+01:00", hour)
		startTime, _ := time.Parse(time.RFC3339, tsString)
		endTime := startTime.Add(1 * time.Hour)
		pricePoint := powerPrices[tsString]
		if pricePoint.PriceMWhEUR != prices[hour] {
			t.Errorf("expected the price for %s to be %f, but it was %f", tsString, prices[hour], pricePoint.PriceMWhEUR)
		}
		if pricePoint.PriceKWhNOK != prices[hour]/1000 {
			t.Errorf("expected prices kwh nok to be %f, was %f", pricePoint.PriceKWhNOK, prices[hour])
		}
		if !pricePoint.From.Equal(startTime) {
			t.Errorf("expected start time to be %s, was %s", pricePoint.From, startTime)
		}
		if !pricePoint.To.Equal(endTime) {
			t.Errorf("expected end time to be %s, was %s", pricePoint.To, endTime)
		}
		if pricePoint.ExchangeRate != exchangeRate.Rate {
			t.Errorf("expected exchange rate to be %f, was %f", exchangeRate.Rate, pricePoint.ExchangeRate)
		}
	}
}

func Test15m(t *testing.T) {
	prices := []float64{48.74, 48.66, 48.58, 48.56, 48.64, 48.84, 49.5, 49.92, 50.81, 51.79, 51.25, 50.58, 48.27, 47.41, 47.44, 49.14, 51.15, 55.02, 54.85, 53.46, 51.13, 49.74, 48.83, 47.39}
	xmlData, err := os.ReadFile("./testdata/15m.xml")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	var powerPricesXML PublicationMarketDocument
	err = xml.Unmarshal(xmlData, &powerPricesXML)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	powerPrices := CalculatePriceForcast(context.Background(), powerPricesXML, exchangeRate)

	for hour := range 24 {
		tsString := fmt.Sprintf("2025-02-23T%02d:00:00+01:00", hour)
		startTime, _ := time.Parse(time.RFC3339, tsString)
		endTime := startTime.Add(1 * time.Hour)
		pricePoint := powerPrices[tsString]
		if pricePoint.PriceMWhEUR != prices[hour] {
			t.Errorf("expected the price for %s to be %f, but it was %f", tsString, prices[hour], pricePoint.PriceMWhEUR)
		}
		if pricePoint.PriceKWhNOK != prices[hour]/1000 {
			t.Errorf("expected prices kwh nok to be %f, was %f", pricePoint.PriceKWhNOK, prices[hour])
		}
		if !pricePoint.From.Equal(startTime) {
			t.Errorf("expected start time to be %s, was %s", pricePoint.From, startTime)
		}
		if !pricePoint.To.Equal(endTime) {
			t.Errorf("expected end time to be %s, was %s", pricePoint.To, endTime)
		}
		if pricePoint.ExchangeRate != exchangeRate.Rate {
			t.Errorf("expected exchange rate to be %f, was %f", exchangeRate.Rate, pricePoint.ExchangeRate)
		}
	}
}
