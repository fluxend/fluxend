package project

import (
	"errors"
	"fluxend/internal/config/constants"
	"fluxend/internal/domain/auth"
	"fluxend/internal/domain/shared"
	"fluxend/tests/fixtures/mocks/organization"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

// stubProjectRepo is a minimal testify mock covering only what GetProjectToken calls.
// Using an inline mock here avoids the import cycle that would arise from importing
// tests/fixtures/mocks/project (which itself imports this package).
type stubProjectRepo struct {
	mock.Mock
}

func (m *stubProjectRepo) GetByUUID(projectUUID uuid.UUID) (Project, error) {
	args := m.Called(projectUUID)
	return args.Get(0).(Project), args.Error(1)
}

func (m *stubProjectRepo) ListForUser(p shared.PaginationParams, userID uuid.UUID) ([]Project, error) {
	panic("not expected")
}
func (m *stubProjectRepo) List(p shared.PaginationParams) ([]Project, error) { panic("not expected") }
func (m *stubProjectRepo) GetDatabaseNameByUUID(id uuid.UUID) (string, error) {
	panic("not expected")
}
func (m *stubProjectRepo) GetUUIDByDatabaseName(dbName string) (uuid.UUID, error) {
	panic("not expected")
}
func (m *stubProjectRepo) GetOrganizationUUIDByProjectUUID(id uuid.UUID) (uuid.UUID, error) {
	panic("not expected")
}
func (m *stubProjectRepo) ExistsByUUID(id uuid.UUID) (bool, error)      { panic("not expected") }
func (m *stubProjectRepo) ExistsByNameForOrganization(name string, orgUUID uuid.UUID) (bool, error) {
	panic("not expected")
}
func (m *stubProjectRepo) Create(p *Project) (*Project, error)                { panic("not expected") }
func (m *stubProjectRepo) Update(p *Project) (*Project, error)                { panic("not expected") }
func (m *stubProjectRepo) UpdateStatusByDatabaseName(db, status string) (bool, error) {
	panic("not expected")
}
func (m *stubProjectRepo) Delete(id uuid.UUID) (bool, error) { panic("not expected") }

func buildServiceForTokenTests(t *testing.T) (*ServiceImpl, *stubProjectRepo, *organization.MockRepository) {
	projectRepo := &stubProjectRepo{}
	orgRepo := organization.NewMockRepository(t)
	policy := &Policy{organizationRepo: orgRepo}

	svc := &ServiceImpl{
		projectPolicy: policy,
		projectRepo:   projectRepo,
	}

	return svc, projectRepo, orgRepo
}

func TestGetProjectToken_Suite(t *testing.T) {
	t.Run("returns error when project not found", func(t *testing.T) {
		svc, projectRepo, _ := buildServiceForTokenTests(t)

		projectUUID := uuid.New()
		authUser := auth.User{Uuid: uuid.New(), RoleID: constants.UserRoleDeveloper}

		projectRepo.On("GetByUUID", projectUUID).Return(Project{}, errors.New("project.error.notFound"))

		_, err := svc.GetProjectToken(projectUUID, authUser)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "notFound")
	})

	t.Run("returns forbidden when user is not an org member", func(t *testing.T) {
		svc, projectRepo, orgRepo := buildServiceForTokenTests(t)

		orgUUID := uuid.New()
		projectUUID := uuid.New()
		authUser := auth.User{Uuid: uuid.New(), RoleID: constants.UserRoleDeveloper}

		projectRepo.On("GetByUUID", projectUUID).Return(Project{
			Uuid:             projectUUID,
			OrganizationUuid: orgUUID,
			JWTSecret:        "somesecret",
		}, nil)
		orgRepo.On("IsOrganizationMember", orgUUID, authUser.Uuid).Return(false, nil)

		_, err := svc.GetProjectToken(projectUUID, authUser)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "viewForbidden")
	})

	t.Run("returned token is verifiable with the project's own secret", func(t *testing.T) {
		svc, projectRepo, orgRepo := buildServiceForTokenTests(t)

		orgUUID := uuid.New()
		projectUUID := uuid.New()
		userUUID := uuid.New()
		authUser := auth.User{Uuid: userUUID, RoleID: constants.UserRoleDeveloper}
		projectSecret := "a-distinct-per-project-secret-that-is-long-enough"

		projectRepo.On("GetByUUID", projectUUID).Return(Project{
			Uuid:             projectUUID,
			OrganizationUuid: orgUUID,
			JWTSecret:        projectSecret,
		}, nil)
		orgRepo.On("IsOrganizationMember", orgUUID, userUUID).Return(true, nil)

		tokenString, err := svc.GetProjectToken(projectUUID, authUser)

		require.NoError(t, err)
		require.NotEmpty(t, tokenString)

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(projectSecret), nil
		})
		require.NoError(t, err)
		assert.True(t, token.Valid)

		expectedRole := "usr_" + strings.ReplaceAll(userUUID.String(), "-", "_")
		assert.Equal(t, expectedRole, claims["role"])
	})

	t.Run("token does not verify against the global JWT_SECRET", func(t *testing.T) {
		svc, projectRepo, orgRepo := buildServiceForTokenTests(t)

		orgUUID := uuid.New()
		projectUUID := uuid.New()
		userUUID := uuid.New()
		authUser := auth.User{Uuid: userUUID, RoleID: constants.UserRoleDeveloper}

		projectRepo.On("GetByUUID", projectUUID).Return(Project{
			Uuid:             projectUUID,
			OrganizationUuid: orgUUID,
			JWTSecret:        "per-project-secret-distinct-from-global-one-xxxxxxxx",
		}, nil)
		orgRepo.On("IsOrganizationMember", orgUUID, userUUID).Return(true, nil)

		tokenString, err := svc.GetProjectToken(projectUUID, authUser)
		require.NoError(t, err)

		claims := jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("test_jwt_secret_key_that_is_long_enough_for_validation"), nil
		})
		assert.Error(t, err, "token signed with project secret must not verify against global JWT_SECRET")
	})

	t.Run("token from project A is rejected by project B's secret", func(t *testing.T) {
		svc, projectRepo, orgRepo := buildServiceForTokenTests(t)

		orgUUID := uuid.New()
		userUUID := uuid.New()
		authUser := auth.User{Uuid: userUUID, RoleID: constants.UserRoleDeveloper}

		projectAUUID := uuid.New()
		secretA := "secret-for-project-alpha-long-enough-aaaaaaaaaaaaa"
		projectBUUID := uuid.New()
		secretB := "secret-for-project-beta-long-enough-bbbbbbbbbbbbb"

		projectRepo.On("GetByUUID", projectAUUID).Return(Project{
			Uuid:             projectAUUID,
			OrganizationUuid: orgUUID,
			JWTSecret:        secretA,
		}, nil)
		projectRepo.On("GetByUUID", projectBUUID).Return(Project{
			Uuid:             projectBUUID,
			OrganizationUuid: orgUUID,
			JWTSecret:        secretB,
		}, nil)
		orgRepo.On("IsOrganizationMember", orgUUID, userUUID).Return(true, nil).Times(2)

		tokenA, err := svc.GetProjectToken(projectAUUID, authUser)
		require.NoError(t, err)

		tokenB, err := svc.GetProjectToken(projectBUUID, authUser)
		require.NoError(t, err)

		claims := jwt.MapClaims{}

		_, err = jwt.ParseWithClaims(tokenA, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretB), nil
		})
		assert.Error(t, err, "project A token must not verify against project B secret")

		_, err = jwt.ParseWithClaims(tokenB, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretA), nil
		})
		assert.Error(t, err, "project B token must not verify against project A secret")
	})
}
