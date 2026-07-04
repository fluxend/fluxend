package repositories

import (
	"fluxend/internal/domain/project"
	"fluxend/internal/domain/shared"
	"fluxend/pkg"
	"fmt"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type ProjectRepository struct {
	db shared.DB
}

func NewProjectRepository(injector *do.Injector) (project.Repository, error) {
	db := do.MustInvoke[shared.DB](injector)
	return &ProjectRepository{db: db}, nil
}

func (r *ProjectRepository) ListForUser(paginationParams shared.PaginationParams, authUserId uuid.UUID) ([]project.Project, error) {
	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query := `
		SELECT 
			%s 
		FROM 
			fluxend.projects projects
		JOIN 
			fluxend.organization_members organization_members ON projects.organization_uuid = organization_members.organization_uuid
		WHERE 
			organization_members.user_uuid = :user_uuid
		ORDER BY 
			:sort DESC
		LIMIT 
			:limit 
		OFFSET 
			:offset;
	`

	query = fmt.Sprintf(query, pkg.GetColumnsWithAlias[project.Project]("projects"))

	params := map[string]interface{}{
		"user_uuid": authUserId,
		"sort":      paginationParams.Sort,
		"limit":     paginationParams.Limit,
		"offset":    offset,
	}

	var projects []project.Project
	return projects, r.db.SelectNamedList(&projects, query, params)
}

func (r *ProjectRepository) List(paginationParams shared.PaginationParams) ([]project.Project, error) {
	offset := (paginationParams.Page - 1) * paginationParams.Limit
	query := `SELECT %s FROM fluxend.projects ORDER BY :sort DESC LIMIT :limit OFFSET :offset;`

	query = fmt.Sprintf(query, pkg.GetColumns[project.Project]())

	params := map[string]interface{}{
		"sort":   paginationParams.Sort,
		"limit":  paginationParams.Limit,
		"offset": offset,
	}

	var projects []project.Project
	return projects, r.db.SelectNamedList(&projects, query, params)
}

func (r *ProjectRepository) GetByUUID(projectUUID uuid.UUID) (project.Project, error) {
	query := "SELECT %s FROM fluxend.projects WHERE uuid = $1"
	query = fmt.Sprintf(query, pkg.GetColumns[project.Project]())

	var fetchedProject project.Project
	return fetchedProject, r.db.GetWithNotFound(&fetchedProject, "project.error.notFound", query, projectUUID)
}

func (r *ProjectRepository) GetDatabaseNameByUUID(projectUUID uuid.UUID) (string, error) {
	query := "SELECT db_name FROM fluxend.projects WHERE uuid = $1"

	var dbName string
	return dbName, r.db.GetWithNotFound(&dbName, "project.error.notFound", query, projectUUID)
}

func (r *ProjectRepository) GetUUIDByDatabaseName(dbName string) (uuid.UUID, error) {
	query := "SELECT uuid FROM fluxend.projects WHERE db_name = $1"

	var projectUUID uuid.UUID
	return projectUUID, r.db.GetWithNotFound(&projectUUID, "project.error.notFound", query, dbName)
}

func (r *ProjectRepository) GetOrganizationUUIDByProjectUUID(id uuid.UUID) (uuid.UUID, error) {
	query := "SELECT organization_uuid FROM fluxend.projects WHERE uuid = $1"

	var organizationUUID uuid.UUID
	return organizationUUID, r.db.GetWithNotFound(&organizationUUID, "project.error.notFound", query, id)
}

func (r *ProjectRepository) ExistsByUUID(id uuid.UUID) (bool, error) {
	return r.db.Exists("fluxend.projects", "uuid = $1", id)
}

func (r *ProjectRepository) ExistsByNameForOrganization(name string, organizationUUID uuid.UUID) (bool, error) {
	return r.db.Exists("fluxend.projects", "name = $1 AND organization_uuid = $2", name, organizationUUID)
}

func (r *ProjectRepository) Create(project *project.Project) (*project.Project, error) {
	return project, r.db.WithTransaction(func(tx shared.Tx) error {
		query := `
			INSERT INTO fluxend.projects (
				name, db_name, description, db_port,
				organization_uuid, created_by, updated_by, jwt_secret
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING uuid
		`

		return tx.QueryRowx(
			query,
			project.Name,
			project.DBName,
			project.Description,
			project.DBPort,
			project.OrganizationUuid,
			project.CreatedBy,
			project.UpdatedBy,
			project.JWTSecret,
		).Scan(&project.Uuid)
	})
}

func (r *ProjectRepository) Update(projectInput *project.Project) (*project.Project, error) {
	query := `
		UPDATE fluxend.projects 
		SET name = :name, description = :description, updated_at = :updated_at, updated_by = :updated_by
		WHERE uuid = :uuid`

	err := r.db.ExecWithErr(query, projectInput)

	return projectInput, err
}

func (r *ProjectRepository) UpdateStatusByDatabaseName(databaseName, status string) (bool, error) {
	rowsAffected, err := r.db.ExecWithRowsAffected("UPDATE fluxend.projects SET status = $1 WHERE db_name = $2", status, databaseName)
	if err != nil {
		return false, err
	}
	return rowsAffected == 1, nil
}

func (r *ProjectRepository) Delete(projectUUID uuid.UUID) (bool, error) {
	rowsAffected, err := r.db.ExecWithRowsAffected("DELETE FROM fluxend.projects WHERE uuid = $1", projectUUID)
	if err != nil {
		return false, err
	}
	return rowsAffected == 1, nil
}
