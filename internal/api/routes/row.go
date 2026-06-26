package routes

import (
	"fluxend/internal/api/handlers"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

func RegisterRowRoutes(e *echo.Echo, container *do.Injector, authMiddleware echo.MiddlewareFunc) {
	rowController := do.MustInvoke[*handlers.RowHandler](container)

	rowsGroup := e.Group("tables/:fullTableName/rows", authMiddleware)

	rowsGroup.POST("", rowController.Insert)
	rowsGroup.PATCH("", rowController.Update)
	rowsGroup.DELETE("", rowController.Delete)
}
