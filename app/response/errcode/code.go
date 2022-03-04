package errcode

import "sync"

// Error Codes
const (
	ValidationError             = "VALIDATION_ERROR"
	ServiceError                = "SERVICE_ERROR"
	NotFoundError               = "NOT_FOUND"
	InvalidRequest              = "INVALID_REQUEST"
	SystemError                 = "SYSTEM_ERROR"
	APIEndpointNotExist         = "API_ENDPOINT_NOT_EXIST"
	NotificationError           = "NOTIFICATION_ERROR"
	NotificationAttemptNotFound = "NOTIFICATION_ATTEMPT_NOT_EXIST"

	// Validation error
	OnlyFailedNotificationCanRetry = "ONLY_FAILED_NOTIFICATION_CAN_RETRY"
)

// Message :
var Message sync.Map

func init() {
	Message.Store(ValidationError, "Validation error")
	Message.Store(InvalidRequest, "Request input is invalid")
	Message.Store(SystemError, "System busy, please try again")
	Message.Store(ServiceError, "Service error")
	Message.Store(NotFoundError, "No results found")
	Message.Store(APIEndpointNotExist, "API endpoint not exist")
	Message.Store(NotificationError, "Notification error")
	Message.Store(NotificationAttemptNotFound, "Notification attempt not exist")
	Message.Store(OnlyFailedNotificationCanRetry, "Only failed notification can be retried")
}
