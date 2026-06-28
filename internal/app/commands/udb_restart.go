package commands

import (
	"fluxend/internal/app"
	"fluxend/internal/domain/project"
	"fluxend/internal/domain/shared"
	"fmt"
	"github.com/samber/do"
	"github.com/spf13/cobra"
)

var udbRestart = &cobra.Command{
	Use:   "udb.restart",
	Short: "Restart all PostGREST instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		return restartPostgrestInstances()
	},
}

func restartPostgrestInstances() error {
	container := app.InitializeContainer()

	// Inject dependencies
	projectRepository := do.MustInvoke[project.Repository](container)
	postgrestService := do.MustInvoke[shared.PostgrestService](container)

	projects, err := projectRepository.List(shared.PaginationParams{Page: 1, Limit: 1000})
	if err != nil {
		return fmt.Errorf("error fetching projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found")

		return nil
	}

	fmt.Printf("Found %d projects\n", len(projects))

	// Restart PostGREST instances for each project
	for i, currentProject := range projects {
		if currentProject.DBName == "" {
			continue
		}

		// Log restart process for each project
		fmt.Printf("Restarting PostGREST instance for project %s (%d/%d)\n", currentProject.DBName, i+1, len(projects))

		hasContainer := postgrestService.HasContainer(currentProject.DBName)
		if hasContainer {
			postgrestService.RemoveContainer(currentProject.DBName)
		}

		postgrestService.StartContainer(currentProject.DBName, currentProject.JWTSecret)
	}

	return nil
}
