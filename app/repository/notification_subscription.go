package repository

import (
	"context"
	"xenotification/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
