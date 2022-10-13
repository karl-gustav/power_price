package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/karl-gustav/power_price/calculator"
	"github.com/karl-gustav/power_price/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	priceStoragePath  = "power-price/norway-v2"
	apiKeyStoragePath = "power-price/api-keys/users"
	gcpProject        = "my-cloud-collection"
)

type ApiKey struct {
	Email   string `firestore:"email"`
	Blocked bool   `firestore:"blocked"`
	Reason  string `firestore:"reason"`
	Name    string `firestore:"name"`
	Quota   int    `firestore:"quota"`
}

type ZoneUsage struct {
	No1Counter int `firestore:"no1Counter"`
	No2Counter int `firestore:"no2Counter"`
	No3Counter int `firestore:"no3Counter"`
	No4Counter int `firestore:"no4Counter"`
	No5Counter int `firestore:"no5Counter"`
}

func (u *ZoneUsage) GetZoneCount(shortZone string) int {
	switch shortZone {
	case "NO1":
		return u.No1Counter
	case "NO2":
		return u.No2Counter
	case "NO3":
		return u.No3Counter
	case "NO4":
		return u.No4Counter
	case "NO5":
		return u.No5Counter
	default:
		panic("invalid zone sent to GetZoneCount: " + shortZone)
	}
}

func StoreCache(ctx context.Context, day time.Time, zone calculator.Zone, prices map[string]calculator.PricePoint) error {
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return err
	}
	defer client.Close()
	collection := client.Doc(fmt.Sprintf(
		"%s/%s/%s",
		priceStoragePath,
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
	documentRef := client.Doc(fmt.Sprintf(
		"%s/%s/%s",
		priceStoragePath,
		zone,
		day.Format(common.StdDateFormat),
	))
	document, err := documentRef.Get(ctx)
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
	documentRef := client.Doc(fmt.Sprintf(
		"%s/%s",
		apiKeyStoragePath,
		key,
	))
	document, err := documentRef.Get(ctx)
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

func GetKeyUsage(ctx context.Context, key string) (*ZoneUsage, error) {
	date := time.Now().In(common.Loc).Format(common.StdDateFormat)
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	documentRef := client.Doc(fmt.Sprintf(
		"%s/%s/usage/%s",
		apiKeyStoragePath,
		key,
		date,
	))
	usageDoc, err := documentRef.Get(ctx)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return &ZoneUsage{}, nil
		} else {
			return nil, err
		}
	}
	var usage ZoneUsage
	err = usageDoc.DataTo(&usage)
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

func IncrementKeyUsage(ctx context.Context, key, shortZone string) error {
	date := time.Now().In(common.Loc).Format(common.StdDateFormat)
	client, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		return err
	}
	defer client.Close()
	documentRef := client.Doc(fmt.Sprintf(
		"%s/%s/usage/%s",
		apiKeyStoragePath,
		key,
		date,
	))
	_, err = documentRef.Create(ctx, newZoneUsage(shortZone))
	if err != nil {
		if grpc.Code(err) != codes.AlreadyExists {
			return err
		} else {
			_, err = documentRef.Update(ctx, []firestore.Update{
				{
					Path:  strings.ToLower(shortZone) + "Counter",
					Value: firestore.Increment(1),
				},
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func newZoneUsage(shortZone string) ZoneUsage {
	var zoneUsage ZoneUsage
	switch shortZone {
	case "NO1":
		zoneUsage.No1Counter = 1
	case "NO2":
		zoneUsage.No2Counter = 1
	case "NO3":
		zoneUsage.No3Counter = 1
	case "NO4":
		zoneUsage.No4Counter = 1
	case "NO5":
		zoneUsage.No5Counter = 1
	default:
		panic("invalid zone sent to IncrementKeyUsage: " + shortZone)
	}
	return zoneUsage
}
