package postgrest

import (
	"fluxend/internal/config/constants"
	"fluxend/internal/domain/project"
	"fluxend/internal/domain/shared"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/samber/do"
)

const defaultBinaryPath = "postgrest"

type Config struct {
	DBUser        string
	DBPassword    string
	DBHost        string
	DBSchema      string
	DBRole        string
	CustomOrigins string
	BinaryPath    string
}

type process struct {
	cmd      *exec.Cmd
	port     int
	stopping bool
}

type ServiceImpl struct {
	projectRepo project.Repository
	config      *Config
	mutex       sync.Mutex
	processes   map[string]*process
}

func NewPostgrestService(injector *do.Injector) (shared.PostgrestService, error) {
	corsOrigins := os.Getenv("CUSTOM_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*"
	}

	binaryPath := os.Getenv("POSTGREST_BINARY")
	if binaryPath == "" {
		binaryPath = defaultBinaryPath
	}

	config := &Config{
		DBUser:        os.Getenv("POSTGREST_DB_USER"),
		DBPassword:    os.Getenv("POSTGREST_DB_PASSWORD"),
		DBHost:        os.Getenv("POSTGREST_DB_HOST"),
		DBSchema:      os.Getenv("POSTGREST_DEFAULT_SCHEMA"),
		DBRole:        os.Getenv("POSTGREST_DEFAULT_ROLE"),
		CustomOrigins: corsOrigins,
		BinaryPath:    binaryPath,
	}

	projectRepo := do.MustInvoke[project.Repository](injector)

	return &ServiceImpl{
		projectRepo: projectRepo,
		config:      config,
		processes:   make(map[string]*process),
	}, nil
}

func (s *ServiceImpl) Start(dbName, jwtSecret string) {
	s.mutex.Lock()
	if _, exists := s.processes[dbName]; exists {
		s.mutex.Unlock()

		return
	}

	port, err := allocateFreePort()
	if err != nil {
		s.mutex.Unlock()
		s.markStartFailed(dbName, err)

		return
	}

	cmd := exec.Command(s.config.BinaryPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGRST_DB_URI=postgres://%s:%s@%s/%s", s.config.DBUser, s.config.DBPassword, s.config.DBHost, dbName),
		"PGRST_DB_ANON_ROLE="+s.config.DBRole,
		"PGRST_DB_SCHEMA="+s.config.DBSchema,
		"PGRST_JWT_SECRET="+jwtSecret,
		"PGRST_SERVER_HOST=127.0.0.1",
		fmt.Sprintf("PGRST_SERVER_PORT=%d", port),
		"PGRST_SERVER_CORS_ALLOWED_ORIGINS="+s.config.CustomOrigins,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		s.mutex.Unlock()
		s.markStartFailed(dbName, err)

		return
	}

	proc := &process{cmd: cmd, port: port}
	s.processes[dbName] = proc
	s.mutex.Unlock()

	// Write the active status before supervise can run: if the process dies
	// instantly, supervise's error status must land after this write, not race it
	_, err = s.projectRepo.UpdateStatusByDatabaseName(dbName, constants.ProjectStatusActive)
	if err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to update project status to active")
	}

	go s.supervise(dbName, proc)

	log.Info().
		Str("action", constants.ActionPostgrest).
		Str("db", dbName).
		Int("port", port).
		Msg("process started successfully")
}

func (s *ServiceImpl) StartAll() {
	projects, err := s.projectRepo.List(shared.PaginationParams{Page: 1, Limit: 1000})
	if err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("error", err.Error()).
			Msg("failed to list projects for startup")

		return
	}

	for _, currentProject := range projects {
		if currentProject.DBName == "" {
			continue
		}

		s.Start(currentProject.DBName, currentProject.JWTSecret)
	}
}

func (s *ServiceImpl) Stop(dbName string) {
	s.mutex.Lock()
	proc, exists := s.processes[dbName]
	if exists {
		proc.stopping = true
	}
	s.mutex.Unlock()

	if exists {
		if err := proc.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Error().
				Str("action", constants.ActionPostgrest).
				Str("db", dbName).
				Str("error", err.Error()).
				Msg("failed to stop process")
		}
	}

	_, err := s.projectRepo.UpdateStatusByDatabaseName(dbName, constants.ProjectStatusInactive)
	if err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to update project status to inactive")
	}
}

func (s *ServiceImpl) IsRunning(dbName string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.processes[dbName]

	return exists
}

func (s *ServiceImpl) Endpoint(dbName string) (string, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	proc, exists := s.processes[dbName]
	if !exists {
		return "", false
	}

	return fmt.Sprintf("http://127.0.0.1:%d", proc.port), true
}

func (s *ServiceImpl) RefreshSchemaCache(dbName string) {
	// Why we need this: https://postgrest.org/en/v10/schema_cache.html
	s.mutex.Lock()
	proc, exists := s.processes[dbName]
	s.mutex.Unlock()

	if !exists {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Msg("cannot refresh schema cache, process not running")

		return
	}

	if err := proc.cmd.Process.Signal(syscall.SIGUSR1); err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to refresh schema cache")
	}
}

func (s *ServiceImpl) supervise(dbName string, proc *process) {
	err := proc.cmd.Wait()

	s.mutex.Lock()
	stopping := proc.stopping
	delete(s.processes, dbName)
	s.mutex.Unlock()

	if stopping {
		return
	}

	logEvent := log.Error().
		Str("action", constants.ActionPostgrest).
		Str("db", dbName)
	if err != nil {
		logEvent = logEvent.Str("error", err.Error())
	}
	logEvent.Msg("process exited unexpectedly")

	if _, err := s.projectRepo.UpdateStatusByDatabaseName(dbName, constants.ProjectStatusError); err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to update project status to error")
	}
}

func (s *ServiceImpl) markStartFailed(dbName string, startErr error) {
	log.Error().
		Str("action", constants.ActionPostgrest).
		Str("db", dbName).
		Str("error", startErr.Error()).
		Msg("failed to start process")

	if _, err := s.projectRepo.UpdateStatusByDatabaseName(dbName, constants.ProjectStatusError); err != nil {
		log.Error().
			Str("action", constants.ActionPostgrest).
			Str("db", dbName).
			Str("error", err.Error()).
			Msg("failed to update project status to error")
	}
}

func allocateFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}
