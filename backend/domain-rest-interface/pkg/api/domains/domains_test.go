package domains

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hm-edu/domain-rest-interface/ent"
	"github.com/hm-edu/domain-rest-interface/ent/enttest"
	"github.com/hm-edu/domain-rest-interface/pkg/database"
	"github.com/hm-edu/domain-rest-interface/pkg/model"
	"github.com/hm-edu/domain-rest-interface/pkg/store"
	pb "github.com/hm-edu/portal-apis"

	"github.com/hm-edu/portal-common/helper"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	// Importing the go-sqlite3 is required to create a sqlite3 database.
	_ "github.com/mattn/go-sqlite3"
)

type MockPkiService struct {
}

func (s *MockPkiService) RevokeCertificate(context.Context, *pb.RevokeSslRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *MockPkiService) IssueCertificate(context.Context, *pb.IssueSslRequest, ...grpc.CallOption) (*pb.IssueSslResponse, error) {
	return &pb.IssueSslResponse{}, nil
}

func (s *MockPkiService) ListCertificates(context.Context, *pb.ListSslRequest, ...grpc.CallOption) (*pb.ListSslResponse, error) {
	return &pb.ListSslResponse{}, nil
}

func (s *MockPkiService) CertificateDetails(context.Context, *pb.CertificateDetailsRequest, ...grpc.CallOption) (*pb.SslCertificateDetails, error) {
	return &pb.SslCertificateDetails{}, nil
}
func TestCreateDomainsWithoutTokenAndMiddleware(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(store.NewDomainStore(client), &MockPkiService{})
	assert.Panics(t, func() { _ = h.CreateDomain(c) })
}

func TestSimplePermssions(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()
	st := store.NewDomainStore(client)
	h := NewHandler(st, &MockPkiService{})

	_, _ = st.Create(context.Background(), &ent.Domain{Fqdn: "example.com", Owner: "test", Approved: false})

	domains, err := h.enumerateDomains(context.Background(), "test")
	assert.NoError(t, err)
	assert.Len(t, domains, 1)
	assert.Equal(t, &model.Domain{Delegations: []*model.Delegation{}, FQDN: "example.com", ID: 1, Owner: "test", Approved: false, Permissions: model.Permissions{CanDelete: true, CanTransfer: false, CanApprove: false, CanDelegate: false}}, domains[0])
}

func TestDeletePermissions(t *testing.T) {

	tc := []struct {
		ID      int
		User    string
		Success bool
	}{{ID: 3, User: "max", Success: false}, {ID: 4, User: "max", Success: true}, {ID: 3, User: "test", Success: false}, {ID: 4, User: "test", Success: false}, {ID: 5, User: "test", Success: false}, {ID: 1, User: "max", Success: true}}

	for _, c := range tc {

		t.Run(fmt.Sprintf("%v_%v", c.ID, c.User), func(t *testing.T) {
			e := echo.New()
			client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
			defer func(*ent.Client) {
				_ = client.Close()
			}(client)
			database.DB.Internal, _, _ = sqlmock.New()
			st := store.NewDomainStore(client)
			h := NewHandler(st, &MockPkiService{})

			_, _ = st.Create(context.Background(), &ent.Domain{Fqdn: "bar.example.com", Owner: "test", Approved: false})
			_, _ = st.Create(context.Background(), &ent.Domain{Fqdn: "foo.example.com", Owner: "test", Approved: true})
			_, _ = st.Create(context.Background(), &ent.Domain{Fqdn: "example.com", Owner: "max", Approved: true})
			_, _ = st.Create(context.Background(), &ent.Domain{Fqdn: "test.bar.example.com", Owner: "max", Approved: true})
			_, _ = st.Create(context.Background(), &ent.Domain{Fqdn: "test.example.com", Owner: "max", Approved: true})
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/%d", c.ID), nil)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			ctx.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": c.User}})
			ctx.SetPath("/:id")
			ctx.SetParamNames("id")
			ctx.SetParamValues(fmt.Sprint(c.ID))

			resp := h.DeleteDomain(ctx)
			if c.Success {
				assert.NoError(t, resp)
			} else {
				assert.Error(t, resp)
			}
		})
	}
}

