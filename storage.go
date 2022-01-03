package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
)

const (
	masterStorageKey = "power-price/norway"
	gcpProject       = "my-cloud-collection"
)

func StoreCache(ctx context.Context, day time.Time, zone Zone, prices map[string]PricePoint) error {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return err
	}
	collection := client.Doc(fmt.Sprintf(
		"%s/%s/%s",
		masterStorageKey,
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
	containerRef := client.Doc(fmt.Sprintf(
		"%s/%s/%s",
		masterStorageKey,
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
