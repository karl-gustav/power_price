package calculator

import (
	"encoding/xml"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/karl-gustav/power_price/common"
	"github.com/karl-gustav/power_price/currency"
)

var exchangeRate = currency.ExchangeRate{
	Rate: 10,
}

func Test60mResolution(t *testing.T) {
	first60MinDate := "2025-01-22T00:00:00+01:00"
	first60MinPrice := 47.14
	last60MinDate := "2025-01-22T23:00:00+01:00"
	last60MinPrice := 58.51
	xml60min, err := os.ReadFile("./testdata/60m.xml")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	var powerPricesXML PublicationMarketDocument
	err = xml.Unmarshal(xml60min, &powerPricesXML)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	powerPrices := CalculatePriceForcast(powerPricesXML, exchangeRate)

	var keys []string
	for key := range powerPrices {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	if keys[0] != first60MinDate {
		t.Errorf("first date is expected to be %s, was %s", first60MinDate, keys[0])
	}
	if keys[len(keys)-1] != last60MinDate {
		t.Errorf("last date is expected to be %s, was %s", last60MinDate, keys[len(keys)-1])
	}
	if len(powerPrices) != 24 {
		t.Errorf("expected 24 values in parsed power prices was %d", len(powerPrices))
	}

	first60MinFrom, _ := time.Parse(time.RFC3339, first60MinDate)
	first60MinTo := first60MinFrom.Add(1 * time.Hour)
	if !powerPrices[first60MinDate].From.Equal(first60MinFrom) {
		t.Errorf("expected first valid_from to be %s, was %s", first60MinFrom.In(common.Loc), powerPrices[first60MinDate].From)
	}
	if !powerPrices[first60MinDate].To.Equal(first60MinTo) {
		t.Errorf("expected first valid_to to be %s, was %s", first60MinTo.In(common.Loc), powerPrices[first60MinDate].To)
	}

	last60MinFrom, _ := time.Parse(time.RFC3339, last60MinDate)
	last60MinTo := last60MinFrom.Add(1 * time.Hour)
	if !powerPrices[last60MinDate].From.Equal(last60MinFrom) {
		t.Errorf("expected last valid_from to be %s, was %s", last60MinFrom.In(common.Loc), powerPrices[last60MinDate].From)
	}
	if !powerPrices[last60MinDate].To.Equal(last60MinTo) {
		t.Errorf("expected last valid_to to be %s, was %s", last60MinTo.In(common.Loc), powerPrices[last60MinDate].To)
	}

	if powerPrices[first60MinDate].PriceMWhEUR != first60MinPrice {
		t.Errorf("expected price to be %f, was %f", first60MinPrice, powerPrices[first60MinDate].PriceMWhEUR)
	}
	if powerPrices[last60MinDate].PriceMWhEUR != last60MinPrice {
		t.Errorf("expected price to be %f, was %f", last60MinPrice, powerPrices[last60MinDate].PriceMWhEUR)
	}
}

func Test15mResolution(t *testing.T) {
	first15MinDate := "2025-02-23T00:00:00+01:00"
	first15MinPrice := 48.74
	last15MinDate := "2025-02-23T23:00:00+01:00"
	last15MinPrice := 47.39
	xml15min, err := os.ReadFile("./testdata/15m.xml")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	var powerPricesXML PublicationMarketDocument
	err = xml.Unmarshal(xml15min, &powerPricesXML)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	powerPrices := CalculatePriceForcast(powerPricesXML, exchangeRate)

	var keys []string
	for key := range powerPrices {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	if keys[0] != first15MinDate {
		t.Errorf("first date is expected to be %s, was %s", first15MinDate, keys[0])
	}
	if keys[len(keys)-1] != last15MinDate {
		t.Errorf("last date is expected to be %s, was %s", last15MinDate, keys[len(keys)-1])
	}
	if len(powerPrices) != 24 {
		t.Errorf("expected 24 values in parsed power prices was %d", len(powerPrices))
	}

	first15MinFrom, _ := time.Parse(time.RFC3339, first15MinDate)
	first15MinTo := first15MinFrom.Add(1 * time.Hour)
	if !powerPrices[first15MinDate].From.Equal(first15MinFrom) {
		t.Errorf("expected valid_from to be %s, was %s", first15MinFrom.In(common.Loc), powerPrices[first15MinDate].From)
	}
	if !powerPrices[first15MinDate].To.Equal(first15MinTo) {
		t.Errorf("expected valid_to to be %s, was %s", first15MinTo.In(common.Loc), powerPrices[first15MinDate].To)
	}

	last15MinFrom, _ := time.Parse(time.RFC3339, last15MinDate)
	last15MinTo := last15MinFrom.Add(1 * time.Hour)
	if !powerPrices[last15MinDate].From.Equal(last15MinFrom) {
		t.Errorf("expected first valid_from to be %s, was %s", last15MinFrom.In(common.Loc), powerPrices[last15MinDate].From)
	}
	if !powerPrices[last15MinDate].To.Equal(last15MinTo) {
		t.Errorf("expected first valid_to to be %s, was %s", last15MinTo.In(common.Loc), powerPrices[last15MinDate].To)
	}

	if powerPrices[first15MinDate].PriceMWhEUR != first15MinPrice {
		t.Errorf("expected last price to be %f, was %f", first15MinPrice, powerPrices[first15MinDate].PriceMWhEUR)
	}
	if powerPrices[last15MinDate].PriceMWhEUR != last15MinPrice {
		t.Errorf("expected last price to be %f, was %f", last15MinPrice, powerPrices[last15MinDate].PriceMWhEUR)
	}
}
