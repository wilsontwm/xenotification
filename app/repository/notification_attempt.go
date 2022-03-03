package repository

import (
	"context"
	"xenotification/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindLastNotificationAttempt :
func (r Repository) FindLastNotificationAttempt(notificationID primitive.ObjectID) (*model.NotificationAttempt, error) {
	v := new(model.NotificationAttempt)
	if err := r.db.Collection(model.CollectionNotificationAttempt).FindOne(
		context.Background(),
		bson.M{"notificationId": notificationID},
		options.FindOne().SetSort(bson.M{"attemptNo": -1}),
	).Decode(v); err != nil {
		return nil, err
	}

	return v, nil
}

// UpsertNotificationAttempt :
func (r Repository) UpsertNotificationAttempt(att *model.NotificationAttempt) error {
	_, err := r.db.Collection(model.CollectionNotificationAttempt).UpdateOne(
		context.Background(),
		bson.M{"_id": att.ID},
		bson.M{"$set": att},
		options.Update().SetUpsert(true),
	)
	return err
}
