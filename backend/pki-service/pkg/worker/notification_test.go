package worker

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hm-edu/pki-service/ent"
	"github.com/hm-edu/pki-service/ent/certificate"
	"github.com/hm-edu/pki-service/ent/enttest"
	"github.com/hm-edu/pki-service/pkg/database"

	// Importing the go-sqlite3 is required to create a sqlite3 database.
	_ "github.com/mattn/go-sqlite3"
)

func TestLoadCertificates(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()
	n := Notifier{
		Db: client,
	}
	d1 := client.Domain.Create().SetFqdn("test.example.com").SaveX(context.Background())
	d2 := client.Domain.Create().SetFqdn("test1.example.com").SaveX(context.Background())
	d3 := client.Domain.Create().SetFqdn("test2.example.com").SaveX(context.Background())
	d4 := client.Domain.Create().SetFqdn("test3.example.com").SaveX(context.Background())

	client.Certificate.Create().SetCommonName("test.example.com").SetNotAfter(time.Now().Add(29*24*time.Hour)).SetStatus(certificate.StatusIssued).AddDomains(d1, d2).SaveX(context.Background())
	client.Certificate.Create().SetCommonName("test2.example.com").SetNotAfter(time.Now().Add(22*24*time.Hour)).SetStatus(certificate.StatusIssued).AddDomains(d2, d3).SaveX(context.Background())
	client.Certificate.Create().SetCommonName("test3.example.com").SetNotAfter(time.Now().Add(31 * 24 * time.Hour)).SetStatus(certificate.StatusIssued).AddDomains(d4).SaveX(context.Background())

	certs, err := n.loadCertificates()
	if err != nil {
		t.Error(err)
	}
	if len(certs) != 2 {
		t.Errorf("Expected 2 certificates, got %d", len(certs))
	}
}
