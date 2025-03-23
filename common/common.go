package common

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
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

func GetUrl(ctx context.Context, url string, secrets ...string) ([]byte, error) {
	resp, err := http.Get(url)
	for _, secret := range secrets {
		url = strings.ReplaceAll(url, secret, "***secret***")
	}
	slog.InfoContext(ctx, fmt.Sprintf("Making GET request for %s", url), slog.String("url", url))
	if err != nil {
		for _, secret := range secrets {
			err = errors.New(strings.ReplaceAll(err.Error(), secret, "***secret***"))
		}
		return nil, fmt.Errorf("Couldn't make GET request to %s:\n%v", url, err)
	}
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("None 200 response code %v from %s:\n%s", resp.StatusCode, url, body)
	}
	return body, nil
}
