package api

import (
	"github.com/hm-edu/domain-service/pkg/model"
	"github.com/hm-edu/domain-service/pkg/store"
)

type Handler struct {
	domainStore *store.DomainStore

	validator *model.Validator
}

func NewHandler(ds *store.DomainStore) *Handler {
	v := model.NewValidator()
	return &Handler{
		domainStore: ds,
		validator:   v,
	}
}
