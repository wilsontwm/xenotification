package httprequest

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/go-resty/resty/v2"
	"github.com/imdario/mergo"
)

// HttpAPI :
func HttpAPI(method, requestURL string, headers map[string]string, request, response interface{}) (int, error) {
	if reflect.ValueOf(response).Kind() != reflect.Ptr {
		return 0, errors.New("response struct should be pointer")
	}

	head := map[string]string{
		"Content-Type": "application/json",
	}

	mergo.Merge(&headers, head)

	// startAt := time.Now().UTC()
	var resp *resty.Response
	client := resty.New().
		SetDebug(false).
		SetHeaders(headers).
		// SetTimeout(15 * time.Minute).
		OnAfterResponse(func(_ *resty.Client, rsp *resty.Response) error {
			// if c.logger != nil {
			// 	go c.logger.LogAPI(c.requestID, request, rsp)
			// }
			return nil
		}).R()

	var err error
	switch method {
	case "get":
		resp, err = client.Get(requestURL)
	default:
		resp, err = client.SetBody(request).Post(requestURL)
	}
	if err != nil {
		return 0, err
	}

	if response != nil {
		err := json.Unmarshal(resp.Body(), &response)
		if err != nil {
			return 0, err
		}
	}

	return resp.StatusCode(), nil
}
