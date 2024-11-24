package smime

import (
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/model"
)

// Handler is a wrapper around the domainstore and a validator.
type Handler struct {
	validator      *model.Validator
	smime          pb.SmimeServiceClient
	rejectStudents bool
}

// NewHandler generates a new handler for acting on the domain storage.
func NewHandler(smime pb.SmimeServiceClient, rejectStudents bool) *Handler {
	v := model.NewValidator()
	return &Handler{
		validator:      v,
		smime:          smime,
		rejectStudents: rejectStudents,
	}
}
