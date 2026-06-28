package postgrest

import (
	"fluxend/internal/config/constants"
	"fluxend/internal/domain/project"
	"fluxend/internal/domain/shared"
	"fluxend/pkg"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/samber/do"
	"os"
	"strings"
)

const (
	ImageName = "postgrest/postgrest"
)

type Config struct {
	DBUser        string
	DBPassword    string
	DBHost        string
	DBSchema      string
	DBRole        string
	BaseDomain    string
	CustomOrigins string
	URLScheme     string
}

type ServiceImpl struct {
	projectRepo project.Repository
	config      *Config
}

func NewPostgrestService(injector *do.Injector) (shared.PostgrestService, error) {
	corsOrigins := os.Getenv("CUSTOM_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*"
	}

	config := &Config{
		DBUser:        os.Getenv("POSTGREST_DB_USER"),
		DBPassword:    os.Getenv("POSTGREST_DB_PASSWORD"),
		DBHost:        os.Getenv("POSTGREST_DB_HOST"),
		DBSchema:      os.Getenv("POSTGREST_DEFAULT_SCHEMA"),
		DBRole:        os.Getenv("POSTGREST_DEFAULT_ROLE"),
		BaseDomain:    os.Getenv("BASE_DOMAIN"),
		URLScheme:     os.Getenv("URL_SCHEME"),
		CustomOrigins: corsOrigins,
	}

	projectRepo := do.MustInvoke[project.Repository](injector)

	return &ServiceImpl{
		projectRepo: projectRepo,
		config:      config,
	}, nil
}

func (s *ServiceImpl) StartContainer(dbName, jwtSecret string) {
	if err := pkg.ExecuteCommand(s.buildStartCommand(dbName, jwtSecret)); err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to start container")

		_, err := s.projectRepo.UpdateStatusByDatabaseName(dbName, constants.ProjectStatusError)
		if err != nil {
			log.Error().
				Str("action", constants.ActionPostgrest).
				Str("dbName", dbName).
				Str("error", err.Error()).
				Msg("failed to update project status to error")

			return
		}

	}

	// Update project status to active
	_, err := s.projectRepo.UpdateStatusByDatabaseName(dbName, constants.ProjectStatusActive)
	if err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to update project status to active")

		return
	}

	log.Info().
		Str("action", constants.ActionPostgrest).
		Str("db", dbName).
		Msg("container started successfully")
}

func (s *ServiceImpl) RemoveContainer(dbName string) {
	containerName := s.getContainerName(dbName)

	if err := pkg.ExecuteCommand([]string{"docker", "stop", containerName}); err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to stop container")
	}

	if err := pkg.ExecuteCommand([]string{"docker", "rm", containerName}); err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to remove container")
	}

	_, err := s.projectRepo.UpdateStatusByDatabaseName(dbName, constants.ProjectStatusInactive)
	if err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to update project status to inactive")

		return
	}
}

func (s *ServiceImpl) HasContainer(dbName string) bool {
	cmd := []string{"docker", "inspect", "--format='{{.State.Running}}'", s.getContainerName(dbName)}
	output, err := pkg.ExecuteCommandWithOutput(cmd)
	if err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to check if container exists")

		return false
	}

	return strings.Contains(output, "true")
}

func (s *ServiceImpl) buildStartCommand(dbName, jwtSecret string) []string {
	response := []string{
		"docker", "run", "-d", "--name", s.getContainerName(dbName),
		"--network", "fluxend_network",
		"-e", fmt.Sprintf("PGRST_DB_URI=postgres://%s:%s@%s/%s", s.config.DBUser, s.config.DBPassword, s.config.DBHost, dbName),
		"-e", "PGRST_DB_ANON_ROLE=" + s.config.DBRole,
		"-e", "PGRST_DB_SCHEMA=" + s.config.DBSchema,
		"-e", "PGRST_JWT_SECRET=" + jwtSecret,
		"-e", "PGRST_SERVER_CORS_ALLOWED_ORIGINS=" + s.config.CustomOrigins,
		"-e", "PGRST_SERVER_CORS_ALLOWED_HEADERS=*",
		"-e", "PGRST_SERVER_CORS_ALLOWED_METHODS=GET,POST,PATCH,PUT,DELETE,OPTIONS,HEAD",
		"--label", "traefik.enable=true",
		"--label", fmt.Sprintf("traefik.http.routers.%s.rule=Host(`%s.%s`)", dbName, dbName, s.config.BaseDomain),
		"--label", fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=3000", dbName),

		// Add tracking middleware
		"--label", fmt.Sprintf("traefik.http.routers.%s.middlewares=track-%s", dbName, dbName),
		"--label", fmt.Sprintf("traefik.http.middlewares.track-%s.forwardauth.address=%s://api.%s/projects/%s/logs/capture", dbName, s.config.URLScheme, s.config.BaseDomain, dbName),
	}

	if s.config.URLScheme == "https" {
		response = append(response,
			"--label", fmt.Sprintf("traefik.http.routers.%s.tls=true", dbName),
			"--label", fmt.Sprintf("traefik.http.routers.%s.tls.certresolver=le", dbName),
			"--label", fmt.Sprintf("traefik.http.routers.%s.entrypoints=websecure", dbName),
		)
	} else {
		response = append(response, "--label", fmt.Sprintf("traefik.http.routers.%s.entrypoints=web", dbName))
	}

	response = append(response, ImageName)
	return response
}

func (s *ServiceImpl) RefreshSchemaCache(dbName string) {
	// Why we need this: https://postgrest.org/en/v10/schema_cache.html
	// docker kill -s SIGUSR1 <container>

	containerName := s.getContainerName(dbName)

	cmd := []string{"docker", "kill", "-s", "SIGUSR1", containerName}
	if err := pkg.ExecuteCommand(cmd); err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to refresh schema cache")

		return
	}
}

func (s *ServiceImpl) getContainerName(dbName string) string {
	return fmt.Sprintf("postgrest_%s", dbName)
}
