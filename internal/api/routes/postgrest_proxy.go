package routes

import (
	"fluxend/internal/api/handlers"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

func RegisterPostgrestProxy(e *echo.Echo, container *do.Injector) {
	proxyHandler := do.MustInvoke[*handlers.PostgrestProxyHandler](container)

	e.Any("/rest/:dbName", proxyHandler.Proxy)
	e.Any("/rest/:dbName/*", proxyHandler.Proxy)
}
