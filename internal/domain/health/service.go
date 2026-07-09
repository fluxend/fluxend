package health

import (
	"bufio"
	"fluxend/internal/domain/admin"
	"fluxend/internal/domain/auth"
	"fluxend/internal/domain/project"
	"fluxend/internal/domain/shared"
	"fluxend/pkg"
	"fluxend/pkg/errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/samber/do"
)

const statusOk = "OK"
const statusError = "ERROR"

type Service interface {
	Pulse(authUser auth.User) (Health, error)
}

type ServiceImpl struct {
	adminPolicy      *admin.Policy
	projectRepo      project.Repository
	postgrestService shared.PostgrestService
}

func NewHealthService(injector *do.Injector) (Service, error) {
	policy := admin.NewAdminPolicy()
	projectRepo := do.MustInvoke[project.Repository](injector)
	postgrestService := do.MustInvoke[shared.PostgrestService](injector)

	return &ServiceImpl{
		adminPolicy:      policy,
		projectRepo:      projectRepo,
		postgrestService: postgrestService,
	}, nil
}

func (s *ServiceImpl) Pulse(authUser auth.User) (Health, error) {
	if !s.adminPolicy.CanAccess(authUser) {
		return Health{}, errors.NewForbiddenError("setting.error.listForbidden")
	}

	diskTotal, diskAvailable, diskUsed, err := s.getDiskStats("/")
	if err != nil {
		return Health{}, fmt.Errorf("failed to get disk stats: %w", err)
	}

	cpuUsage, err := s.getCPUUsage()
	if err != nil {
		return Health{}, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	response := Health{
		DatabaseStatus:  statusOk,
		AppStatus:       statusOk,
		PostgrestStatus: statusOk,

		DiskUsage:     pkg.FormatPercentage(diskUsed, diskTotal),
		DiskAvailable: pkg.FormatBytes(diskAvailable),
		DiskTotal:     pkg.FormatBytes(diskTotal),
		CPUUsage:      cpuUsage,
		CPUCores:      runtime.NumCPU(),
	}

	allProjects, err := s.projectRepo.List(shared.PaginationParams{Page: 1, Limit: 1000})
	if err != nil {
		response.DatabaseStatus = statusError

		log.Error().Msg("Failed to list projects: " + err.Error())
	} else {
		for _, currentProject := range allProjects {
			if !s.postgrestService.IsRunning(currentProject.DBName) {
				log.Error().Msgf("Expected PostgREST process for project %s is not running", currentProject.DBName)

				response.PostgrestStatus = statusError
				break
			}
		}
	}

	return response, nil
}

func (s *ServiceImpl) getDiskStats(path string) (total, available, used uint64, err error) {
	var stat syscall.Statfs_t
	if err = syscall.Statfs(path, &stat); err != nil {
		return 0, 0, 0, err
	}

	total = stat.Blocks * uint64(stat.Bsize)
	available = stat.Bavail * uint64(stat.Bsize)
	used = total - available

	return total, available, used, nil
}

func (s *ServiceImpl) getCPUStats() (cpuStats, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return cpuStats{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) >= 8 {
				stats := cpuStats{}
				stats.user, _ = strconv.ParseUint(fields[1], 10, 64)
				stats.nice, _ = strconv.ParseUint(fields[2], 10, 64)
				stats.system, _ = strconv.ParseUint(fields[3], 10, 64)
				stats.idle, _ = strconv.ParseUint(fields[4], 10, 64)
				stats.iowait, _ = strconv.ParseUint(fields[5], 10, 64)
				stats.irq, _ = strconv.ParseUint(fields[6], 10, 64)
				stats.softirq, _ = strconv.ParseUint(fields[7], 10, 64)
				if len(fields) > 8 {
					stats.steal, _ = strconv.ParseUint(fields[8], 10, 64)
				}
				return stats, nil
			}
		}
	}
	return cpuStats{}, fmt.Errorf("cpu stats not found")
}

func (s *ServiceImpl) getCPUUsage() (string, error) {
	// Get initial reading
	prev, err := s.getCPUStats()
	if err != nil {
		return "0%", err
	}

	// Wait a brief moment
	time.Sleep(100 * time.Millisecond)

	// Get second reading
	curr, err := s.getCPUStats()
	if err != nil {
		return "0%", err
	}

	prevIdle := prev.idle + prev.iowait
	currIdle := curr.idle + curr.iowait

	prevNonIdle := prev.user + prev.nice + prev.system + prev.irq + prev.softirq + prev.steal
	currNonIdle := curr.user + curr.nice + curr.system + curr.irq + curr.softirq + curr.steal

	prevTotal := prevIdle + prevNonIdle
	currTotal := currIdle + currNonIdle

	totalDiff := currTotal - prevTotal
	idleDiff := currIdle - prevIdle

	if totalDiff == 0 {
		return "0%", nil
	}

	cpuUsage := float64(totalDiff-idleDiff) / float64(totalDiff) * 100
	return fmt.Sprintf("%.1f%%", cpuUsage), nil
}
