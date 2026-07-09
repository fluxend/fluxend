package handlers

import (
	"fluxend/internal/api/response"
	"fluxend/internal/domain/logging"
	"fluxend/internal/domain/shared"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

type PostgrestProxyHandler struct {
	postgrestService shared.PostgrestService
	logService       logging.Service
}

func NewPostgrestProxyHandler(injector *do.Injector) (*PostgrestProxyHandler, error) {
	postgrestService := do.MustInvoke[shared.PostgrestService](injector)
	logService := do.MustInvoke[logging.Service](injector)

	return &PostgrestProxyHandler{
		postgrestService: postgrestService,
		logService:       logService,
	}, nil
}

func (h *PostgrestProxyHandler) Proxy(c echo.Context) error {
	dbName := c.Param("dbName")

	endpoint, running := h.postgrestService.Endpoint(dbName)
	if !running {
		return response.NotFoundResponse(c, "Project API is not available")
	}

	target, err := url.Parse(endpoint)
	if err != nil {
		return response.ErrorResponse(c, err)
	}

	request := c.Request()
	forwardedPath := "/" + c.Param("*")

	go h.logService.Store(&logging.StoreInput{
		Endpoint:  forwardedPath,
		DbName:    dbName,
		IPAddress: c.RealIP(),
		Host:      request.Host,
		Method:    request.Method,
	})

	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(r *http.Request) {
		director(r)
		r.URL.Path = forwardedPath
		r.URL.RawPath = ""
	}

	proxy.ServeHTTP(c.Response(), request)

	return nil
}
