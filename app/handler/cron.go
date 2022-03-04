package handler

import (
	"fmt"
	"net/http"
	"time"

	"xenotification/app/model"
	"xenotification/app/response"
	"xenotification/app/response/errcode"
	"xenotification/app/types"

	"github.com/go-redsync/redsync"
	"github.com/ivpusic/grpool"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CronSendNotification :
func (h Handler) CronSendNotification(c echo.Context) error {
	// Get all the failed notification attempts to be retried
	var failedNotificationsToRetry []*model.Notification

	cursor := ""
	for {
		notifications, newCursor, err := h.repository.FindRetryNotifications(cursor)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
		}

		failedNotificationsToRetry = append(failedNotificationsToRetry, notifications...)

		if newCursor != "" {
			cursor = newCursor
		} else {
			break
		}
	}

	pool := grpool.NewPool(20, 20)
	defer pool.Release()

	pool.WaitCount(len(failedNotificationsToRetry))

	for l, each := range failedNotificationsToRetry {
		pool.JobQueue <- func(i int, notification *model.Notification) func() {
			return func() {
				defer pool.JobDone()
				// Lock based on the request ID first
				notificationRequestLock := h.redsync.NewMutex(fmt.Sprintf("%s-%s", notification.Type, notification.RequestID), redsync.SetExpiry(120*time.Second))
				if err := notificationRequestLock.Lock(); err != nil {
					return
				}
				defer notificationRequestLock.Unlock()

				notificationAttempt := new(model.NotificationAttempt)
				notificationAttempt.ID = primitive.NewObjectID()
				notificationAttempt.NotificationID = notification.ID
				notificationAttempt.MerchantID = notification.MerchantID
				notificationAttempt.AttemptNo = notification.AttemptNo + 1
				notificationAttempt.Status = types.NotificationStatusPending
				notificationAttempt.CreatedAt = time.Now().UTC()
				notificationAttempt.UpdatedAt = time.Now().UTC()

				if err := h.repository.UpsertNotificationAttempt(notificationAttempt); err != nil {
					return
				}

				acceptableCodes := []int{}
				if subscription, err := h.repository.FindNotificationSubscription(model.SubscriptionKey{MerchantID: notification.MerchantID, Type: notification.Type}); err == nil {
					acceptableCodes = subscription.AcceptableStatusCodes
				}

				var resp interface{}
				h.triggerNotification(notification, acceptableCodes, &resp)
			}
		}(l, each)
	}
	pool.WaitAll()

	return nil

}
