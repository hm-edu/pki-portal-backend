package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/hm-edu/domain-service/ent"
	"github.com/hm-edu/domain-service/pkg/model"
	"github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/portal-common/models"
)

func (h *Handler) CreateDomain(c *fiber.Ctx) error {
	var d ent.Domain
	req := &model.DomainCreateRequest{}
	if err := req.Bind(c, &d, h.validator); err != nil {
		return c.Status(http.StatusUnprocessableEntity).JSON(models.Error{})
	}
	d.Owner = auth.UserFromRequest(c)
	err := h.domainStore.CreateDomain(c.UserContext(), &d)
	if err != nil {
		return c.Status(http.StatusForbidden).JSON(models.Error{})
	}
	return c.Status(http.StatusCreated).JSON(model.DomainToOutput(&d))
}

func (h *Handler) DeleteDomain(c *fiber.Ctx) error {

	domains, err := h.domainStore.ListDomains(c.UserContext(), auth.UserFromRequest(c))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.Error{})
	}
	return c.Status(http.StatusOK).JSON(domains)
}

func (h *Handler) ListDomains(c *fiber.Ctx) error {

	domains, err := h.domainStore.ListDomains(c.UserContext(), auth.UserFromRequest(c))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.Error{})
	}
	return c.Status(http.StatusOK).JSON(helper.Map(domains, model.DomainToOutput))
}
