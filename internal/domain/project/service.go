package project

import (
	"crypto/rand"
	"encoding/hex"
	"fluxend/internal/domain/auth"
	"fluxend/internal/domain/shared"
	"fluxend/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/samber/do"
	mathrand "math/rand"
	"strings"
	"time"
)

type Service interface {
	List(paginationParams shared.PaginationParams, organizationUUID uuid.UUID, authUser auth.User) ([]Project, error)
	GetByUUID(projectUUID uuid.UUID, authUser auth.User) (Project, error)
	GetDatabaseNameByUUID(projectUUID uuid.UUID, authUser auth.User) (string, error)
	GetProjectToken(projectUUID uuid.UUID, authUser auth.User) (string, error)
	Create(request *CreateProjectInput, authUser auth.User) (Project, error)
	Update(projectUUID uuid.UUID, authUser auth.User, request *UpdateProjectInput) (*Project, error)
	Delete(projectUUID uuid.UUID, authUser auth.User) (bool, error)
}

type ServiceImpl struct {
	projectPolicy    *Policy
	databaseRepo     shared.DatabaseService
	projectRepo      Repository
	postgrestService shared.PostgrestService
}

func NewProjectService(injector *do.Injector) (Service, error) {
	policy := do.MustInvoke[*Policy](injector)
	databaseRepo := do.MustInvoke[shared.DatabaseService](injector)
	projectRepo := do.MustInvoke[Repository](injector)
	postgrestService := do.MustInvoke[shared.PostgrestService](injector)

	return &ServiceImpl{
		projectPolicy:    policy,
		databaseRepo:     databaseRepo,
		projectRepo:      projectRepo,
		postgrestService: postgrestService,
	}, nil
}

func (s *ServiceImpl) List(paginationParams shared.PaginationParams, organizationUUID uuid.UUID, authUser auth.User) ([]Project, error) {
	if !s.projectPolicy.CanAccess(organizationUUID, authUser) {
		return []Project{}, errors.NewForbiddenError("project.error.listForbidden")
	}

	return s.projectRepo.ListForUser(paginationParams, authUser.Uuid)
}

func (s *ServiceImpl) GetByUUID(projectUUID uuid.UUID, authUser auth.User) (Project, error) {
	fetchedProject, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return Project{}, err
	}

	if !s.projectPolicy.CanAccess(fetchedProject.OrganizationUuid, authUser) {
		return Project{}, errors.NewForbiddenError("project.error.viewForbidden")
	}

	return fetchedProject, nil
}

func (s *ServiceImpl) GetDatabaseNameByUUID(projectUUID uuid.UUID, authUser auth.User) (string, error) {
	fetchedProject, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return "", err
	}

	if !s.projectPolicy.CanAccess(fetchedProject.OrganizationUuid, authUser) {
		return "", errors.NewForbiddenError("project.error.viewForbidden")
	}

	return fetchedProject.DBName, nil
}

func (s *ServiceImpl) Create(request *CreateProjectInput, authUser auth.User) (Project, error) {
	if !s.projectPolicy.CanCreate(request.OrganizationUUID, authUser) {
		return Project{}, errors.NewForbiddenError("project.error.createForbidden")
	}

	err := s.validateNameForDuplication(request.Name, request.OrganizationUUID)
	if err != nil {
		return Project{}, err
	}

	projectInput := Project{
		Name:             request.Name,
		Description:      request.Description,
		OrganizationUuid: request.OrganizationUUID,
		DBName:           s.generateDBName(),
		DBPort:           s.generateDBPort(),
		JWTSecret:        s.generateJWTSecret(),
		CreatedBy:        authUser.Uuid,
		UpdatedBy:        authUser.Uuid,
	}

	_, err = s.projectRepo.Create(&projectInput)
	if err != nil {
		return Project{}, err
	}

	err = s.databaseRepo.Create(projectInput.DBName, uuid.NullUUID{UUID: authUser.Uuid, Valid: true})
	if err != nil {
		// TODO: handle better
		s.projectRepo.Delete(projectInput.Uuid)

		return Project{}, err
	}

	go s.postgrestService.StartContainer(projectInput.DBName, projectInput.JWTSecret)

	return projectInput, nil
}

func (s *ServiceImpl) Update(projectUUID uuid.UUID, authUser auth.User, request *UpdateProjectInput) (*Project, error) {
	fetchedProject, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, err
	}

	if !s.projectPolicy.CanUpdate(fetchedProject.OrganizationUuid, authUser) {
		return &Project{}, errors.NewForbiddenError("project.error.updateForbidden")
	}

	if err = fetchedProject.PopulateModel(&fetchedProject, request); err != nil {
		return nil, err
	}

	fetchedProject.UpdatedAt = time.Now()
	fetchedProject.UpdatedBy = authUser.Uuid

	if err = s.validateNameForDuplication(request.Name, fetchedProject.OrganizationUuid); err != nil {
		return &Project{}, err
	}

	return s.projectRepo.Update(&fetchedProject)
}

func (s *ServiceImpl) Delete(projectUUID uuid.UUID, authUser auth.User) (bool, error) {
	fetchedProject, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return false, err
	}

	if !s.projectPolicy.CanUpdate(fetchedProject.OrganizationUuid, authUser) {
		return false, errors.NewForbiddenError("project.error.updateForbidden")
	}

	if err = s.databaseRepo.DropIfExists(fetchedProject.DBName); err != nil {
		return false, err
	}

	go s.postgrestService.RemoveContainer(fetchedProject.DBName)

	return s.projectRepo.Delete(projectUUID)
}

func (s *ServiceImpl) generateDBName() string {
	return "udb" + strings.ReplaceAll(strings.ToLower(uuid.New().String()), "-", "")
}

func (s *ServiceImpl) generateDBPort() int {
	return mathrand.Intn(65535-5000+1) + 5000
}

func (s *ServiceImpl) generateJWTSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *ServiceImpl) GetProjectToken(projectUUID uuid.UUID, authUser auth.User) (string, error) {
	fetchedProject, err := s.projectRepo.GetByUUID(projectUUID)
	if err != nil {
		return "", err
	}

	if !s.projectPolicy.CanAccess(fetchedProject.OrganizationUuid, authUser) {
		return "", errors.NewForbiddenError("project.error.viewForbidden")
	}

	role := "usr_" + strings.ReplaceAll(authUser.Uuid.String(), "-", "_")
	claims := jwt.MapClaims{
		"role": role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(fetchedProject.JWTSecret))
}

func (s *ServiceImpl) validateNameForDuplication(name string, organizationUUID uuid.UUID) error {
	exists, err := s.projectRepo.ExistsByNameForOrganization(name, organizationUUID)
	if err != nil {
		return err
	}

	if exists {
		return errors.NewUnprocessableError("project.error.duplicateName")
	}

	return nil
}
