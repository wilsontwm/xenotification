package types

type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "PENDING"
	NotificationStatusSuccess NotificationStatus = "SUCCESS"
	NotificationStatusFailed  NotificationStatus = "FAILED"
)
