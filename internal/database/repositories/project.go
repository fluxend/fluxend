package repositories

import (
	"database/sql"
	"errors"
	"fluxend/internal/domain/project"
	"fluxend/internal/domain/shared"
	"fluxend/pkg"
	flxErrs "fluxend/pkg/errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/samber/do"
)

type ProjectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(injector *do.Injector) (project.Repository, error) {
	db := do.MustInvoke[*sqlx.DB](injector)

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

	rows, err := r.db.NamedQuery(query, params)
	if err != nil {
		return nil, pkg.FormatError(err, "select", pkg.GetMethodName())
	}
	defer rows.Close()

	var projects []project.Project
	for rows.Next() {
		var organization project.Project
		if err := rows.StructScan(&organization); err != nil {
			return nil, pkg.FormatError(err, "scan", pkg.GetMethodName())
		}
		projects = append(projects, organization)
	}

	if err := rows.Err(); err != nil {
		return nil, pkg.FormatError(err, "iterate", pkg.GetMethodName())
	}

	return projects, nil
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

	rows, err := r.db.NamedQuery(query, params)
	if err != nil {
		return nil, pkg.FormatError(err, "select", pkg.GetMethodName())
	}
	defer rows.Close()

	var projects []project.Project
	for rows.Next() {
		var fetchedProject project.Project
		if err := rows.StructScan(&fetchedProject); err != nil {
			return nil, pkg.FormatError(err, "scan", pkg.GetMethodName())
		}
		projects = append(projects, fetchedProject)
	}

	if err := rows.Err(); err != nil {
		return nil, pkg.FormatError(err, "iterate", pkg.GetMethodName())
	}

	return projects, nil
}

func (r *ProjectRepository) GetByUUID(projectUUID uuid.UUID) (project.Project, error) {
	query := "SELECT %s FROM fluxend.projects WHERE uuid = $1"
	query = fmt.Sprintf(query, pkg.GetColumns[project.Project]())

	var fetchedProject project.Project
	err := r.db.Get(&fetchedProject, query, projectUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return project.Project{}, flxErrs.NewNotFoundError("project.error.notFound")
		}

		return project.Project{}, pkg.FormatError(err, "fetch", pkg.GetMethodName())
	}

	return fetchedProject, nil
}

func (r *ProjectRepository) GetDatabaseNameByUUID(projectUUID uuid.UUID) (string, error) {
	query := "SELECT db_name FROM fluxend.projects WHERE uuid = $1"

	var dbName string
	err := r.db.Get(&dbName, query, projectUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", flxErrs.NewNotFoundError("project.error.notFound")
		}

		return "", pkg.FormatError(err, "fetch", pkg.GetMethodName())
	}

	return dbName, nil
}

func (r *ProjectRepository) GetUUIDByDatabaseName(dbName string) (uuid.UUID, error) {
	query := "SELECT uuid FROM fluxend.projects WHERE db_name = $1"

	var projectUUID uuid.UUID
	err := r.db.Get(&projectUUID, query, dbName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.UUID{}, flxErrs.NewNotFoundError("project.error.notFound")
		}

		return uuid.UUID{}, pkg.FormatError(err, "fetch", pkg.GetMethodName())
	}

	return projectUUID, nil
}

func (r *ProjectRepository) GetOrganizationUUIDByProjectUUID(id uuid.UUID) (uuid.UUID, error) {
	query := "SELECT organization_uuid FROM fluxend.projects WHERE uuid = $1"

	var organizationUUID uuid.UUID
	err := r.db.Get(&organizationUUID, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.UUID{}, flxErrs.NewNotFoundError("project.error.notFound")
		}

		return uuid.UUID{}, pkg.FormatError(err, "fetch", pkg.GetMethodName())
	}

	return organizationUUID, nil
}

func (r *ProjectRepository) ExistsByUUID(id uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM fluxend.projects WHERE uuid = $1)"

	var exists bool
	err := r.db.Get(&exists, query, id)
	if err != nil {
		return false, pkg.FormatError(err, "fetch", pkg.GetMethodName())
	}

	return exists, nil
}

func (r *ProjectRepository) ExistsByNameForOrganization(name string, organizationUUID uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM fluxend.projects WHERE name = $1 AND organization_uuid = $2)"

	var exists bool
	err := r.db.Get(&exists, query, name, organizationUUID)
	if err != nil {
		return false, pkg.FormatError(err, "fetch", pkg.GetMethodName())
	}

	return exists, nil
}

func (r *ProjectRepository) Create(project *project.Project) (*project.Project, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, pkg.FormatError(err, "transactionBegin", pkg.GetMethodName())
	}

	query := `
		INSERT INTO fluxend.projects (
			name, db_name, description, db_port, 
			organization_uuid, created_by, updated_by
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING uuid
	`

	queryErr := tx.QueryRowx(
		query,
		project.Name,
		project.DBName,
		project.Description,
		project.DBPort,
		project.OrganizationUuid,
		project.CreatedBy,
		project.UpdatedBy,
	).Scan(&project.Uuid)

	if queryErr != nil {
		err := tx.Rollback()
		if err != nil {
			return nil, err
		}

		return nil, pkg.FormatError(queryErr, "insert", pkg.GetMethodName())
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, pkg.FormatError(err, "transactionCommit", pkg.GetMethodName())
	}

	return project, nil
}

func (r *ProjectRepository) Update(projectInput *project.Project) (*project.Project, error) {
	query := `
		UPDATE fluxend.projects 
		SET name = :name, description = :description, updated_at = :updated_at, updated_by = :updated_by
		WHERE uuid = :uuid`

	res, err := r.db.NamedExec(query, projectInput)
	if err != nil {
		return &project.Project{}, pkg.FormatError(err, "update", pkg.GetMethodName())
	}

	_, err = res.RowsAffected()
	if err != nil {
		return &project.Project{}, pkg.FormatError(err, "affectedRows", pkg.GetMethodName())
	}

	return projectInput, nil
}

func (r *ProjectRepository) UpdateStatusByDatabaseName(databaseName, status string) (bool, error) {
	query := "UPDATE fluxend.projects SET status = $1 WHERE db_name = $2"
	res, err := r.db.Exec(query, status, databaseName)
	if err != nil {
		return false, pkg.FormatError(err, "update", pkg.GetMethodName())
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, pkg.FormatError(err, "affectedRows", pkg.GetMethodName())
	}

	return rowsAffected == 1, nil
}

func (r *ProjectRepository) Delete(projectUUID uuid.UUID) (bool, error) {
	query := "DELETE FROM fluxend.projects WHERE uuid = $1"
	res, err := r.db.Exec(query, projectUUID)
	if err != nil {
		return false, pkg.FormatError(err, "delete", pkg.GetMethodName())
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, pkg.FormatError(err, "affectedRows", pkg.GetMethodName())
	}

	return rowsAffected == 1, nil
}
