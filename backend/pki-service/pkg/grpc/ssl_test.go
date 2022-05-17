package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/ent/enttest"
	pb "github.com/hm-edu/portal-apis"
	"go.uber.org/zap"

	// Importing the go-sqlite3 is required to create a sqlite3 database.
	_ "github.com/mattn/go-sqlite3"
)

func TestListCertificates(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	server := sslAPIServer{db: client, logger: zap.L()}
	d1, err := client.Domain.Create().SetFqdn("test.com").Save(context.Background())
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	d2, err := client.Domain.Create().SetFqdn("www.test.com").Save(context.Background())
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	_, err = client.Certificate.Create().AddDomainIDs(d1.ID, d2.ID).SetSerial("ABC1").SetIssuedBy("Test").SetCommonName("test.com").Save(context.Background())
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	_, err = client.Certificate.Create().AddDomainIDs(d1.ID).SetSerial("ABC2").SetCreated(time.Now()).SetCommonName("test.com").Save(context.Background())
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	_, err = client.Certificate.Create().AddDomainIDs(d2.ID).SetSerial("ABC3").SetSslId(1).SetCommonName("test.com").Save(context.Background())
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	ret, err := server.ListCertificates(context.TODO(), &pb.ListSslRequest{Domains: []string{"test.com", "www.test.com"}})
	if err != nil {
		t.Error(err)
	}
	if len(ret.Items) != 3 {
		t.Error("Expected 2 certificate, got", len(ret.Items))
	}
	ret, err = server.ListCertificates(context.TODO(), &pb.ListSslRequest{Domains: []string{"test.com"}})
	if err != nil {
		t.Error(err)
	}
	if len(ret.Items) != 1 {
		t.Error("Expected 1 certificate, got", len(ret.Items))
	}
}
