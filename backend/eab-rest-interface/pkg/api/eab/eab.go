package eab

import (
	"net/http"

	"github.com/hm-edu/eab-rest-interface/ent"
	"github.com/hm-edu/eab-rest-interface/ent/eabkey"
	"github.com/hm-edu/eab-rest-interface/pkg/api/models"
	"github.com/hm-edu/eab-rest-interface/pkg/database"
	"github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/portal-common/logging"
	"github.com/smallstep/certificates/acme"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// GetExternalAccountKeys godoc
// @Summary Gets existing external account keys.
// @Description Gets all existing external account keys.
// @Tags EAB
// @Accept json
// @Produce json
// @Router /eab/ [get]
// @Security API
// @Success 200 {object} []models.EAB
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) GetExternalAccountKeys(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "list")
	defer span.End()
	user, err := auth.UserFromRequest(c)
	if err != nil {
		logger.Error("Error getting user from request", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Error getting user from request")
	}
	logger.Info("Requesting external account keys")
	keys, _, err := database.DB.NoSQL.GetExternalAccountKeys(ctx, h.provisionerID, "", -1)
	if err != nil {
		logger.Error("Failed to get external account keys", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get external account keys")
	}

	mappedKeys, err := database.DB.Db.EABKey.Query().Where(eabkey.User(user)).All(ctx)
	if err != nil {
		logger.Error("Failed to get external account keys", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get external account keys")
	}
	keys = helper.Where(keys, func(key *acme.ExternalAccountKey) bool {
		return helper.Any(mappedKeys, func(x *ent.EABKey) bool { return key.ID == x.EabKey })
	})
	return c.JSON(http.StatusOK, helper.Map(keys, models.NewEAB))
}

// CreateNewKey godoc
// @Summary Create a new key.
// @Description Creates a new key.
// @Tags EAB
// @Accept json
// @Produce json
// @Router /eab/ [post]
// @Security API
// @Success 201 {object} models.EAB
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) CreateNewKey(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "add")
	defer span.End()
	user, err := auth.UserFromRequest(c)
	if err != nil {
		logger.Error("Error getting user from request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Request"}
	}
	logger.Info("Requesting new external account key")
	key, err := database.DB.NoSQL.CreateExternalAccountKey(ctx, h.provisionerID, "")

	if err != nil {
		logger.Error("Failed to create new external account key", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get external account keys")
	}

	_, err = database.DB.Db.EABKey.Create().SetEabKey(key.ID).SetUser(user).Save(ctx)
	if err != nil {
		logger.Error("Failed to create new external account key", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get external account keys")
	}

	return c.JSON(http.StatusCreated, models.NewEAB(key))
}

// DeleteKey godoc
// @Summary Deletes an EAB Key.
// @Description Delete an existing EAB Key.
// @Tags EAB
// @Accept json
// @Produce json
// @Router /eab/{id} [DELETE]
// @Param id path string true "EAB ID"
// @Security API
// @Success 204
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) DeleteKey(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "delete")
	defer span.End()
	user, err := auth.UserFromRequest(c)
	if err != nil {
		logger.Error("Failed to get user from request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Request"}
	}
	logger.Info("Requesting deletion of external account key")
	key := c.Param("id")
	mapping, err := database.DB.Db.EABKey.Query().Where(eabkey.And(eabkey.User(user), eabkey.EabKey(key))).First(ctx)
	if err != nil {
		logger.Error("Failed to get external account keys", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get external account keys")
	}

	err = database.DB.NoSQL.DeleteExternalAccountKey(ctx, h.provisionerID, key)
	if err != nil {
		logger.Error("Failed to delete external account key", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete external account key")
	}

	err = database.DB.Db.EABKey.DeleteOne(mapping).Exec(ctx)
	if err != nil {
		logger.Error("Failed to delete external account key", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete external account key")
	}

	return c.NoContent(http.StatusNoContent)
}
