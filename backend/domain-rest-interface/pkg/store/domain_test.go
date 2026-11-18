package store

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hm-edu/domain-rest-interface/ent"
	"github.com/hm-edu/domain-rest-interface/ent/enttest"
	"github.com/hm-edu/domain-rest-interface/pkg/database"
	"github.com/stretchr/testify/assert"

	// Importing the go-sqlite3 is required to create a sqlite3 database.
	_ "github.com/mattn/go-sqlite3"
)

func TestListDomainsMatchesUsersIgnoringCase(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()
	st := NewDomainStore(client)
	ctx := context.Background()

	// Rows written before owners and delegations were normalized may still
	// carry mixed case and whitespace; they are inserted via the raw client
	// to bypass the normalization in the store.
	legacy, err := client.Domain.Create().SetFqdn("legacy.example.com").SetOwner("Legacy.Owner@HM.EDU").SetApproved(true).Save(ctx)
	assert.NoError(t, err)
	_, err = client.Delegation.Create().SetDomain(legacy).SetUser("Some.User@HM.EDU").Save(ctx)
	assert.NoError(t, err)

	// Rows written through the store are normalized.
	fresh, err := st.Create(ctx, &ent.Domain{Fqdn: "fresh.example.com", Owner: " Fresh.Owner@HM.EDU ", Approved: true})
	assert.NoError(t, err)
	assert.Equal(t, "fresh.owner@hm.edu", fresh.Owner)
	fresh, err = st.AddDelegation(ctx, fresh.ID, " Fresh.User@HM.EDU ")
	assert.NoError(t, err)
	if assert.Len(t, fresh.Edges.Delegations, 1) {
		assert.Equal(t, "fresh.user@hm.edu", fresh.Edges.Delegations[0].User)
	}

	tc := []struct {
		user string
		want []string
	}{
		{user: "legacy.owner@hm.edu", want: []string{"legacy.example.com"}},
		{user: "LEGACY.OWNER@hm.edu", want: []string{"legacy.example.com"}},
		{user: " some.user@hm.edu ", want: []string{"legacy.example.com"}},
		{user: "Fresh.Owner@hm.edu", want: []string{"fresh.example.com"}},
		{user: "FRESH.USER@HM.EDU", want: []string{"fresh.example.com"}},
		// "_" must not act as a wildcard.
		{user: "some_user@hm.edu", want: []string{}},
		{user: "nobody@hm.edu", want: []string{}},
	}
	for _, c := range tc {
		t.Run(c.user, func(t *testing.T) {
			domains, err := st.ListDomains(ctx, c.user, true, false)
			assert.NoError(t, err)
			got := []string{}
			for _, d := range domains {
				got = append(got, d.Fqdn)
			}
			assert.Equal(t, c.want, got)
		})
	}
}
