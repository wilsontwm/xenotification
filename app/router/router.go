package router

import (
	"xenotification/app/bootstrap"
	"xenotification/app/handler"
	midware "xenotification/app/middleware"

	"github.com/labstack/echo/v4"
)

// Router :
type Router struct {
	apiMiddleware *midware.Middleware
	handler       *handler.Handler
}

// New :
func New(e *echo.Echo, bs *bootstrap.Bootstrap) {
	router := Router{
		apiMiddleware: midware.New(bs),
		handler:       handler.New(bs),
	}

	e.GET("/health", router.handler.APIHealthCheck)
	merchantV1(e, &router)
}
