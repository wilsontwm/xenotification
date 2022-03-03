package model

import (
	"time"
	"xenotification/app/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationAttempt :
type NotificationAttempt struct {
	ID             primitive.ObjectID       `bson:"_id" json:"_id"`
	NotificationID primitive.ObjectID       `bson:"notificationId" json:"notificationId"`
	MerchantID     string                   `bson:"merchantId" json:"merchantId"`
	AttemptNo      uint                     `bson:"attemptNo" json:"attemptNo"`
	Status         types.NotificationStatus `bson:"status" json:"status"`
	StatusCode     int                      `bson:"statusCode" json:"statusCode"`
	Error          *string                  `bson:"error" json:"error"`
	SentAt         *time.Time               `bson:"sentAt" json:"sentAt"`
	Model          `bson:",inline"`
}
