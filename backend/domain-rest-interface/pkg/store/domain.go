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
	"github.com/hm-edu/domain-rest-interface/pkg/model"
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
	q := s.db.Domain.Query()
	if approved {
		err := q.Where(domain.Approved(approved)).Select(domain.FieldFqdn).Scan(ctx, &domains)
		if err != nil {
			return nil, err
		}
	} else {
		err := q.Select(domain.FieldFqdn).Scan(ctx, &domains)
		if err != nil {
			return nil, err
		}
	}
	return domains, nil
}

// ListDomains returns all domains that are owned or delegated to one user
func (s *DomainStore) ListDomains(ctx context.Context, user string, requireApproval bool, admin bool) ([]*ent.Domain, error) {
	if err := database.DB.Internal.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the database: %w", err)
	}

	var (
		domains []*ent.Domain
		err     error
	)

	if admin {
		domains, err = s.db.Domain.Query().WithDelegations().All(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		predicates := domain.Or(domain.HasDelegationsWith(delegationUserEqualFold(user)), domainOwnerEqualFold(user))
		if requireApproval {
			predicates = domain.And(predicates, domain.Approved(true))
		}
		domains, err = s.db.Domain.Query().WithDelegations().Where(predicates).All(ctx)
		if err != nil {
			return nil, err
		}
	}
	if len(domains) != 0 {
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
	return s.db.Domain.Create().SetFqdn(d.Fqdn).SetOwner(model.NormalizeUser(d.Owner)).SetApproved(d.Approved).Save(ctx)
}

// Owner sets the owner of a domain.
func (s *DomainStore) Owner(ctx context.Context, id int, owner string) (*ent.Domain, error) {
	if err := database.DB.Internal.Ping(); err != nil {
		return nil, fmt.Errorf("pinging the database: %w", err)
	}
	return s.db.Domain.UpdateOneID(id).SetOwner(model.NormalizeUser(owner)).Save(ctx)
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
	err := s.db.Delegation.Create().SetDomainID(d).SetUser(model.NormalizeUser(user)).Exec(ctx)
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

// delegationUserEqualFold matches delegations whose user equals the given one
// ignoring case and surrounding whitespace. Rows written before users were
// normalized may still carry mixed case, so the comparison lowers both sides
// instead of relying on an exact match (or on ILIKE, whose wildcards would
// make "_" in an address match any character).
func delegationUserEqualFold(user string) predicate.Delegation {
	return predicate.Delegation(func(s *sql.Selector) {
		s.Where(lowerEquals(s.C(delegation.FieldUser), user))
	})
}

// domainOwnerEqualFold matches domains whose owner equals the given one ignoring
// case and surrounding whitespace. See delegationUserEqualFold.
func domainOwnerEqualFold(owner string) predicate.Domain {
	return predicate.Domain(func(s *sql.Selector) {
		s.Where(lowerEquals(s.C(domain.FieldOwner), owner))
	})
}

func lowerEquals(column, value string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.WriteString("LOWER(").Ident(column).WriteString(") = ").Arg(model.NormalizeUser(value))
	})
}
