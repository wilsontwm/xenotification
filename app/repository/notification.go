package repository

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
	"xenotification/app/constant"
	"xenotification/app/model"
	"xenotification/app/types"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// FindNotificationByID :
func (r Repository) FindNotificationByID(id primitive.ObjectID) (*model.Notification, error) {
	v := new(model.Notification)
	if err := r.db.Collection(model.CollectionNotification).FindOne(
		context.Background(),
		bson.M{"_id": id},
	).Decode(v); err != nil {
		return nil, err
	}

	return v, nil
}

// FindNotification :
func (r Repository) FindNotification(typ string, requestID string) (*model.Notification, error) {
	v := new(model.Notification)
	if err := r.db.Collection(model.CollectionNotification).FindOne(
		context.Background(),
		bson.M{"type": typ, "requestId": requestID},
	).Decode(v); err != nil {
		return nil, err
	}

	return v, nil
}

// UpsertNotification :
func (r Repository) UpsertNotification(notification *model.Notification) error {
	_, err := r.db.Collection(model.CollectionNotification).UpdateOne(
		context.Background(),
		bson.M{"_id": notification.ID},
		bson.M{"$set": notification},
		options.Update().SetUpsert(true),
	)
	return err
}

// CreateNotification :
func (r Repository) CreateNotification(notification *model.Notification) error {
	return r.db.Client().UseSession(context.Background(), func(sctx mongo.SessionContext) error {
		if err := sctx.StartTransaction(
			options.Transaction().
				SetReadConcern(readconcern.Snapshot()).
				SetWriteConcern(writeconcern.New(writeconcern.WMajority())),
		); err != nil {
			return err
		}

		if err := r.UpsertNotification(notification); err != nil {
			_ = sctx.AbortTransaction(sctx)
			return err
		}

		// Create the first attempt
		notificationAttempt := new(model.NotificationAttempt)
		notificationAttempt.ID = primitive.NewObjectID()
		notificationAttempt.NotificationID = notification.ID
		notificationAttempt.MerchantID = notification.MerchantID
		notificationAttempt.AttemptNo = 1
		notificationAttempt.Status = types.NotificationStatusPending
		notificationAttempt.CreatedAt = time.Now().UTC()
		notificationAttempt.UpdatedAt = time.Now().UTC()

		if err := r.UpsertNotificationAttempt(notificationAttempt); err != nil {
			_ = sctx.AbortTransaction(sctx)
			return err
		}

		for {
			err := sctx.CommitTransaction(sctx)
			switch e := err.(type) {
			case nil:
				return nil
			case mongo.CommandError:
				return e
			default:
				return e
			}
		}
	})
}

// FindRetryNotifications :
func (r Repository) FindRetryNotifications(cursor string) ([]*model.Notification, string, error) {
	notifications := make([]*model.Notification, 0)

	ctx := context.Background()
	query := bson.M{
		"status":    types.NotificationStatusFailed,
		"attemptNo": bson.M{"$lt": constant.RetryAttemptCount},
		"$or": bson.A{
			bson.M{"attemptedAt": bson.M{"$exists": false}},
			bson.M{"attemptedAt": nil},
			bson.M{"attemptedAt": bson.M{"$lte": time.Now().Add(-1 * constant.RetryAttemptDuration)}},
		},
	}

	var limit int64 = 50

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

	nextCursor, err := r.db.Collection(model.CollectionNotification).Find(
		ctx,
		query,
		options.Find().SetLimit(limit+1).SetSkip(currentSkip),
	)

	if err != nil {
		return nil, "", err
	}
	defer nextCursor.Close(ctx)

	for nextCursor.Next(ctx) {
		notification := new(model.Notification)
		if err := nextCursor.Decode(notification); err != nil {
			return nil, "", errors.New("entity decode error")
		}
		notifications = append(notifications, notification)
	}

	if err := nextCursor.Err(); err != nil {
		return nil, "", err
	}

	if len(notifications) > int(limit) {
		return notifications[:len(notifications)-1], hex.EncodeToString([]byte(fmt.Sprintf("%d", currentSkip+limit))), nil
	}

	return notifications, "", nil
}
