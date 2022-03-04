package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"xenotification/app/bootstrap"
	"xenotification/app/kit/validator"
	"xenotification/app/response/transformer"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func setupTest() Handler {
	bs := bootstrap.New()
	h := Handler{
		repository: bs.Repository,
		redsync:    bs.Redsync,
	}
	return h
}

func TestUpsertSubscription(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()
	h := setupTest()

	var input struct {
		MerchantID      string `json:"merchantId"`
		Type            string `json:"type"`
		NotificationURL string `json:"notificationUrl"`
	}

	input.MerchantID = "123456"
	input.Type = "TEST"
	input.NotificationURL = "https://xendit.free.beeceptor.com/notify"

	data, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPut, "/v1/subscription", strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, h.UpsertSubscription(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// Get the body and check
		var response struct {
			Item struct {
				MerchantID      string `json:"merchantId"`
				Type            string `json:"type"`
				NotificationURL string `json:"notificationUrl"`
			} `json:"item"`
		}

		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response)) {
			assert.Equal(t, input.MerchantID, response.Item.MerchantID)
			assert.Equal(t, input.Type, response.Item.Type)
			assert.Equal(t, input.NotificationURL, response.Item.NotificationURL)
		}
	}
}

func TestGetSubscriptions(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()
	h := setupTest()

	var input struct {
		MerchantID string `query:"merchantId"`
	}

	input.MerchantID = "123456"

	data, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodGet, "/v1/subscriptions", strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, h.GetSubscriptions(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// Get the body and check
		var response struct {
			Items []transformer.NotificationSubscription `json:"items"`
		}

		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response)) {
			assert.Less(t, 0, len(response.Items))
		}
	}
}

func TestDeleteSubscription(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()
	h := setupTest()

	var input struct {
		MerchantID      string `json:"merchantId"`
		Type            string `json:"type"`
		NotificationURL string `json:"notificationUrl"`
	}

	input.MerchantID = "123456"
	input.Type = "TEST2"
	input.NotificationURL = "https://google.com"

	data, _ := json.Marshal(input)

	req1 := httptest.NewRequest(http.MethodPut, "/v1/subscription", strings.NewReader(string(data)))
	req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec1 := httptest.NewRecorder()
	c := e.NewContext(req1, rec1)

	if assert.NoError(t, h.UpsertSubscription(c)) {
		assert.Equal(t, http.StatusOK, rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodDelete, "/v1/subscription", strings.NewReader(string(data)))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c = e.NewContext(req2, rec2)

	// Assertions
	if assert.NoError(t, h.DeleteSubscription(c)) {
		assert.Equal(t, http.StatusOK, rec2.Code)

		// Get the body and check
		var response struct {
			Item bool `json:"item"`
		}

		if assert.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &response)) {
			assert.Equal(t, true, response.Item)

			// Check in the db if there is still record
			_, err := h.repository.FindNotification(input.Type, input.MerchantID)
			assert.Equal(t, mongo.ErrNoDocuments, err)
		}
	}
}
