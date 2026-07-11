package api

import (
	"encoding/json"
	userDto "fluxend/internal/api/dto/user"
	"fluxend/pkg"
	"fmt"
	"net/http"
	"testing"

	"fluxend/tests/integration"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userResponse struct {
	integration.APIResponse
	Content struct {
		User  userDto.Response `json:"user"`
		Token string           `json:"token,omitempty"`
	}
}

type userProfileResponse struct {
	integration.APIResponse
	Content struct {
		userDto.Response
	}
}

func TestUserRegistration_Suite(t *testing.T) {
	server := integration.NewTestServer()
	defer server.Close()

	t.Run("successful registration", func(t *testing.T) {
		userInput := getFakeUserData()

		resp := server.PostJSON(t, "/users/register", userInput)
		defer resp.Body.Close()

		// Assert response status
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Parse response
		var response userResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))

		// Assert response structure
		assert.True(t, response.Success)
		assert.NotNil(t, response.Content)
		assert.NotEmpty(t, response.Content.Token)

		assert.NotNil(t, response.Content.User)

		createdUser := response.Content.User
		userUUID, err := uuid.Parse(createdUser.Uuid.String())
		require.NoError(t, err)

		// Assert user data
		assert.Equal(t, userInput.Username, createdUser.Username)
		assert.Equal(t, userInput.Email, createdUser.Email)
		assert.Equal(t, userInput.Bio, createdUser.Bio)

		// Verify user exists in database
		var dbUser struct {
			Uuid     uuid.UUID `db:"uuid"`
			Username string    `db:"username"`
			Email    string    `db:"email"`
			Bio      string    `db:"bio"`
		}
		err = server.DB.Get(&dbUser, "SELECT uuid, username, email, bio FROM authentication.users WHERE uuid = $1", userUUID)
		require.NoError(t, err)
		assert.Equal(t, userInput.Username, dbUser.Username)
		assert.Equal(t, userInput.Email, dbUser.Email)

		// Register cleanup
		server.AddCleanup(func() error {
			return server.CleanupUser(userUUID)
		})
	})

	t.Run("duplicate email registration fails", func(t *testing.T) {
		// First registration
		userInputA := getFakeUserData()

		responseA := server.PostJSON(t, "/users/register", userInputA)
		defer responseA.Body.Close()
		assert.Equal(t, http.StatusCreated, responseA.StatusCode)

		var userResponseA userResponse
		require.NoError(t, json.NewDecoder(responseA.Body).Decode(&userResponseA))
		createdUserA := userResponseA.Content.User
		userAUUID, err := uuid.Parse(createdUserA.Uuid.String())
		require.NoError(t, err)

		// Second registration with same email
		userInputB := userInputA
		userInputB.Username = "user_" + pkg.Faker.RandomStringWithLength(10)

		responseB := server.PostJSON(t, "/users/register", userInputB)
		defer responseB.Body.Close()
		assert.Equal(t, http.StatusBadRequest, responseB.StatusCode)

		var userResponseB userResponse
		require.NoError(t, json.NewDecoder(responseB.Body).Decode(&userResponseB))
		assert.False(t, userResponseB.Success)
		assert.Contains(t, userResponseB.Errors, "User with this email already exists")

		// Cleanup
		server.AddCleanup(func() error {
			return server.CleanupUser(userAUUID)
		})
	})

	t.Run("duplicate username registration fails", func(t *testing.T) {
		// First registration
		userInputA := getFakeUserData()
		fmt.Println(userInputA)

		responseA := server.PostJSON(t, "/users/register", userInputA)
		defer responseA.Body.Close()
		assert.Equal(t, http.StatusCreated, responseA.StatusCode)

		var userResponseA userResponse
		require.NoError(t, json.NewDecoder(responseA.Body).Decode(&userResponseA))
		createdUserA := userResponseA.Content.User
		userAUUID, err := uuid.Parse(createdUserA.Uuid.String())
		require.NoError(t, err)

		// Second registration with same username
		userInputB := userInputA
		userInputB.Email = pkg.Faker.RandomStringWithLength(10) + "@gmail.com"

		responseB := server.PostJSON(t, "/users/register", userInputB)
		defer responseB.Body.Close()
		assert.Equal(t, http.StatusBadRequest, responseB.StatusCode)

		var userResponseB userResponse
		require.NoError(t, json.NewDecoder(responseB.Body).Decode(&userResponseB))
		assert.False(t, userResponseB.Success)
		assert.Contains(t, userResponseB.Errors, "User with this username already exists")

		// Cleanup
		server.AddCleanup(func() error {
			return server.CleanupUser(userAUUID)
		})
	})
}

