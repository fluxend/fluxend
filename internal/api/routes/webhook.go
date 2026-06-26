package routes

import (
	"fluxend/internal/api/handlers"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

func RegisterWebhookRoutes(e *echo.Echo, container *do.Injector, authMiddleware echo.MiddlewareFunc) {
	webhookController := do.MustInvoke[*handlers.WebhookHandler](container)

	webhooksGroup := e.Group("tables/:fullTableName/webhooks", authMiddleware)

	webhooksGroup.GET("", webhookController.List)
	webhooksGroup.POST("", webhookController.Store)
	webhooksGroup.DELETE("/:webhookUUID", webhookController.Delete)
}
