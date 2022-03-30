package store

import (
	"context"

	"github.com/hm-edu/domain-service/ent"
	"github.com/hm-edu/domain-service/ent/delegation"
	"github.com/hm-edu/domain-service/ent/domain"
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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return domains, nil
}

func (s *DomainStore) CreateDomain(ctx context.Context, d *ent.Domain) error {
	return s.db.Domain.Create().SetFqdn(d.Fqdn).SetOwner(d.Owner).Exec(ctx)
}

func (s *DomainStore) DeleteDomain(ctx context.Context, d *ent.Domain) error {
	return s.db.Domain.DeleteOneID(d.ID).Exec(ctx)
}
