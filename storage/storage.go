package storage

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/karl-gustav/power_price/calculator"
	"github.com/karl-gustav/power_price/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	priceStorageKey  = "power-price/norway/%s/%s"
	apiKeyStorageKey = "power-price/api-keys/users/%s"
	gcpProject       = "my-cloud-collection"
)

type ApiKey struct {
	Email   string `firestore:"email"`
	Blocked bool   `firestore:"blocked"`
	Reason  string `firestore:"reason"`
	Name    string `firestore:"name"`
}

func StoreCache(ctx context.Context, day time.Time, zone calculator.Zone, prices map[string]calculator.PricePoint) error {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return err
	}
	defer client.Close()
	collection := client.Doc(fmt.Sprintf(
		priceStorageKey,
		zone,
		day.Format(common.StdDateFormat),
	))

	_, err = collection.Set(ctx, prices)
	return err
}

func GetCache(ctx context.Context, day time.Time, zone calculator.Zone) (ok bool, pricepoints map[string]calculator.PricePoint, err error) {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return false, nil, err
	}
	defer client.Close()
	containerRef := client.Doc(fmt.Sprintf(
		priceStorageKey,
		zone,
		day.Format(common.StdDateFormat),
	))
	document, err := containerRef.Get(ctx)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return false, nil, nil
		}
		return false, nil, err
	}
	container := make(map[string]calculator.PricePoint)
	err = document.DataTo(&container)
	if err != nil {
		return false, nil, err
	}
	return true, container, nil
}

func GetApiKey(ctx context.Context, key string) (ok bool, apiKey *ApiKey, err error) {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return false, nil, err
	}
	defer client.Close()
	containerRef := client.Doc(fmt.Sprintf(
		apiKeyStorageKey,
		key,
	))
	document, err := containerRef.Get(ctx)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return false, nil, nil
		}
		return false, nil, err
	}
	err = document.DataTo(&apiKey)
	if err != nil {
		return false, nil, err
	}
	return true, apiKey, nil
}
