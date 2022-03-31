package store

import (
	"context"
	"fmt"

	"github.com/hm-edu/domain-service/ent"
	"github.com/hm-edu/domain-service/ent/delegation"
	"github.com/hm-edu/domain-service/ent/domain"
	"github.com/hm-edu/domain-service/ent/predicate"
	"github.com/hm-edu/portal-common/helper"
)

// DomainStore provides the storage of domains
type DomainStore struct {
	db *ent.Client
}

// NewDomainStore creates a new instance of a DomainStore
func NewDomainStore(db *ent.Client) *DomainStore {
	return &DomainStore{
		db: db,
	}
}

// ListDomains returns all domains that are owned or delegated to one user
func (s *DomainStore) ListDomains(ctx context.Context, owner string) ([]*ent.Domain, error) {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, err
	}

	domains, err := tx.Domain.Query().Where(domain.Or(domain.HasDelegationsWith(delegation.User(owner)), domain.Owner(owner))).All(ctx)

	if err != nil {
		return nil, err
	}
	fqdns := helper.Map(domains, func(d *ent.Domain) predicate.Domain { return domain.FqdnHasSuffix("." + d.Fqdn) })

	childs, err := tx.Domain.Query().Where(domain.Or(fqdns...)).All(ctx)

	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	domains = append(domains, childs...)
	return domains, nil
}

func (s *DomainStore) GetDomain(ctx context.Context, fqdn string) (*ent.Domain, error) {
	return s.db.Domain.Query().Where(domain.FqdnEQ(fqdn)).First(ctx)
}

func (s *DomainStore) CreateDomain(ctx context.Context, d *ent.Domain) error {
	return s.db.Domain.Create().SetFqdn(d.Fqdn).SetOwner(d.Owner).SetApproved(d.Approved).Exec(ctx)
}
func (s *DomainStore) Approve(ctx context.Context, d *ent.Domain) error {
	return s.db.Domain.UpdateOne(d).SetApproved(true).Exec(ctx)
}

func (s *DomainStore) DeleteDomains(ctx context.Context, d []*ent.Domain) error {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("starting a transaction: %w", err)
	}

	_, err = s.db.Domain.Delete().Where(domain.IDIn(helper.Map(d, func(domain *ent.Domain) int { return domain.ID })...)).Exec(ctx)

	if err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("%w: %v", err, rerr)
		}
	}

	return tx.Commit()

}
