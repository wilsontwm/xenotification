package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"xenotification/app/kit/helper"
	httprequest "xenotification/app/kit/httpRequest"
	"xenotification/app/model"
	"xenotification/app/response"
	"xenotification/app/response/errcode"
	"xenotification/app/response/transformer"
	"xenotification/app/types"

	"github.com/go-redsync/redsync"
	"github.com/ivpusic/grpool"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SimulateNotification :
func (h Handler) SimulateNotification(c echo.Context) error {

	var input struct {
		MerchantID      string `json:"merchantId" validate:"required"`
		NotificationURL string `json:"notificationURL" validate:"required"`
		NotificationKey string `json:"notificationKey"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewException(c, errcode.InvalidRequest, err))
	}

	if err := c.Validate(&input); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.NewException(c, errcode.ValidationError, err))
	}

	notification := new(model.Notification)
	notification.ID = primitive.NewObjectID()
	notification.MerchantID = input.MerchantID
	notification.NotificationKey = input.NotificationKey
	notification.Payload = "This is test from Xendit"
	notification.NotificationURL = input.NotificationURL
	notification.IsSimulation = true
	notification.CreatedAt = time.Now().UTC()
	notification.UpdatedAt = time.Now().UTC()

	var resp interface{}
	lastAttempt, err := h.triggerNotification(notification, []int{}, &resp)
	if err != nil {
		return c.JSON(http.StatusBadGateway, response.NewException(c, errcode.NotificationError, err))
	}

	return c.JSON(http.StatusOK,
		map[string]interface{}{
			"item":     transformer.ToNotificationWithAttempt(notification, lastAttempt),
			"response": resp,
		})
}

// SendNotification :
func (h Handler) SendNotification(c echo.Context) error {

	var input struct {
		MerchantID string      `json:"merchantId" validate:"required"`
		RequestID  string      `json:"requestId" validate:"required"`
		Type       string      `json:"type" validate:"required"`
		Payload    interface{} `json:"payload" validate:"required"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewException(c, errcode.InvalidRequest, err))
	}

	if err := c.Validate(&input); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.NewException(c, errcode.ValidationError, err))
	}

	type notificationWithAttempt struct {
		notification *model.Notification
		lastAttempt  *model.NotificationAttempt
	}

	getNotificationWithAttempt := func(notification *model.Notification) (*notificationWithAttempt, error) {
		lastAttempt, err := h.repository.FindLastNotificationAttempt(notification.ID)
		if err != nil {
			return nil, err
		}

		return &notificationWithAttempt{
			notification: notification,
			lastAttempt:  lastAttempt,
		}, nil
	}

	// Lock based on the request ID first
	notificationRequestLock := h.redsync.NewMutex(fmt.Sprintf("%s-%s", input.Type, input.RequestID), redsync.SetExpiry(120*time.Second))
	if err := notificationRequestLock.Lock(); err != nil {
		// Try to get the notification if there is
		notification, err := h.repository.FindNotification(input.Type, input.RequestID)
		if notification != nil {
			notify, err := getNotificationWithAttempt(notification)
			if err != nil {
				return c.JSON(http.StatusNotFound, response.NewException(c, errcode.NotificationAttemptNotFound, err))
			}

			return c.JSON(http.StatusOK, response.Item{
				Item: transformer.ToNotificationWithAttempt(notify.notification, notify.lastAttempt),
			})
		}
		return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
	}
	defer notificationRequestLock.Unlock()

	// Check if the merchant has subscribe to the notification
	subscription, err := h.repository.FindNotificationSubscription(model.SubscriptionKey{MerchantID: input.MerchantID, Type: input.Type})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusOK, response.Item{Item: nil})
		}
		return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
	}

	// Check if there is notification for the request id
	notification, err := h.repository.FindNotification(input.Type, input.RequestID)
	if err != nil && err != mongo.ErrNoDocuments {
		return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
	} else if notification != nil {
		notify, err := getNotificationWithAttempt(notification)
		if err != nil {
			return c.JSON(http.StatusNotFound, response.NewException(c, errcode.NotificationAttemptNotFound, err))
		}

		return c.JSON(http.StatusOK, response.Item{
			Item: transformer.ToNotificationWithAttempt(notify.notification, notify.lastAttempt),
		})
	}

	// Create notification
	notification = new(model.Notification)
	notification.ID = primitive.NewObjectID()
	notification.MerchantID = input.MerchantID
	notification.RequestID = input.RequestID
	notification.Type = input.Type
	notification.Payload = input.Payload
	notification.NotificationKey = subscription.NotificationKey
	notification.NotificationURL = subscription.NotificationURL
	notification.Status = types.NotificationStatusPending
	notification.CreatedAt = time.Now().UTC()
	notification.UpdatedAt = time.Now().UTC()

	if err := h.repository.CreateNotification(notification); err != nil {
		return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
	}

	var resp interface{}
	lastAttempt, err := h.triggerNotification(notification, subscription.AcceptableStatusCodes, &resp)
	if err != nil {
		return c.JSON(http.StatusBadGateway, response.NewException(c, errcode.NotificationError, err))
	}

	return c.JSON(http.StatusOK,
		map[string]interface{}{
			"item":     transformer.ToNotificationWithAttempt(notification, lastAttempt),
			"response": resp,
		})
}

