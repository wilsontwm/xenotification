package handler

import (
	"fmt"
	"net/http"
	"time"

	httprequest "xenotification/app/kit/httpRequest"
	"xenotification/app/model"
	"xenotification/app/response"
	"xenotification/app/response/errcode"
	"xenotification/app/response/transformer"
	"xenotification/app/types"

	"github.com/go-redsync/redsync"
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
	lastAttempt, err := h.triggerNotification(notification, &resp)
	if err != nil {
		return c.JSON(http.StatusBadGateway, response.NewException(c, errcode.NotificationError, err))
	}

	return c.JSON(http.StatusOK,
		map[string]interface{}{
			"item":     transformer.ToNotification(notification, lastAttempt),
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

	// Lock based on the request ID first
	notificationRequestLock := h.redsync.NewMutex(fmt.Sprintf("%s-%s", input.Type, input.RequestID), redsync.SetExpiry(120*time.Second))
	if err := notificationRequestLock.Lock(); err != nil {
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
			Item: transformer.ToNotification(notify.notification, notify.lastAttempt),
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
	lastAttempt, err := h.triggerNotification(notification, &resp)
	if err != nil {
		return c.JSON(http.StatusBadGateway, response.NewException(c, errcode.NotificationError, err))
	}

	return c.JSON(http.StatusOK,
		map[string]interface{}{
			"item":     transformer.ToNotification(notification, lastAttempt),
			"response": resp,
		})
}

func (h Handler) triggerNotification(notification *model.Notification, resp interface{}) (*model.NotificationAttempt, error) {
	headers := map[string]string{
		"X-Xendit-Key": notification.NotificationKey,
	}
	statusCode, notificationErr := httprequest.HttpAPI(http.MethodPost, notification.NotificationURL, headers, notification.Payload, &resp)

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
	if statusCode >= 400 || statusCode < 200 {
		lastAttempt.Status = types.NotificationStatusFailed
		lastAttempt.StatusCode = statusCode
		if notificationErr != nil {
			e := notificationErr.Error()
			lastAttempt.Error = &e
		}

	} else {
		lastAttempt.Status = types.NotificationStatusSuccess
		lastAttempt.StatusCode = statusCode
		lastAttempt.SentAt = &now
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
