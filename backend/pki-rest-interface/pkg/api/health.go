package api

import (
	"net/http"
	"sync/atomic"

	"github.com/labstack/echo/v4"
)

type healthResponse struct {
	Status string `json:"status"`
}

// healthzHandler godoc
// @Summary Liveness check
// @Description Used by Kubernetes Liveness Probe
// @Tags Kubernetes
// @Accept json
// @Produce json
// @Router /healthz [get]
// @Success 200 {object} healthResponse "OK"
// @Failure 503 {object} healthResponse "Service Unavailable"
func (s *Server) healthzHandler(c echo.Context) (err error) {
	if atomic.LoadInt32(&healthy) == 1 {
		return c.JSON(http.StatusOK, healthResponse{Status: "OK"})
	}
	return c.JSON(http.StatusServiceUnavailable, healthResponse{Status: "Service Unavailable"})
}

// readyzHandler godoc
// @Summary Readiness check
// @Description Used by Kubernetes Readiness Probe
// @Tags Kubernetes
// @Accept json
// @Produce json
// @Router /readyz [get]
// @Success 200 {object} healthResponse "OK"
// @Failure 503 {object} healthResponse "Service Unavailable"
func (s *Server) readyzHandler(c echo.Context) (err error) {
	if atomic.LoadInt32(&ready) == 1 {
		return c.JSON(http.StatusOK, healthResponse{Status: "OK"})
	}
	return c.JSON(http.StatusServiceUnavailable, healthResponse{Status: "Service Unavailable"})
}
