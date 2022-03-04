package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"xenotification/app/env"
	"xenotification/app/response"
	"xenotification/app/response/errcode"

	redis "github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/ulule/limiter/v3"
	r "github.com/ulule/limiter/v3/drivers/store/redis"
)

// Rate :
type Rate struct {
	limiter.Rate
}

// RateLimitLogin :
var (
	RateLimitMaxPerRequest = newRate(1, 10)
)

func newRate(periodInSecond, limit int64) Rate {
	return Rate{
		Rate: limiter.Rate{
			Period: time.Duration(periodInSecond) * time.Second,
			Limit:  limit,
		},
	}
}

var instance *limiter.Limiter

func init() {
	store, err := r.NewStore(redis.NewClient(&redis.Options{
		Addr:     env.Config.Redis.Host,
		Password: env.Config.Redis.Password,
		DB:       1,
	}))
	if err != nil {
		panic(err)
	}
	instance = limiter.New(store, limiter.Rate{})
}

// APIRateLimit :
func (mw *Middleware) APIRateLimit(i int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := Rate{
				Rate: limiter.Rate{
					Period: time.Duration(1) * time.Minute,
					Limit:  i,
				},
			}

			realIP := c.RealIP()

			formatter := fmt.Sprintf("%x%x%x", c.Request().URL.Path, c.Request().Method, realIP)
			context, err := instance.Store.Get(c.Request().Context(), formatter, r.Rate)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, response.NewException(c, errcode.SystemError, err))
			}

			if context.Reached {
				return c.JSON(http.StatusTooManyRequests, response.NewException(c, errcode.TooManyRequests, errors.New("Too many requests, please try again later")))

			}
			return next(c)
		}
	}
}
