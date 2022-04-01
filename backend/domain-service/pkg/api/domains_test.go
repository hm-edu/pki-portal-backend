package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hm-edu/domain-service/ent"
	"github.com/hm-edu/domain-service/ent/enttest"
	"github.com/hm-edu/domain-service/pkg/store"
	"github.com/hm-edu/portal-common/helper"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	// Importing the go-sqlite3 is required to create a sqlite3 database.
	_ "github.com/mattn/go-sqlite3"
)

func TestCreateDomainsWithoutTokenAndMiddleware(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(store.NewDomainStore(client))
	assert.Panics(t, func() { _ = h.CreateDomain(c) })
}

func TestCreateDomainsFqdns(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	testcases := []string{"{}", "{fqdn:1}", `{fqdn:"%:"}`}
	for _, tc := range testcases {
		t.Run(fmt.Sprintf("TestCreateDomains %s", tc), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
			h := NewHandler(store.NewDomainStore(client))
			resp := h.CreateDomain(c)
			if assert.Error(t, resp) {
				assert.Equal(t, http.StatusBadRequest, resp.(*echo.HTTPError).Code)
			}
		})
	}
	assert.Len(t, client.Domain.Query().AllX(context.Background()), 0)
}
func TestCreateDomainsWithToken(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
	h := NewHandler(store.NewDomainStore(client))
	resp := h.CreateDomain(c)
	if assert.NoError(t, resp) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// The ugly linebreak is required for the json comparison
		assert.Equal(t, `{"id":1,"fqdn":"example.com","owner":"test","delegations":[],"approved":false}
`, rec.Body.String())
	}
}

func TestCreateDomainsTwice(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
	h := NewHandler(store.NewDomainStore(client))
	resp := h.CreateDomain(c)
	if assert.NoError(t, resp) {
		assert.Equal(t, http.StatusCreated, rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
	resp = h.CreateDomain(c)
	if assert.Error(t, resp) {
		assert.Equal(t, http.StatusBadRequest, resp.(*echo.HTTPError).Code)
	}
}

func TestCreateDomainsAutoApprove(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
	st := store.NewDomainStore(client)
	_, err := st.Create(c.Request().Context(), &ent.Domain{Fqdn: "example.com", Owner: "test", Approved: true})
	assert.NoError(t, err)
	h := NewHandler(st)
	resp := h.CreateDomain(c)
	if assert.NoError(t, resp) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// The ugly linebreak is required for the json comparison
		assert.Equal(t, `{"id":2,"fqdn":"foo.example.com","owner":"test","delegations":[],"approved":true}
`, rec.Body.String())
	}
}

func TestCreateDomainsNoAutoApproveOtherUser(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "max"}})
	st := store.NewDomainStore(client)
	_, err := st.Create(c.Request().Context(), &ent.Domain{Fqdn: "example.com", Owner: "test", Approved: true})
	assert.NoError(t, err)
	h := NewHandler(st)
	resp := h.CreateDomain(c)
	if assert.NoError(t, resp) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// The ugly linebreak is required for the json comparison
		assert.Equal(t, `{"id":2,"fqdn":"foo.example.com","owner":"max","delegations":[],"approved":false}
`, rec.Body.String())
	}
}

func TestCreateDomainsNoAutoApproveOtherUserChild(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "max"}})
	st := store.NewDomainStore(client)
	_, err := st.Create(c.Request().Context(), &ent.Domain{Fqdn: "example.com", Owner: "test", Approved: true})
	assert.NoError(t, err)
	h := NewHandler(st)
	resp := h.CreateDomain(c)
	if assert.NoError(t, resp) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// The ugly linebreak is required for the json comparison
		assert.Equal(t, `{"id":2,"fqdn":"foo.example.com","owner":"max","delegations":[],"approved":false}
`, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"mail.foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
	resp = h.CreateDomain(c)
	if assert.NoError(t, resp) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// The ugly linebreak is required for the json comparison
		assert.Equal(t, `{"id":3,"fqdn":"mail.foo.example.com","owner":"test","delegations":[],"approved":true}
`, rec.Body.String())
	}

	list, err := st.ListDomains(c.Request().Context(), "test")
	assert.NoError(t, err)
	assert.Len(t, list, 3)
	assert.False(t, helper.First(list, func(d *ent.Domain) bool { return d.Fqdn == "foo.example.com" }).Approved)
}
