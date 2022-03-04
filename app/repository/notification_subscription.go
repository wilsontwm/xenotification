package repository

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"xenotification/app/model"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindNotificationSubscriptions :
func (r Repository) FindNotificationSubscriptions(merchantID string, cursor string, limit int64) ([]*model.NotificationSubscription, string, error) {
	notificationSubs := make([]*model.NotificationSubscription, 0)

	ctx := context.Background()
	query := bson.M{
		"_id.merchantId": merchantID,
	}

	if limit <= 0 {
		limit = 50
	}

	currentSkip := int64(0)

	if cursor != "" {
		data, err := hex.DecodeString(cursor)
		if err != nil {
			return nil, "", err
		}

		currentSkip, err = strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return nil, "", err
		}
	}

	nextCursor, err := r.db.Collection(model.CollectionNotificationSubscription).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSkip(currentSkip),
	)

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		notificationSub := new(model.NotificationSubscription)
		if err := nextCursor.Decode(notificationSub); err != nil {
			return nil, "", errors.New("entity decode error")
		}
		notificationSubs = append(notificationSubs, notificationSub)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(notificationSubs) > int(limit) {
		return notificationSubs[:len(notificationSubs)-1], hex.EncodeToString([]byte(fmt.Sprintf("%d", currentSkip+limit))), nil
	}

	return notificationSubs, "", nil
}

// FindNotificationSubscription :
func (r Repository) FindNotificationSubscription(id model.SubscriptionKey) (*model.NotificationSubscription, error) {
	v := new(model.NotificationSubscription)
	if err := r.db.Collection(model.CollectionNotificationSubscription).FindOne(
		context.Background(),
		bson.M{"_id": id},
	).Decode(v); err != nil {
		return nil, err
	}

	return v, nil
}

// UpsertNotificationSubscription :
func (r Repository) UpsertNotificationSubscription(sub *model.NotificationSubscription) error {
	_, err := r.db.Collection(model.CollectionNotificationSubscription).UpdateOne(
		context.Background(),
		bson.M{"_id": sub.ID},
		bson.M{"$set": sub},
		options.Update().SetUpsert(true),
	)
	return err
}

// DeleteNotificationSubscription :
func (r Repository) DeleteNotificationSubscription(id model.SubscriptionKey) error {
	_, err := r.db.Collection(model.CollectionNotificationSubscription).DeleteOne(
		context.Background(),
		bson.M{"_id": id},
		options.Delete(),
	)
	return err
}
