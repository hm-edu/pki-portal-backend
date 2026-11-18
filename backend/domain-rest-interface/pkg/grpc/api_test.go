package grpc

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hm-edu/domain-rest-interface/ent"
	"github.com/hm-edu/domain-rest-interface/ent/enttest"
	"github.com/hm-edu/domain-rest-interface/pkg/database"
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	pb "github.com/hm-edu/portal-apis"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	// Importing the go-sqlite3 is required to create a sqlite3 database.
	_ "github.com/mattn/go-sqlite3"
)

func TestPermissionsIgnoreUserCase(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()
	st := store.NewDomainStore(client)
	ctx := context.Background()

	d, err := st.Create(ctx, &ent.Domain{Fqdn: "example.com", Owner: "owner@hm.edu", Approved: true})
	assert.NoError(t, err)
	_, err = st.AddDelegation(ctx, d.ID, "Delegate@HM.EDU")
	assert.NoError(t, err)

	api := newDomainAPIServer(st, zap.NewNop(), []string{})

	perms, err := api.CheckPermission(ctx, &pb.CheckPermissionRequest{User: " DELEGATE@hm.edu ", Domains: []string{"example.com", "other.com"}})
	assert.NoError(t, err)
	if assert.Len(t, perms.Permissions, 2) {
		assert.Equal(t, "example.com", perms.Permissions[0].Domain)
		assert.True(t, perms.Permissions[0].Granted)
		assert.Equal(t, "other.com", perms.Permissions[1].Domain)
		assert.False(t, perms.Permissions[1].Granted)
	}

	list, err := api.ListDomains(ctx, &pb.ListDomainsRequest{User: "Delegate@HM.EDU", Approved: true})
	assert.NoError(t, err)
	assert.Equal(t, []string{"example.com"}, list.Domains)

	list, err = api.ListDomains(ctx, &pb.ListDomainsRequest{User: "OWNER@hm.edu", Approved: true})
	assert.NoError(t, err)
	assert.Equal(t, []string{"example.com"}, list.Domains)
}