func TestListHandler(t *testing.T) {

	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()
	st := store.NewDomainStore(client)
	h := NewHandler(st, &MockPkiService{})
	_, _ = st.Create(context.Background(), &ent.Domain{Fqdn: "bar.example.com", Owner: "test", Approved: false})
	_, _ = st.Create(context.Background(), &ent.Domain{Fqdn: "foo.example.com", Owner: "test", Approved: true})
	_, _ = st.Create(context.Background(), &ent.Domain{Fqdn: "example.com", Owner: "max", Approved: true})
	_, _ = st.Create(context.Background(), &ent.Domain{Fqdn: "test.bar.example.com", Owner: "max", Approved: true})
	_, _ = st.Create(context.Background(), &ent.Domain{Fqdn: "test.example.com", Owner: "max", Approved: true})
	req := httptest.NewRequest(http.MethodDelete, "/", nil)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
	err := h.ListDomains(ctx)
	assert.NoError(t, err)
	list := []model.Domain{}
	err = json.Unmarshal(rec.Body.Bytes(), &list)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestCreateDomainsFqdns(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()

	testcases := []string{"{}", "{fqdn:1}", `{fqdn:"%:"}`}
	for _, tc := range testcases {
		t.Run(fmt.Sprintf("TestCreateDomains %s", tc), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
			h := NewHandler(store.NewDomainStore(client), &MockPkiService{})
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
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
	h := NewHandler(store.NewDomainStore(client), &MockPkiService{})
	resp := h.CreateDomain(c)
	if assert.NoError(t, resp) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// The ugly linebreak is required for the json comparison
		assert.Equal(t, `{"id":1,"fqdn":"example.com","owner":"test","delegations":[],"approved":false,"permissions":{"can_delete":true,"can_approve":false,"can_transfer":false,"can_delegate":false}}
`, rec.Body.String())
	}
}

func TestCreateDomainsTwice(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
	h := NewHandler(store.NewDomainStore(client), &MockPkiService{})
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
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
	st := store.NewDomainStore(client)
	_, err := st.Create(c.Request().Context(), &ent.Domain{Fqdn: "example.com", Owner: "test", Approved: true})
	assert.NoError(t, err)
	h := NewHandler(st, &MockPkiService{})
	resp := h.CreateDomain(c)
	if assert.NoError(t, resp) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// The ugly linebreak is required for the json comparison
		assert.Equal(t, `{"id":2,"fqdn":"foo.example.com","owner":"test","delegations":[],"approved":true,"permissions":{"can_delete":true,"can_approve":false,"can_transfer":true,"can_delegate":true}}
`, rec.Body.String())
	}
}

func TestCreateDomainsNoAutoApproveOtherUser(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "max"}})
	st := store.NewDomainStore(client)
	_, err := st.Create(c.Request().Context(), &ent.Domain{Fqdn: "example.com", Owner: "test", Approved: true})
	assert.NoError(t, err)
	h := NewHandler(st, &MockPkiService{})
	resp := h.CreateDomain(c)
	if assert.NoError(t, resp) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// The ugly linebreak is required for the json comparison
		assert.Equal(t, `{"id":2,"fqdn":"foo.example.com","owner":"max","delegations":[],"approved":false,"permissions":{"can_delete":true,"can_approve":false,"can_transfer":false,"can_delegate":false}}
`, rec.Body.String())
	}
}

