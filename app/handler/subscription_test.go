package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"xenotification/app/bootstrap"
	"xenotification/app/kit/validator"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func setup() Handler {
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
	h := setup()

	var input struct {
		MerchantID      string `json:"merchantId"`
		Type            string `json:"type"`
		NotificationURL string `json:"notificationUrl"`
	}

	input.MerchantID = "123456"
	input.Type = "TEST"
	input.NotificationURL = "https://google.com"

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

func TestDeleteSubscription(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()
	h := setup()

	var input struct {
		MerchantID      string `json:"merchantId"`
		Type            string `json:"type"`
		NotificationURL string `json:"notificationUrl"`
	}

	input.MerchantID = "123456"
	input.Type = "TEST2"

	data, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodDelete, "/v1/subscription", strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, h.DeleteSubscription(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// Get the body and check
		var response struct {
			Item bool `json:"item"`
		}

		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response)) {
			assert.Equal(t, true, response.Item)

			// Check in the db if there is still record
			_, err := h.repository.FindNotification(input.Type, input.MerchantID)
			assert.Equal(t, mongo.ErrNoDocuments, err)
		}
	}
}
