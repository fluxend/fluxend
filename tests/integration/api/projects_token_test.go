package api

import (
	"encoding/json"
	"fmt"
	organizationDto "fluxend/internal/api/dto/organization"
	"fluxend/pkg"
	"fluxend/tests/integration"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type projectTokenResponse struct {
	integration.APIResponse
	Content struct {
		Token string `json:"token"`
	} `json:"content"`
}

type organizationResponse struct {
	integration.APIResponse
	Content organizationDto.Response `json:"content"`
}

// seedProjectWithSecret inserts a project row directly into the DB with a known jwt_secret,
// bypassing the Docker container provisioning that the API endpoint would trigger.
func seedProjectWithSecret(t *testing.T, server *integration.TestServer, orgUUID, userUUID uuid.UUID, jwtSecret string) uuid.UUID {
	t.Helper()

	var projectUUID uuid.UUID
	err := server.DB.QueryRow(`
		INSERT INTO fluxend.projects (name, db_name, description, db_port, organization_uuid, created_by, updated_by, jwt_secret)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING uuid`,
		pkg.Faker.RandomStringWithLength(10),
		"udb"+strings.ReplaceAll(uuid.New().String(), "-", ""),
		"test project",
		pkg.Faker.IntBetween(20000, 40000),
		orgUUID,
		userUUID,
		userUUID,
		jwtSecret,
	).Scan(&projectUUID)
	require.NoError(t, err)

	return projectUUID
}

func TestProjectToken_Suite(t *testing.T) {
	server := integration.NewTestServer()
	defer server.Close()

	// Register a user
	userInput := getFakeUserData()
	registerResp := server.PostJSON(t, "/users/register", userInput)
	defer registerResp.Body.Close()
	require.Equal(t, http.StatusCreated, registerResp.StatusCode)

	var registeredUser userResponse
	require.NoError(t, json.NewDecoder(registerResp.Body).Decode(&registeredUser))
	userUUID, _ := uuid.Parse(registeredUser.Content.User.Uuid.String())
	authToken := registeredUser.Content.Token

	// Create an organization
	orgInput := map[string]string{"name": "test-org-" + pkg.Faker.RandomStringWithLength(6)}
	orgResp := server.PostJSONWithAuth(t, "/organizations", authToken, orgInput)
	defer orgResp.Body.Close()
	require.Equal(t, http.StatusCreated, orgResp.StatusCode)

	var org organizationResponse
	require.NoError(t, json.NewDecoder(orgResp.Body).Decode(&org))
	orgUUID := org.Content.Uuid

	// Seed two projects with distinct known secrets
	secretA := "integration-test-secret-for-project-alpha-aaaaaaaa"
	secretB := "integration-test-secret-for-project-beta-bbbbbbbbb"
	projectAUUID := seedProjectWithSecret(t, server, orgUUID, userUUID, secretA)
	projectBUUID := seedProjectWithSecret(t, server, orgUUID, userUUID, secretB)

	server.AddCleanup(func() error {
		server.DB.Exec("DELETE FROM fluxend.projects WHERE uuid IN ($1, $2)", projectAUUID, projectBUUID)
		server.DB.Exec("DELETE FROM fluxend.organizations WHERE uuid = $1", orgUUID)
		return server.CleanupUser(userUUID)
	})

	t.Run("unauthenticated request returns 401", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.BaseURL+fmt.Sprintf("/projects/%s/token", projectAUUID), nil)
		require.NoError(t, err)

		resp, err := server.Client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("unknown project UUID returns 404", func(t *testing.T) {
		resp := server.GetWithAuth(t, fmt.Sprintf("/projects/%s/token", uuid.New()), authToken)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("member receives a valid signed token", func(t *testing.T) {
		resp := server.GetWithAuth(t, fmt.Sprintf("/projects/%s/token", projectAUUID), authToken)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tokenResp projectTokenResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&tokenResp))
		assert.True(t, tokenResp.Success)
		require.NotEmpty(t, tokenResp.Content.Token)

		// Token must parse and verify with the project's own secret
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenResp.Content.Token, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secretA), nil
		})
		require.NoError(t, err)
		assert.True(t, token.Valid)

		// Role claim must match the user's PostgREST role
		expectedRole := "usr_" + strings.ReplaceAll(userUUID.String(), "-", "_")
		assert.Equal(t, expectedRole, claims["role"])
	})

	t.Run("token is not valid against the global JWT_SECRET", func(t *testing.T) {
		resp := server.GetWithAuth(t, fmt.Sprintf("/projects/%s/token", projectAUUID), authToken)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var tokenResp projectTokenResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&tokenResp))

		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(tokenResp.Content.Token, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte("test_jwt_secret_key_that_is_long_enough_for_validation"), nil
		})
		assert.Error(t, err, "project token must not be verifiable with the global JWT_SECRET")
	})

	t.Run("project A token is rejected by project B's secret", func(t *testing.T) {
		respA := server.GetWithAuth(t, fmt.Sprintf("/projects/%s/token", projectAUUID), authToken)
		defer respA.Body.Close()
		require.Equal(t, http.StatusOK, respA.StatusCode)

		var tokenRespA projectTokenResponse
		require.NoError(t, json.NewDecoder(respA.Body).Decode(&tokenRespA))

		respB := server.GetWithAuth(t, fmt.Sprintf("/projects/%s/token", projectBUUID), authToken)
		defer respB.Body.Close()
		require.Equal(t, http.StatusOK, respB.StatusCode)

		var tokenRespB projectTokenResponse
		require.NoError(t, json.NewDecoder(respB.Body).Decode(&tokenRespB))

		claims := jwt.MapClaims{}

		_, err := jwt.ParseWithClaims(tokenRespA.Content.Token, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secretB), nil
		})
		assert.Error(t, err, "project A token must not verify against project B secret")

		_, err = jwt.ParseWithClaims(tokenRespB.Content.Token, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secretA), nil
		})
		assert.Error(t, err, "project B token must not verify against project A secret")
	})
}
