package transformer

import (
	"log"
	"time"

	"xenotification/app/model"
	"xenotification/app/types"
)

// Notification :
type Notification struct {
	ID              string                   `json:"id"`
	MerchantID      string                   `json:"merchantId"`
	Type            string                   `json:"type"`
	NotificationURL string                   `json:"notificationUrl"`
	NotificationKey string                   `json:"notificationKey"`
	RequestID       string                   `json:"requestID"`
	Payload         interface{}              `json:"payload"`
	Status          types.NotificationStatus `json:"status"`
	AttemptNo       uint                     `json:"attemptNo"`
	SentAt          *time.Time               `json:"sentAt,omitempty"`
	CreatedAt       time.Time                `json:"createdAt"`
	UpdatedAt       time.Time                `json:"updatedAt"`
}

// NotificationWithAttempt :
type NotificationWithAttempt struct {
	ID              string                   `json:"id"`
	MerchantID      string                   `json:"merchantId"`
	Type            string                   `json:"type"`
	NotificationURL string                   `json:"notificationUrl"`
	NotificationKey string                   `json:"notificationKey"`
	RequestID       string                   `json:"requestID"`
	Payload         interface{}              `json:"payload"`
	Status          types.NotificationStatus `json:"status"`
	StatusCode      int                      `json:"statusCode"`
	AttemptNo       uint                     `json:"attemptNo"`
	SentAt          *time.Time               `json:"sentAt,omitempty"`
	CreatedAt       time.Time                `json:"createdAt"`
	UpdatedAt       time.Time                `json:"updatedAt"`
}

// ToNotification :
func ToNotification(i *model.Notification) (o Notification) {
	o.ID = i.ID.Hex()
	o.MerchantID = i.MerchantID
	o.Type = i.Type
	o.NotificationURL = i.NotificationURL
	o.NotificationKey = i.NotificationKey
	o.RequestID = i.RequestID
	o.Payload = i.Payload
	o.Status = i.Status
	o.AttemptNo = i.AttemptNo
	o.SentAt = i.AttemptedAt
	o.CreatedAt = i.CreatedAt
	o.UpdatedAt = i.UpdatedAt
	log.Printf("No: %+v\n", i)
	return
}

// ToNotificationWithAttempt :
func ToNotificationWithAttempt(i *model.Notification, j *model.NotificationAttempt) (o NotificationWithAttempt) {
	o.ID = i.ID.Hex()
	o.MerchantID = i.MerchantID
	o.Type = i.Type
	o.NotificationURL = i.NotificationURL
	o.NotificationKey = i.NotificationKey
	o.RequestID = i.RequestID
	o.Payload = i.Payload
	o.Status = j.Status
	o.StatusCode = j.StatusCode
	o.AttemptNo = j.AttemptNo
	o.SentAt = j.SentAt
	o.CreatedAt = i.CreatedAt
	o.UpdatedAt = i.UpdatedAt

	return
}
