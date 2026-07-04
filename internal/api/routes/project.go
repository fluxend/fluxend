package routes

import (
	"fluxend/internal/api/handlers"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

func RegisterProjectRoutes(e *echo.Echo, container *do.Injector, authMiddleware echo.MiddlewareFunc, allowProjectMiddleware echo.MiddlewareFunc) {
	projectController := do.MustInvoke[*handlers.ProjectHandler](container)
	statHandler := do.MustInvoke[*handlers.StatHandler](container)

	projectsGroup := e.Group("projects", authMiddleware, allowProjectMiddleware)

	projectsGroup.POST("", projectController.Store)
	projectsGroup.GET("", projectController.List)
	projectsGroup.GET("/:projectUUID", projectController.Show)
	projectsGroup.PUT("/:projectUUID", projectController.Update)
	projectsGroup.DELETE("/:projectUUID", projectController.Delete)
	projectsGroup.GET("/:projectUUID/openapi", projectController.GenerateOpenAPI)
	projectsGroup.GET("/:projectUUID/logs", projectController.ListLogs)
	projectsGroup.GET("/:projectUUID/token", projectController.Token)

	projectsGroup.GET("/:projectUUID/stats", statHandler.Retrieve)

	// track postgrest requests
	e.GET("projects/:dbName/logs/capture", projectController.StoreLogs)
}
