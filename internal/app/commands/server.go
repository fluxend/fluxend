package commands

import (
	"fluxend/internal/api/middlewares"
	"fluxend/internal/api/routes"
	"fluxend/internal/app"
	"fluxend/internal/config/constants"
	"fluxend/internal/domain/logging"
	"fluxend/internal/domain/setting"
	"fluxend/internal/domain/user"
	"fmt"
	"github.com/getsentry/sentry-go"
	echoSentry "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"github.com/spf13/cobra"
	echoSwagger "github.com/swaggo/echo-swagger"
	"os"
	"strings"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Fluxend API server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func startServer() {
	e := SetupServer(app.InitializeContainer())
	validateEnvVariables()

	e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}

func SetupServer(container *do.Injector) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.CORSWithConfig(getCorsConfig()))
	e.Use(middleware.Recover())

	if os.Getenv("SENTRY_DSN") != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              os.Getenv("SENTRY_DSN"),
			TracesSampleRate: 1.0,
		}); err != nil {
			fmt.Printf("Sentry initialization failed: %v\n", err)
		}

		e.Use(echoSentry.New(echoSentry.Options{}))
	}

	registerRoutes(e, container)

	return e
}

func getCorsConfig() middleware.CORSConfig {
	return middleware.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) {
			if isOriginAllowed(origin) {
				return true, nil
			}
			return false, nil
		},
		AllowMethods: []string{
			echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"content-type",
			"Accept",
			"Authorization",
			"authorization",
			"X-Project",
			"x-project",
			"Content-Range",
			"Range-Unit",
			"range",
			"Prefer",
		},
		ExposeHeaders: []string{
			echo.HeaderContentLength, echo.HeaderContentType,
		},
		AllowCredentials: true,
	}
}

func registerRoutes(e *echo.Echo, container *do.Injector) {
	settingService := do.MustInvoke[setting.Service](container)
	userRepo := do.MustInvoke[user.Repository](container)

	authMiddleware := middlewares.Authentication(userRepo)
	allowProjectMiddleware := middlewares.AllowProject(settingService)
	allowFormMiddleware := middlewares.AllowForm(settingService)
	allowStorageMiddleware := middlewares.AllowStorage(settingService)
	allowBackupMiddleware := middlewares.AllowBackup(settingService)

	requestLogRepo := do.MustInvoke[logging.Repository](container)
	requestLogMiddleware := middlewares.RequestLogger(requestLogRepo)
	e.Use(requestLogMiddleware)

	routes.RegisterUserRoutes(e, container, authMiddleware)
	routes.RegisterAdminRoutes(e, container, authMiddleware)
	routes.RegisterOrganizationRoutes(e, container, authMiddleware)
	routes.RegisterProjectRoutes(e, container, authMiddleware, allowProjectMiddleware)
	routes.RegisterTableRoutes(e, container, authMiddleware)
	routes.RegisterRowRoutes(e, container, authMiddleware)
	routes.RegisterWebhookRoutes(e, container, authMiddleware)
	routes.RegisterFormRoutes(e, container, authMiddleware, allowFormMiddleware)
	routes.RegisterStorageRoutes(e, container, authMiddleware, allowStorageMiddleware)
	routes.RegisterFunctionRoutes(e, container, authMiddleware)
	routes.RegisterBackup(e, container, authMiddleware, allowBackupMiddleware)

	e.GET("/", func(c echo.Context) error {
		response := map[string]string{
			"message": "Welcome to Fluxend API",
		}

		return c.JSON(200, response)
	})

	e.GET("/docs/*", echoSwagger.WrapHandler)
}

func validateEnvVariables() {
	requiredVars := []string{
		"APP_ENV",
		"BASE_URL",
		"CONSOLE_URL",
		"API_URL",
		"API_CONTAINER_NAME",
		"DATABASE_CONTAINER_NAME",
		"FRONTEND_CONTAINER_NAME",
		"DATABASE_HOST",
		"DATABASE_USER",
		"DATABASE_PASSWORD",
		"DATABASE_NAME",
		"JWT_SECRET",
		"STORAGE_DRIVER",
		"POSTGREST_DB_HOST",
		"POSTGREST_DB_USER",
		"POSTGREST_DB_PASSWORD",
		"POSTGREST_DEFAULT_SCHEMA",
		"POSTGREST_DEFAULT_ROLE",
	}

	for _, envVar := range requiredVars {
		if os.Getenv(envVar) == "" {
			log.Fatal().Msg(fmt.Sprintf("Environment variable %s is required but not set", envVar))
		}
	}

	if len(os.Getenv("JWT_SECRET")) < constants.JWTSecretMinLength {
		log.Fatal().Msg(fmt.Sprintf("JWT_SECRET must be at least %d characters long", constants.JWTSecretMinLength))
	}
}

func isOriginAllowed(origin string) bool {
	customOrigins := strings.Split(os.Getenv("CUSTOM_ORIGINS"), ",")

	allowedOrigins := []string{
		os.Getenv("CONSOLE_URL"),
		os.Getenv("BASE_URL"),
	}

	allowedOrigins = append(allowedOrigins, customOrigins...)

	log.Info().
		Str("origin", origin).
		Strs("allowedOrigins", allowedOrigins).
		Str("CONSOLE_URL", os.Getenv("CONSOLE_URL")).
		Str("CUSTOM_ORIGINS", os.Getenv("CUSTOM_ORIGINS")).
		Msg("CORS origin check")

	for _, allowedOrigin := range allowedOrigins {
		allowedOrigin = strings.TrimSpace(allowedOrigin)
		if allowedOrigin == "" {
			continue
		}

		log.Info().
			Str("checking", allowedOrigin).
			Str("against", origin).
			Bool("exact_match", origin == allowedOrigin).
			Msg("CORS comparison")

		if origin == allowedOrigin || strings.HasSuffix(origin, allowedOrigin) {
			return true
		}
	}

	return false
}
