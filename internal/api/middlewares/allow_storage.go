package middlewares

import (
	"fluxend/internal/api/response"
	"fluxend/internal/domain/setting"
	"github.com/labstack/echo/v4"
)

func AllowStorage(settingService setting.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !settingService.GetBool("allowStorage") {
				return response.ForbiddenResponse(c, "storage.error.disabled")
			}

			return next(c)
		}
	}
}