func TestUserLogin_Suite(t *testing.T) {
	server := integration.NewTestServer()
	defer server.Close()

	username := "jonsnow"
	email := "snow@thewall.com"
	password := "winter_is_coming"

	// Create a user for login tests
	userInput := map[string]interface{}{
		"username": username,
		"email":    email,
		"password": password,
	}

	resp := server.PostJSON(t, "/users/register", userInput)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdUserResponse userResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&createdUserResponse))
	createdUser := createdUserResponse.Content.User
	userUUID, _ := uuid.Parse(createdUser.Uuid.String())

	// Register cleanup
	server.AddCleanup(func() error {
		return server.CleanupUser(userUUID)
	})

	t.Run("successful login", func(t *testing.T) {
		loginData := map[string]string{
			"email":    email,
			"password": password,
		}

		loginResponse := server.PostJSON(t, "/users/login", loginData)
		defer loginResponse.Body.Close()

		assert.Equal(t, http.StatusOK, loginResponse.StatusCode)

		var response userResponse
		require.NoError(t, json.NewDecoder(loginResponse.Body).Decode(&response))

		assert.True(t, response.Success)
		assert.NotEmpty(t, response.Content.User)
		assert.NotEmpty(t, response.Content.Token)

		loggedInUser := response.Content.User
		assert.NotNil(t, loggedInUser.Uuid)

		assert.Equal(t, username, loggedInUser.Username)
		assert.Equal(t, email, loggedInUser.Email)
	})

	t.Run("login with wrong password fails", func(t *testing.T) {
		loginData := map[string]string{
			"email":    email,
			"password": "wrongpassword",
		}

		logInTryResponse := server.PostJSON(t, "/users/login", loginData)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, logInTryResponse.StatusCode)
	})

	t.Run("login with non-existent email fails", func(t *testing.T) {
		loginData := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "password123",
		}

		resp := server.PostJSON(t, "/users/login", loginData)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestUserProfile_Suite(t *testing.T) {
	server := integration.NewTestServer()
	defer server.Close()

	// Create and login user
	userInput := getFakeUserData()

	resp := server.PostJSON(t, "/users/register", userInput)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdUserResponse userResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&createdUserResponse))

	token := createdUserResponse.Content.Token
	createdUser := createdUserResponse.Content.User
	userUUID, _ := uuid.Parse(createdUser.Uuid.String())

	// Register cleanup
	server.AddCleanup(func() error {
		return server.CleanupUser(userUUID)
	})

	t.Run("get user profile with valid token", func(t *testing.T) {
		validProfileWithTokenResponse := server.GetWithAuth(t, fmt.Sprintf("/users/%s", userUUID), token)
		defer validProfileWithTokenResponse.Body.Close()

		assert.Equal(t, http.StatusOK, validProfileWithTokenResponse.StatusCode)

		var currentUser userProfileResponse
		require.NoError(t, json.NewDecoder(validProfileWithTokenResponse.Body).Decode(&currentUser))

		assert.True(t, currentUser.Success)

		userProfile := currentUser.Content
		assert.Equal(t, userInput.Username, userProfile.Username)
		assert.Equal(t, userInput.Email, userProfile.Email)
		assert.Equal(t, userInput.Bio, userProfile.Bio)
	})

	t.Run("get user profile without token fails", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.BaseURL+fmt.Sprintf("/users/%s", userUUID), nil)
		require.NoError(t, err)

		profileWithoutTokenResponse, err := server.Client.Do(req)
		require.NoError(t, err)

		defer profileWithoutTokenResponse.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, profileWithoutTokenResponse.StatusCode)
	})

	t.Run("update user profile", func(t *testing.T) {
		updateData := map[string]string{
			"bio": "Updated bio for profile user",
		}

		validProfileUpdateResponse := server.PutJSONWithAuth(t, fmt.Sprintf("/users/%s", userUUID), token, updateData)
		defer validProfileUpdateResponse.Body.Close()

		assert.Equal(t, http.StatusOK, validProfileUpdateResponse.StatusCode)

		var currentUser userProfileResponse
		require.NoError(t, json.NewDecoder(validProfileUpdateResponse.Body).Decode(&currentUser))

		assert.True(t, currentUser.Success)

		updatedUser := currentUser.Content
		assert.Equal(t, "Updated bio for profile user", updatedUser.Bio)

		// Verify in database
		var dbBio string
		err := server.DB.Get(&dbBio, "SELECT bio FROM authentication.users WHERE uuid = $1", userUUID)
		require.NoError(t, err)
		assert.Equal(t, "Updated bio for profile user", dbBio)
	})

	t.Run("update user profile without token fails", func(t *testing.T) {
		updateData := map[string]string{
			"bio": "Should not update",
		}

		forbiddenProfileResponse := server.PutJSON(t, fmt.Sprintf("/users/%s", userUUID), updateData)
		defer forbiddenProfileResponse.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, forbiddenProfileResponse.StatusCode)
	})
}

