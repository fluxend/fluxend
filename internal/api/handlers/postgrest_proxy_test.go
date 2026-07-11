package handlers

import (
	"fluxend/internal/domain/auth"
	"fluxend/internal/domain/logging"
	"fluxend/internal/domain/shared"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

type stubPostgrestService struct {
	shared.PostgrestService
	endpoint string
}

func (s *stubPostgrestService) Endpoint(dbName string) (string, bool) {
	if s.endpoint == "" {
		return "", false
	}

	return s.endpoint, true
}

type stubLogService struct{}

func (s *stubLogService) List(
	listInput *logging.ListInput,
	paginationParams shared.PaginationParams,
	authUser auth.User,
) ([]logging.RequestLog, shared.PaginationDetails, error) {
	return nil, shared.PaginationDetails{}, nil
}

func (s *stubLogService) Store(storeInput *logging.StoreInput) {}

func newProxyContext(e *echo.Echo, dbName, restPath, query string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rest/%s/%s?%s", dbName, restPath, query), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rest/:dbName/*")
	c.SetParamNames("dbName", "*")
	c.SetParamValues(dbName, restPath)

	return c, rec
}

func TestProxyForwardsPathAndQuery(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s|%s", r.URL.Path, r.URL.RawQuery)
	}))
	defer backend.Close()

	handler := &PostgrestProxyHandler{
		postgrestService: &stubPostgrestService{endpoint: backend.URL},
		logService:       &stubLogService{},
	}

	c, rec := newProxyContext(echo.New(), "udbtest", "items", "select=id")

	if err := handler.Proxy(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if rec.Body.String() != "/items|select=id" {
		t.Fatalf("unexpected forwarded request: %s", rec.Body.String())
	}
}

func TestProxyReturnsNotFoundWhenProcessMissing(t *testing.T) {
	handler := &PostgrestProxyHandler{
		postgrestService: &stubPostgrestService{},
		logService:       &stubLogService{},
	}

	c, rec := newProxyContext(echo.New(), "udbmissing", "items", "")

	if err := handler.Proxy(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
