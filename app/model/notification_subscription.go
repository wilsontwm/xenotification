package model

// SubscriptionKey :
type SubscriptionKey struct {
	MerchantID string `bson:"merchantId" json:"merchantId"`
	Type       string `bson:"type" json:"type"`
}

// NotificationSubscription :
type NotificationSubscription struct {
	ID              SubscriptionKey `bson:"_id" json:"_id"`
	NotificationURL string          `bson:"notificationUrl" json:"notificationUrl"`
	NotificationKey string          `bson:"notificationKey" json:"notificationKey"`
	Model           `bson:",inline"`
}
