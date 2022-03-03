package handler

import (
	"fmt"
	"net/http"

	"xenotification/app/bootstrap"
	"xenotification/app/env"
	"xenotification/app/repository"

	"github.com/go-redsync/redsync"
	"github.com/labstack/echo/v4"
)

// Handler :
type Handler struct {
	repository *repository.Repository
	redsync    *redsync.Redsync
}

// New :
func New(bs *bootstrap.Bootstrap) *Handler {
	return &Handler{
		repository: bs.Repository,
		redsync:    bs.Redsync,
	}
}

// APIHealthCheck :
func (h Handler) APIHealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("Your server version %s is running", env.Config.App.Version),
	})
}
