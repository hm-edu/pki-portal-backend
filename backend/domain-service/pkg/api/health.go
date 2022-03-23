package api

import (
	"net/http"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
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
func (s *APIServer) healthzHandler(c *fiber.Ctx) (err error) {
	if atomic.LoadInt32(&healthy) == 1 {
		return c.JSON(healthResponse{Status: "OK"})
	}
	return c.Status(http.StatusServiceUnavailable).JSON(healthResponse{Status: "Service Unavailable"})
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
func (s *APIServer) readyzHandler(c *fiber.Ctx) (err error) {
	if atomic.LoadInt32(&ready) == 1 {
		return c.JSON(healthResponse{Status: "OK"})
	}
	return c.Status(http.StatusServiceUnavailable).JSON(healthResponse{Status: "Service Unavailable"})
}