// ResendNotification :
func (h Handler) ResendNotification(c echo.Context) error {

	var input struct {
		MerchantID      string `json:"merchantId" validate:"required"`
		Type            string `json:"type" validate:"required"`
		RequestID       string `json:"requestId" validate:"required"`
		NotificationURL string `json:"notificationUrl" validate:"omitempty"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewException(c, errcode.InvalidRequest, err))
	}

	if err := c.Validate(&input); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.NewException(c, errcode.ValidationError, err))
	}

	// Lock based on the request ID first
	notificationRequestLock := h.redsync.NewMutex(fmt.Sprintf("%s-%s", input.Type, input.RequestID), redsync.SetExpiry(30*time.Second))
	if err := notificationRequestLock.Lock(); err != nil {
		return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
	}
	defer notificationRequestLock.Unlock()

	notification, err := h.repository.FindNotification(input.Type, input.RequestID)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.NewException(c, errcode.NotFoundError, err))
	} else if notification.MerchantID != input.MerchantID {
		return c.JSON(http.StatusNotFound, response.NewException(c, errcode.NotFoundError, errors.New("Notification not found")))
	} else if notification.Status != types.NotificationStatusFailed {
		return c.JSON(http.StatusBadRequest, response.NewException(c, errcode.OnlyFailedNotificationCanRetry, errors.New("Only failed notification can be retried")))
	}

	// Check if it's a valid url
	if input.NotificationURL != "" {
		if _, err := url.ParseRequestURI(input.NotificationURL); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, response.NewException(c, errcode.ValidationError, err))
		}
		notification.NotificationURL = input.NotificationURL
	}

	notificationAttempt := new(model.NotificationAttempt)
	notificationAttempt.ID = primitive.NewObjectID()
	notificationAttempt.NotificationID = notification.ID
	notificationAttempt.MerchantID = notification.MerchantID
	notificationAttempt.AttemptNo = notification.AttemptNo + 1
	notificationAttempt.Status = types.NotificationStatusPending
	notificationAttempt.CreatedAt = time.Now().UTC()
	notificationAttempt.UpdatedAt = time.Now().UTC()

	if err := h.repository.UpsertNotificationAttempt(notificationAttempt); err != nil {
		return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
	}

	acceptableCodes := []int{}
	if subscription, err := h.repository.FindNotificationSubscription(model.SubscriptionKey{MerchantID: notification.MerchantID, Type: notification.Type}); err == nil {
		acceptableCodes = subscription.AcceptableStatusCodes
	}

	var resp interface{}
	lastAttempt, err := h.triggerNotification(notification, acceptableCodes, &resp)
	if err != nil {
		return c.JSON(http.StatusBadGateway, response.NewException(c, errcode.NotificationError, err))
	}

	return c.JSON(http.StatusOK,
		map[string]interface{}{
			"item":     transformer.ToNotificationWithAttempt(notification, lastAttempt),
			"response": resp,
		})
}

func (h Handler) triggerNotification(notification *model.Notification, acceptableStatusCodes []int, resp interface{}) (*model.NotificationAttempt, error) {
	headers := map[string]string{
		"X-Xendit-Key": notification.NotificationKey,
	}
	statusCode, notificationErr := httprequest.HttpAPI(http.MethodPost, notification.NotificationURL, headers, notification.Payload, &resp)

	// time.Sleep(60 * time.Second)

	var lastAttempt *model.NotificationAttempt
	if !notification.IsSimulation {
		var err error
		lastAttempt, err = h.repository.FindLastNotificationAttempt(notification.ID)
		if err != nil {
			return nil, err
		}
	} else {
		lastAttempt = new(model.NotificationAttempt)
		lastAttempt.NotificationID = notification.ID
		lastAttempt.MerchantID = notification.MerchantID
	}

	now := time.Now().UTC()
	isSuccess := false
	if len(acceptableStatusCodes) > 0 {
		isSuccess = helper.Contains(acceptableStatusCodes, statusCode)
	} else {
		isSuccess = statusCode < 400 && statusCode >= 200
	}

	if isSuccess {
		lastAttempt.Status = types.NotificationStatusSuccess
		lastAttempt.StatusCode = statusCode
		lastAttempt.SentAt = &now
	} else {
		lastAttempt.Status = types.NotificationStatusFailed
		lastAttempt.StatusCode = statusCode
		if notificationErr != nil {
			e := notificationErr.Error()
			lastAttempt.Error = &e
		}
	}

	lastAttempt.UpdatedAt = time.Now().UTC()
	notification.AttemptNo = lastAttempt.AttemptNo
	notification.AttemptedAt = &now
	notification.Status = lastAttempt.Status
	notification.UpdatedAt = time.Now().UTC()

	if !notification.IsSimulation {
		go h.repository.UpsertNotification(notification)
		go h.repository.UpsertNotificationAttempt(lastAttempt)
	}

	return lastAttempt, nil
}

// GetNotifications :
func (h Handler) GetNotifications(c echo.Context) error {
	var input struct {
		MerchantID string `query:"merchantId"`
		Cursor     string `query:"cursor"`
		Limit      int64  `query:"limit"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewException(c, errcode.InvalidRequest, err))
	}

	notifications, cursor, err := h.repository.FindNotifications(input.MerchantID, input.Cursor, input.Limit)
	if err != nil && err != mongo.ErrNoDocuments {
		return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
	}

	pool := grpool.NewPool(20, 20)
	defer pool.Release()

	pool.WaitCount(len(notifications))
	formattedNotifications := make([]transformer.Notification, len(notifications))
	for l, each := range notifications {
		pool.JobQueue <- func(i int, not *model.Notification) func() {
			return func() {
				defer pool.JobDone()
				formattedNotifications[i] = transformer.ToNotification(not)
			}
		}(l, each)
	}
	pool.WaitAll()

	return c.JSON(http.StatusOK, response.Items{
		Items:  formattedNotifications,
		Count:  len(formattedNotifications),
		Cursor: cursor,
	})
}
