package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"xenotification/app/kit/validator"
	"xenotification/app/response/transformer"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSimulateSuccessfulNotification(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()
	h := setupTest()

	var input struct {
		MerchantID      string `json:"merchantId"`
		NotificationURL string `json:"notificationURL"`
		NotificationKey string `json:"notificationKey"`
	}

	input.MerchantID = "123456"
	input.NotificationURL = "https://xendit.free.beeceptor.com/notify"

	data, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/v1/notify/simulate", strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, h.SimulateNotification(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// Get the body and check
		var response struct {
			Item struct {
				StatusCode int `json:"statusCode"`
			} `json:"item"`
		}

		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response)) {
			assert.Equal(t, http.StatusOK, response.Item.StatusCode)
		}
	}
}

func TestSimulateFailedNotification(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()
	h := setupTest()

	var input struct {
		MerchantID      string `json:"merchantId"`
		NotificationURL string `json:"notificationURL"`
		NotificationKey string `json:"notificationKey"`
	}

	input.MerchantID = "123456"
	input.NotificationURL = "https://xendit.free.beeceptor.com/fail"

	data, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/v1/notify/simulate", strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, h.SimulateNotification(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// Get the body and check
		var response struct {
			Item struct {
				StatusCode int `json:"statusCode"`
			} `json:"item"`
		}

		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response)) {
			assert.Equal(t, http.StatusMethodNotAllowed, response.Item.StatusCode)
		}
	}
}

func TestSendSubscibedNotification(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()
	h := setupTest()

	var input struct {
		MerchantID string      `json:"merchantId"`
		RequestID  string      `json:"requestID"`
		Type       string      `json:"type"`
		Payload    interface{} `json:"payload"`
	}

	input.MerchantID = "123456"
	input.RequestID = fmt.Sprintf("%d", time.Now().Unix())
	input.Type = "TEST"

	type d struct {
		Amount      uint64 `json:"amount"`
		Description string `json:"description"`
	}

	input.Payload = d{
		Amount:      24526,
		Description: "This is triggered from unit test",
	}

	data, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/v1/notify", strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, h.SendNotification(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// Get the body and check
		var response struct {
			Item struct {
				StatusCode int `json:"statusCode"`
			} `json:"item"`
		}

		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response)) {
			assert.Equal(t, http.StatusOK, response.Item.StatusCode)
		}
	}
}

func TestSendUnsubscibedNotification(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()
	h := setupTest()

	var input struct {
		MerchantID string      `json:"merchantId"`
		RequestID  string      `json:"requestID"`
		Type       string      `json:"type"`
		Payload    interface{} `json:"payload"`
	}

	input.MerchantID = "123456"
	input.RequestID = fmt.Sprintf("%d", time.Now().Unix())
	input.Type = fmt.Sprintf("%d", time.Now().Unix())

	type d struct {
		Amount      uint64 `json:"amount"`
		Description string `json:"description"`
	}

	input.Payload = d{
		Amount:      3526262,
		Description: "This is triggered from unit test",
	}

	data, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/v1/notify", strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, h.SendNotification(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// Get the body and check
		var response struct {
			Item interface{} `json:"item"`
		}

		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response)) {
			assert.Equal(t, nil, response.Item)
		}
	}
}

func TestGetNotifications(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()
	h := setupTest()

	var input struct {
		MerchantID string `query:"merchantId"`
	}

	input.MerchantID = "123456"

	data, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodGet, "/v1/notifys", strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, h.GetNotifications(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		// Get the body and check
		var response struct {
			Items []transformer.Notification `json:"items"`
		}

		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response)) {
			assert.Less(t, 0, len(response.Items))
		}
	}
}

func TestResendNotification(t *testing.T) {
	e := echo.New()
	e.Validator = validator.New()
	h := setupTest()

	// Subscribe to a fail type notification first
	var subInput struct {
		MerchantID      string `json:"merchantId"`
		Type            string `json:"type"`
		NotificationURL string `json:"notificationUrl"`
	}

	subInput.MerchantID = "123456"
	subInput.Type = "FAIL"
	subInput.NotificationURL = "https://xendit.free.beeceptor.com/fail"

	data, _ := json.Marshal(subInput)

	req := httptest.NewRequest(http.MethodPut, "/v1/subscription", strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Assertions
	if assert.NoError(t, h.UpsertSubscription(c)) {
		// Send notification for the first time
		var nottInput struct {
			MerchantID string      `json:"merchantId"`
			RequestID  string      `json:"requestID"`
			Type       string      `json:"type"`
			Payload    interface{} `json:"payload"`
		}

		nottInput.MerchantID = subInput.MerchantID
		nottInput.RequestID = fmt.Sprintf("%d", time.Now().Unix())
		nottInput.Type = subInput.Type

		type d struct {
			Amount      uint64 `json:"amount"`
			Description string `json:"description"`
		}

		nottInput.Payload = d{
			Amount:      2426262526,
			Description: "This is triggered from unit test (resend)",
		}

		data, _ := json.Marshal(nottInput)

		req := httptest.NewRequest(http.MethodPost, "/v1/notify", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Assertions
		if assert.NoError(t, h.SendNotification(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			// Get the body and check
			var response struct {
				Item struct {
					StatusCode int `json:"statusCode"`
				} `json:"item"`
			}

			if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response)) && assert.Equal(t, http.StatusMethodNotAllowed, response.Item.StatusCode) {
				// Resend notification
				var resendInput struct {
					MerchantID string `json:"merchantId"`
					RequestID  string `json:"requestID"`
					Type       string `json:"type"`
				}

				resendInput.MerchantID = subInput.MerchantID
				resendInput.RequestID = nottInput.RequestID
				resendInput.Type = subInput.Type

				data, _ := json.Marshal(resendInput)

				req := httptest.NewRequest(http.MethodPost, "/v1/notify/resend", strings.NewReader(string(data)))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)
				if assert.NoError(t, h.ResendNotification(c)) {
					assert.Equal(t, http.StatusOK, rec.Code)
					// Get the body and check
					var response struct {
						Item struct {
							StatusCode int `json:"statusCode"`
							AttemptNo  int `json:"attemptNo"`
						} `json:"item"`
					}

					if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response)) {
						assert.Equal(t, http.StatusMethodNotAllowed, response.Item.StatusCode)
						assert.Equal(t, 2, response.Item.AttemptNo)
					}
				}

			}
		}

	}
}
