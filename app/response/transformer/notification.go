package transformer

import (
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
	StatusCode      int                      `json:"statusCode"`
	SentAt          *time.Time               `json:"sentAt,omitempty"`
	CreatedAt       time.Time                `json:"createdAt"`
	UpdatedAt       time.Time                `json:"updatedAt"`
}

// ToNotification :
func ToNotification(i *model.Notification, j *model.NotificationAttempt) (o Notification) {
	o.ID = i.ID.Hex()
	o.MerchantID = i.MerchantID
	o.Type = i.Type
	o.NotificationURL = i.NotificationURL
	o.NotificationKey = i.NotificationKey
	o.RequestID = i.RequestID
	o.Payload = i.Payload
	o.Status = j.Status
	o.StatusCode = j.StatusCode
	o.SentAt = j.SentAt
	o.CreatedAt = i.CreatedAt
	o.UpdatedAt = i.UpdatedAt

	return
}
