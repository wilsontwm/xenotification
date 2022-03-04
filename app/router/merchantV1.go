package router

import (
	"github.com/labstack/echo/v4"
)

// merchantV1 :
func merchantV1(e *echo.Echo, r *Router) {
	h := r.handler
	mw := r.apiMiddleware
	v1 := e.Group("/v1", mw.OpenTracing("notification"))

	cronRoute := v1.Group("/cron")
	cronRoute.POST("/resend-notification", h.CronSendNotification)

	subscriptionRoute := v1.Group("/subscription")
	subscriptionRoute.GET("s", h.GetSubscriptions)
	subscriptionRoute.PUT("", h.UpsertSubscription)
	subscriptionRoute.DELETE("", h.DeleteSubscription)

	notificationRoute := v1.Group("/notify")
	notificationRoute.GET("s", h.GetNotifications)
	notificationRoute.POST("", h.SendNotification)
	notificationRoute.POST("/resend", h.ResendNotification)
	notificationRoute.POST("/simulate", h.SimulateNotification)

	mockRoute := v1.Group("/mock")
	mockRoute.POST("/:action", h.SendMockRequest)
}
