// Code generated by ent, DO NOT EDIT.

package ent

import (
	"time"

	"github.com/hm-edu/domain-rest-interface/ent/delegation"
	"github.com/hm-edu/domain-rest-interface/ent/domain"
	"github.com/hm-edu/domain-rest-interface/ent/schema"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	delegationMixin := schema.Delegation{}.Mixin()
	delegationMixinFields0 := delegationMixin[0].Fields()
	_ = delegationMixinFields0
	delegationFields := schema.Delegation{}.Fields()
	_ = delegationFields
	// delegationDescCreateTime is the schema descriptor for create_time field.
	delegationDescCreateTime := delegationMixinFields0[0].Descriptor()
	// delegation.DefaultCreateTime holds the default value on creation for the create_time field.
	delegation.DefaultCreateTime = delegationDescCreateTime.Default.(func() time.Time)
	// delegationDescUpdateTime is the schema descriptor for update_time field.
	delegationDescUpdateTime := delegationMixinFields0[1].Descriptor()
	// delegation.DefaultUpdateTime holds the default value on creation for the update_time field.
	delegation.DefaultUpdateTime = delegationDescUpdateTime.Default.(func() time.Time)
	// delegation.UpdateDefaultUpdateTime holds the default value on update for the update_time field.
	delegation.UpdateDefaultUpdateTime = delegationDescUpdateTime.UpdateDefault.(func() time.Time)
	// delegationDescUser is the schema descriptor for user field.
	delegationDescUser := delegationFields[0].Descriptor()
	// delegation.UserValidator is a validator for the "user" field. It is called by the builders before save.
	delegation.UserValidator = delegationDescUser.Validators[0].(func(string) error)
	domainMixin := schema.Domain{}.Mixin()
	domainMixinFields0 := domainMixin[0].Fields()
	_ = domainMixinFields0
	domainFields := schema.Domain{}.Fields()
	_ = domainFields
	// domainDescCreateTime is the schema descriptor for create_time field.
	domainDescCreateTime := domainMixinFields0[0].Descriptor()
	// domain.DefaultCreateTime holds the default value on creation for the create_time field.
	domain.DefaultCreateTime = domainDescCreateTime.Default.(func() time.Time)
	// domainDescUpdateTime is the schema descriptor for update_time field.
	domainDescUpdateTime := domainMixinFields0[1].Descriptor()
	// domain.DefaultUpdateTime holds the default value on creation for the update_time field.
	domain.DefaultUpdateTime = domainDescUpdateTime.Default.(func() time.Time)
	// domain.UpdateDefaultUpdateTime holds the default value on update for the update_time field.
	domain.UpdateDefaultUpdateTime = domainDescUpdateTime.UpdateDefault.(func() time.Time)
	// domainDescFqdn is the schema descriptor for fqdn field.
	domainDescFqdn := domainFields[0].Descriptor()
	// domain.FqdnValidator is a validator for the "fqdn" field. It is called by the builders before save.
	domain.FqdnValidator = domainDescFqdn.Validators[0].(func(string) error)
	// domainDescOwner is the schema descriptor for owner field.
	domainDescOwner := domainFields[1].Descriptor()
	// domain.OwnerValidator is a validator for the "owner" field. It is called by the builders before save.
	domain.OwnerValidator = domainDescOwner.Validators[0].(func(string) error)
	// domainDescApproved is the schema descriptor for approved field.
	domainDescApproved := domainFields[2].Descriptor()
	// domain.DefaultApproved holds the default value on creation for the approved field.
	domain.DefaultApproved = domainDescApproved.Default.(bool)
}
