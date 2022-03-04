package model

import (
	"time"
	"xenotification/app/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Notification :
type Notification struct {
	ID              primitive.ObjectID       `bson:"_id" json:"_id"`
	MerchantID      string                   `bson:"merchantId" json:"merchantId"`
	RequestID       string                   `bson:"requestId" json:"requestId"`
	Type            string                   `bson:"type" json:"type"`
	Payload         interface{}              `bson:"payload" json:"payload"`
	NotificationURL string                   `bson:"notificationUrl" json:"notificationUrl"`
	NotificationKey string                   `bson:"notificationKey" json:"notificationKey"`
	AttemptNo       uint                     `bson:"attemptNo" json:"attemptNo"`
	AttemptedAt     *time.Time               `bson:"attemptedAt" json:"attempedAt"`
	Status          types.NotificationStatus `bson:"status" json:"status"`
	IsSimulation    bool                     `bson:"-" json:"-"`
	Model           `bson:",inline"`
}
