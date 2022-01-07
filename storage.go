package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
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
}

func StoreCache(ctx context.Context, day time.Time, zone Zone, prices map[string]PricePoint) error {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return err
	}
	defer client.Close()
	collection := client.Doc(fmt.Sprintf(
		priceStorageKey,
		zone,
		day.Format(stdDateFormat),
	))

	_, err = collection.Set(ctx, prices)
	return err
}

func GetCache(ctx context.Context, day time.Time, zone Zone) (map[string]PricePoint, error) {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	containerRef := client.Doc(fmt.Sprintf(
		priceStorageKey,
		zone,
		day.Format(stdDateFormat),
	))
	document, err := containerRef.Get(ctx)
	if err != nil {
		return nil, err
	}
	container := make(map[string]PricePoint)
	err = document.DataTo(&container)
	if err != nil {
		return nil, err
	}
	return container, nil
}

func GetApiKey(ctx context.Context, key string) (ok bool, apiKey *ApiKey, err error) {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return false, nil, err
	}
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