func TestUserProfileAccess_Suite(t *testing.T) {
	server := integration.NewTestServer()
	defer server.Close()

	// Create two users
	userInputA := map[string]string{
		"username": "user1",
		"email":    "user1@example.com",
		"password": "password123",
	}

	userInputB := map[string]string{
		"username": "user2",
		"email":    "user2@example.com",
		"password": "password123",
	}

	// Register first user
	responseA := server.PostJSON(t, "/users/register", userInputA)
	defer responseA.Body.Close()
	require.Equal(t, http.StatusCreated, responseA.StatusCode)

	var userResponseA userResponse
	require.NoError(t, json.NewDecoder(responseA.Body).Decode(&userResponseA))
	userUUIDA, _ := uuid.Parse(userResponseA.Content.User.Uuid.String())
	userTokenA := userResponseA.Content.Token

	// Register second user
	responseB := server.PostJSON(t, "/users/register", userInputB)
	defer responseB.Body.Close()
	require.Equal(t, http.StatusCreated, responseB.StatusCode)

	var userResponseB userResponse
	require.NoError(t, json.NewDecoder(responseB.Body).Decode(&userResponseB))
	userUUIDB, _ := uuid.Parse(userResponseB.Content.User.Uuid.String())

	// Register cleanup
	server.AddCleanup(func() error {
		server.CleanupUser(userUUIDA)
		return server.CleanupUser(userUUIDB)
	})

	t.Run("user cannot update other user profile", func(t *testing.T) {
		updateData := map[string]string{
			"bio": "Trying to update someone else's profile",
		}

		resp := server.PutJSONWithAuth(t, fmt.Sprintf("/users/%s", userUUIDB), userTokenA, updateData)
		defer resp.Body.Close()

		// Should fail due to policy check
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func getFakeUserData() userDto.CreateRequest {
	return userDto.CreateRequest{
		Username: "user_" + pkg.Faker.RandomStringWithLength(10),
		Email:    pkg.Faker.RandomStringWithLength(10) + "@gmail.com",
		Password: pkg.Faker.Internet().Password(),
		Bio:      pkg.Faker.RandomStringWithLength(20),
	}
}
