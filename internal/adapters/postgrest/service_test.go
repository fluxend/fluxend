package postgrest

import (
	"fluxend/internal/domain/project"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type stubProjectRepo struct {
	project.Repository
	mutex      sync.Mutex
	lastStatus string
}

func (s *stubProjectRepo) UpdateStatusByDatabaseName(databaseName, status string) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.lastStatus = status

	return true, nil
}

func (s *stubProjectRepo) LastStatus() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.lastStatus
}

func newTestService(t *testing.T) (*ServiceImpl, *stubProjectRepo) {
	t.Helper()

	script := filepath.Join(t.TempDir(), "fake_postgrest.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nwhile true; do sleep 0.1; done\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	repo := &stubProjectRepo{}

	return &ServiceImpl{
		projectRepo: repo,
		config:      &Config{BinaryPath: script},
		processes:   make(map[string]*process),
	}, repo
}

func waitForStopped(service *ServiceImpl, dbName string) bool {
	deadline := time.Now().Add(3 * time.Second)
	for service.IsRunning(dbName) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	return !service.IsRunning(dbName)
}

func TestStartStopLifecycle(t *testing.T) {
	service, repo := newTestService(t)

	service.Start("udbtest", "secret")

	if !service.IsRunning("udbtest") {
		t.Fatal("expected process to be running after Start")
	}

	if repo.LastStatus() != "active" {
		t.Fatalf("expected project status active, got %q", repo.LastStatus())
	}

	endpoint, ok := service.Endpoint("udbtest")
	if !ok || endpoint == "" {
		t.Fatal("expected endpoint for running process")
	}

	service.Stop("udbtest")

	if !waitForStopped(service, "udbtest") {
		t.Fatal("expected process to stop after Stop")
	}

	if _, ok := service.Endpoint("udbtest"); ok {
		t.Fatal("expected no endpoint after Stop")
	}
}

func TestUnexpectedExitMarksProjectError(t *testing.T) {
	service, repo := newTestService(t)

	service.Start("udbcrash", "secret")

	service.mutex.Lock()
	proc := service.processes["udbcrash"]
	service.mutex.Unlock()

	if err := proc.cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}

	if !waitForStopped(service, "udbcrash") {
		t.Fatal("expected process to be reaped after crash")
	}

	if repo.LastStatus() != "error" {
		t.Fatalf("expected project status error, got %q", repo.LastStatus())
	}
}

func TestStartFailureMarksProjectError(t *testing.T) {
	service, repo := newTestService(t)
	service.config.BinaryPath = "/nonexistent/postgrest"

	service.Start("udbmissing", "secret")

	if service.IsRunning("udbmissing") {
		t.Fatal("expected process not to be running")
	}

	if repo.LastStatus() != "error" {
		t.Fatalf("expected project status error, got %q", repo.LastStatus())
	}
}
