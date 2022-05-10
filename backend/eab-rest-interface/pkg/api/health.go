package api

import (
	"net/http"
	"sync/atomic"

	"github.com/hm-edu/eab-rest-interface/pkg/database"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type healthResponse struct {
	Status string `json:"status"`
}

// healthzHandler godoc
func (s *Server) healthzHandler(c echo.Context) error {
	err := database.DB.Internal.Ping()
	if err != nil {
		s.logger.Error("Error connecting to database", zap.Error(err))
		return c.JSON(http.StatusServiceUnavailable, healthResponse{Status: "Service Unavailable"})
	}
	err = database.DB.NoSQLInternal.Ping()
	if err != nil {
		s.logger.Error("Error connecting to acme database", zap.Error(err))
		return c.JSON(http.StatusServiceUnavailable, healthResponse{Status: "Service Unavailable"})
	}
	return c.JSON(http.StatusOK, healthResponse{Status: "OK"})
}

// readyzHandler godoc
func (s *Server) readyzHandler(c echo.Context) (err error) {
	if atomic.LoadInt32(&ready) == 1 {
		return c.JSON(http.StatusOK, healthResponse{Status: "OK"})
	}
	return c.JSON(http.StatusServiceUnavailable, healthResponse{Status: "Service Unavailable"})
}
