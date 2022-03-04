package handler

import (
	"net/http"
	"net/url"
	"time"

	"xenotification/app/kit/helper"
	"xenotification/app/model"
	"xenotification/app/response"
	"xenotification/app/response/errcode"
	"xenotification/app/response/transformer"

	"github.com/ivpusic/grpool"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetSubscriptions :
func (h Handler) GetSubscriptions(c echo.Context) error {
	var input struct {
		MerchantID string `query:"merchantId"`
		Cursor     string `query:"cursor"`
		Limit      int64  `query:"limit"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewException(c, errcode.InvalidRequest, err))
	}

	subscriptions, cursor, err := h.repository.FindNotificationSubscriptions(input.MerchantID, input.Cursor, input.Limit)
	if err != nil && err != mongo.ErrNoDocuments {
		return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
	}

	pool := grpool.NewPool(20, 20)
	defer pool.Release()

	pool.WaitCount(len(subscriptions))
	formattedSubscriptions := make([]transformer.NotificationSubscription, len(subscriptions))
	for l, each := range subscriptions {
		pool.JobQueue <- func(i int, sub *model.NotificationSubscription) func() {
			return func() {
				defer pool.JobDone()
				formattedSubscriptions[i] = transformer.ToNotificationSubscription(sub)
			}
		}(l, each)
	}
	pool.WaitAll()

	return c.JSON(http.StatusOK, response.Items{
		Items:  formattedSubscriptions,
		Count:  len(formattedSubscriptions),
		Cursor: cursor,
	})
}

// UpsertSubscription :
func (h Handler) UpsertSubscription(c echo.Context) error {

	var input struct {
		MerchantID      string `json:"merchantId" validate:"required"`
		Type            string `json:"type" validate:"required"`
		NotificationURL string `json:"notificationUrl" validate:"required"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewException(c, errcode.InvalidRequest, err))
	}

	if err := c.Validate(&input); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.NewException(c, errcode.ValidationError, err))
	}

	// Check if it's a valid url
	if _, err := url.ParseRequestURI(input.NotificationURL); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.NewException(c, errcode.ValidationError, err))
	}

	subscription := new(model.NotificationSubscription)
	subscription.ID.MerchantID = input.MerchantID
	subscription.ID.Type = input.Type
	subscription.NotificationURL = input.NotificationURL
	subscription.NotificationKey = helper.RandomString(24)
	subscription.CreatedAt = time.Now().UTC()
	subscription.UpdatedAt = time.Now().UTC()

	if err := h.repository.UpsertNotificationSubscription(subscription); err != nil {
		return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
	}

	return c.JSON(http.StatusOK, response.Item{
		Item: transformer.ToNotificationSubscription(subscription),
	})
}

// DeleteSubscription :
func (h Handler) DeleteSubscription(c echo.Context) error {

	var input struct {
		MerchantID string `json:"merchantId" validate:"required"`
		Type       string `json:"type" validate:"required"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, response.NewException(c, errcode.InvalidRequest, err))
	}

	if err := c.Validate(&input); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.NewException(c, errcode.ValidationError, err))
	}

	if err := h.repository.DeleteNotificationSubscription(model.SubscriptionKey{MerchantID: input.MerchantID, Type: input.Type}); err != nil {
		return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
	}

	return c.JSON(http.StatusOK, response.Item{
		Item: true,
	})
}
