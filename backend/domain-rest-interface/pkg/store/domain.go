package store

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/hm-edu/domain-rest-interface/ent"
	"github.com/hm-edu/domain-rest-interface/ent/delegation"
	"github.com/hm-edu/domain-rest-interface/ent/domain"
	"github.com/hm-edu/domain-rest-interface/ent/predicate"
	"github.com/hm-edu/domain-rest-interface/pkg/database"
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

// ListAllDomains returns all domains.
func (s *DomainStore) ListAllDomains(ctx context.Context, approved bool) ([]string, error) {
	if err := database.DB.Internal.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the database: %w", err)
	}
	var domains []string
	err := s.db.Domain.Query().Where(domain.Approved(approved)).Select(domain.FieldFqdn).Scan(ctx, &domains)
	if err != nil {
		return nil, err
	}
	return domains, nil
}

// ListDomains returns all domains that are owned or delegated to one user
func (s *DomainStore) ListDomains(ctx context.Context, user string, requireApproval, includeChilds bool) ([]*ent.Domain, error) {
	if err := database.DB.Internal.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the database: %w", err)
	}

	predicates := domain.Or(domain.HasDelegationsWith(delegation.User(user)), domain.Owner(user))
	if requireApproval {
		predicates = domain.And(predicates, domain.Approved(true))
	}
	domains, err := s.db.Domain.Query().WithDelegations().Where(predicates).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(domains) != 0 && includeChilds {
		fqdns := helper.Map(domains, func(d *ent.Domain) predicate.Domain {
			if d.Approved {
				return domain.FqdnHasSuffix("." + d.Fqdn)
			}
			return predicate.Domain(func(s *sql.Selector) { s.Where(sql.False()) })
		})
		ids := helper.Map(domains, func(d *ent.Domain) int { return d.ID })
		childs, err := s.db.Domain.Query().WithDelegations().Where(domain.And(domain.Or(fqdns...), domain.IDNotIn(ids...))).All(ctx)
		if err != nil {
			return nil, err
		}
		domains = append(domains, childs...)
	}
	return domains, nil
}

// GetDomainByID tries to find a domain with the given FQDN.
func (s *DomainStore) GetDomainByID(ctx context.Context, id int) (*ent.Domain, error) {
	if err := database.DB.Internal.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the database: %w", err)
	}
	return s.db.Domain.Query().Where(domain.ID(id)).WithDelegations().First(ctx)
}

// GetDomain tries to find a domain with the given FQDN.
func (s *DomainStore) GetDomain(ctx context.Context, fqdn string) (*ent.Domain, error) {
	if err := database.DB.Internal.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the database: %w", err)
	}
	return s.db.Domain.Query().Where(domain.Fqdn(fqdn)).WithDelegations().First(ctx)
}

// Create tries to create a new domain entry.
func (s *DomainStore) Create(ctx context.Context, d *ent.Domain) (*ent.Domain, error) {
	if err := database.DB.Internal.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the database: %w", err)
	}
	return s.db.Domain.Create().SetFqdn(d.Fqdn).SetOwner(d.Owner).SetApproved(d.Approved).Save(ctx)
}

// Owner sets the owner of a domain.
func (s *DomainStore) Owner(ctx context.Context, id int, owner string) (*ent.Domain, error) {
	if err := database.DB.Internal.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the database: %w", err)
	}
	return s.db.Domain.UpdateOneID(id).SetOwner(owner).Save(ctx)
}

// Approve sets the given domain on approved.
func (s *DomainStore) Approve(ctx context.Context, id int) (*ent.Domain, error) {
	if err := database.DB.Internal.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the database: %w", err)
	}
	return s.db.Domain.UpdateOneID(id).SetApproved(true).Save(ctx)
}

// AddDelegation adds a delegation to a domain.
func (s *DomainStore) AddDelegation(ctx context.Context, d int, user string) (*ent.Domain, error) {
	if err := database.DB.Internal.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the database: %w", err)
	}
	err := s.db.Delegation.Create().SetDomainID(d).SetUser(user).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s.db.Domain.Query().Where(domain.ID(d)).WithDelegations().First(ctx)
}

// DeleteDelegation tries to delete a delegation.
func (s *DomainStore) DeleteDelegation(ctx context.Context, domainID, delegationID int) (*ent.Domain, error) {
	if err := database.DB.Internal.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the database: %w", err)
	}
	err := s.db.Delegation.DeleteOneID(delegationID).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s.db.Domain.Query().Where(domain.ID(domainID)).WithDelegations().First(ctx)
}

// Delete tries to delete all passed domains.
func (s *DomainStore) Delete(ctx context.Context, id int) error {
	if err := database.DB.Internal.Ping(); err != nil {
		return fmt.Errorf("pinging the database: %w", err)
	}

	_, err := s.db.Domain.Delete().Where(domain.ID(id)).Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}