func TestCreateDomainsNoAutoApproveOtherUserChild(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()

	// Prepare root zone
	st := store.NewDomainStore(client)
	_, err := st.Create(context.Background(), &ent.Domain{Fqdn: "example.com", Owner: "test", Approved: true})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "max"}})
	h := NewHandler(st, &MockPkiService{})
	resp := h.CreateDomain(c)
	if assert.NoError(t, resp) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// The ugly linebreak is required for the json comparison
		assert.Equal(t, `{"id":2,"fqdn":"foo.example.com","owner":"max","delegations":[],"approved":false,"permissions":{"can_delete":true,"can_approve":false,"can_transfer":false,"can_delegate":false}}
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
		assert.Equal(t, `{"id":3,"fqdn":"mail.foo.example.com","owner":"test","delegations":[],"approved":true,"permissions":{"can_delete":true,"can_approve":false,"can_transfer":true,"can_delegate":true}}
`, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"2.mail.foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "max"}})
	resp = h.CreateDomain(c)
	if assert.NoError(t, resp) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// The ugly linebreak is required for the json comparison
		assert.Equal(t, `{"id":4,"fqdn":"2.mail.foo.example.com","owner":"max","delegations":[],"approved":false,"permissions":{"can_delete":true,"can_approve":false,"can_transfer":false,"can_delegate":false}}
`, rec.Body.String())
	}

	list, err := st.ListDomains(c.Request().Context(), "test", false, true)
	assert.NoError(t, err)
	assert.Len(t, list, 4)
	assert.False(t, helper.First(list, func(d *ent.Domain) bool { return d.Fqdn == "foo.example.com" }).Approved)
	assert.False(t, helper.First(list, func(d *ent.Domain) bool { return d.Fqdn == "2.mail.foo.example.com" }).Approved)

	list, err = st.ListDomains(c.Request().Context(), "max", false, true)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.False(t, helper.First(list, func(d *ent.Domain) bool { return d.Fqdn == "foo.example.com" }).Approved)
	assert.False(t, helper.First(list, func(d *ent.Domain) bool { return d.Fqdn == "2.mail.foo.example.com" }).Approved)

}

func TestApproveDomainsNotAllowed(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()

	st := store.NewDomainStore(client)
	_, err := st.Create(context.Background(), &ent.Domain{Fqdn: "example.com", Owner: "test", Approved: true})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "max"}})
	assert.NoError(t, err)
	h := NewHandler(st, &MockPkiService{})
	_ = h.CreateDomain(c)

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"mail.foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c = e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"2.mail.foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c = e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "max"}})
	_ = h.CreateDomain(c)

	for i := 1; i <= 4; i++ {
		t.Run(fmt.Sprintf("Approve %d", i), func(t *testing.T) {
			req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, rec)
			c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "max"}})

			c.SetPath("/:id")
			c.SetParamNames("id")
			c.SetParamValues(fmt.Sprint(i))
			resp := h.ApproveDomain(c)
			assert.Error(t, resp)
		})
	}
}

func TestApproveMissingId(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()

	st := store.NewDomainStore(client)
	_, err := st.Create(context.Background(), &ent.Domain{Fqdn: "example.com", Owner: "test", Approved: true})
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
	h := NewHandler(st, &MockPkiService{})
	resp := h.ApproveDomain(c)
	assert.Error(t, resp)
}

func TestApproveDomainsAllowed(t *testing.T) {
	e := echo.New()
	client := enttest.Open(t, "sqlite3", "file:db?mode=memory&cache=shared&_fk=1")
	defer func(*ent.Client) {
		_ = client.Close()
	}(client)
	database.DB.Internal, _, _ = sqlmock.New()

	st := store.NewDomainStore(client)
	_, err := st.Create(context.Background(), &ent.Domain{Fqdn: "example.com", Owner: "test", Approved: true})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "max"}})
	h := NewHandler(st, &MockPkiService{})
	_ = h.CreateDomain(c)

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"mail.foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c = e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
	_ = h.CreateDomain(c)

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"fqdn":"2.mail.foo.example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c = e.NewContext(req, rec)
	c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "max"}})
	_ = h.CreateDomain(c)

	t.Run(fmt.Sprintf("Approve %d", 2), func(t *testing.T) {
		req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c = e.NewContext(req, rec)
		c.SetPath("/:id")
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(2))
		c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "test"}})
		resp := h.ApproveDomain(c)
		assert.NoError(t, resp)
	})

	t.Run(fmt.Sprintf("Approve %d", 4), func(t *testing.T) {
		req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c = e.NewContext(req, rec)
		c.SetPath("/:id")
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(4))
		c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"email": "max"}})
		resp := h.ApproveDomain(c)
		assert.NoError(t, resp)
	})

}
