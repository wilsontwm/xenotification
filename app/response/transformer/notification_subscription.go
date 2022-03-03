package transformer

import (
	"time"

	"xenotification/app/model"
)

// NotificationSubscription :
type NotificationSubscription struct {
	MerchantID      string    `json:"merchantId"`
	Type            string    `json:"type"`
	NotificationURL string    `json:"notificationUrl"`
	NotificationKey string    `json:"notificationKey"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ToNotificationSubscription :
func ToNotificationSubscription(i *model.NotificationSubscription) (o NotificationSubscription) {
	o.MerchantID = i.ID.MerchantID
	o.Type = i.ID.Type
	o.NotificationURL = i.NotificationURL
	o.NotificationKey = i.NotificationKey
	o.CreatedAt = i.CreatedAt
	o.UpdatedAt = i.UpdatedAt

	return
}
