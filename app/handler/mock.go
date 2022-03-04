package handler

import (
	"io/ioutil"
	"log"
	"net/http"

	"xenotification/app/response"
	"xenotification/app/response/errcode"

	"github.com/labstack/echo/v4"
)

// SendMockRequest :
func (h Handler) SendMockRequest(c echo.Context) error {
	action := c.Param("action")

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Error: %+v\n", err)
		return c.JSON(http.StatusBadRequest, response.NewException(c, errcode.InvalidRequest, err))
	}

	log.Printf("URL: %+v\n", c.Request().URL.Path)
	log.Printf("HEADER: %+v\n", c.Request().Header)
	log.Printf("REQUEST: %+v\n", string(body))

	switch action {
	case "fail":
		return c.JSON(http.StatusMethodNotAllowed, map[string]interface{}{
			"message": "Failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "This is success message",
	})
}
