package common

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	StdDateFormat = "2006-01-02"
)

var Loc *time.Location

func init() {
	var err error
	Loc, err = time.LoadLocation("Europe/Oslo")
	if err != nil {
		panic(err)
	}
}

func GetUrl(url string, secrets []string) ([]byte, error) {
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
