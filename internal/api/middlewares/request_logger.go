package middlewares

import (
	"bytes"
	"encoding/json"
	"fluxend/internal/config/constants"
	"fluxend/internal/domain/logging"
	"fluxend/pkg/auth"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var skippableEndpointPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^/projects/[a-f0-9-]{36}/logs$`),
	// PostgREST proxy requests are logged separately by the proxy handler
	regexp.MustCompile(`^/rest/`),
}

func RequestLogger(requestLogRepo logging.Repository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip logging for certain endpoints to avoid cluttering the logs
			for _, pattern := range skippableEndpointPatterns {
				if pattern.MatchString(c.Request().URL.Path) {
					return next(c)
				}
			}

			request := c.Request()
			requestBody := readBody(request)

			res := next(c)
			authUserUUID, _ := auth.NewAuth(c).Uuid()

			logEntry := logging.RequestLog{
				ProjectUuid: extractProjectUUID(c),
				UserUuid:    authUserUUID,
				APIKey:      uuid.MustParse("00000000-0000-0000-0000-000000000000"), // TODO: implement API key parsing/reading later
				Method:      request.Method,
				Status:      c.Response().Status,
				Endpoint:    request.URL.Path,
				IPAddress:   c.RealIP(),
				UserAgent:   request.UserAgent(),
				Params:      request.URL.RawQuery,
				Body:        sanitizeRequestBody(requestBody),
				CreatedAt:   time.Now(),
			}

			log.Info().
				Str("action", constants.ActionAPIRequest).
				Str("user_uuid", authUserUUID.String()).
				Str("method", logEntry.Method).
				Str("endpoint", logEntry.Endpoint).
				Str("ip_address", logEntry.IPAddress).
				Str("user_agent", logEntry.UserAgent).
				Int("status", c.Response().Status).
				Msg("")

			go requestLogRepo.Create(&logEntry)

			return res
		}
	}
}

// extractProjectUUID attempts to extract project UUID from either X-Project header or URL path
func extractProjectUUID(c echo.Context) uuid.UUID {
	// First, try to get project UUID from X-Project header
	projectHeader := c.Request().Header.Get("X-Project")
	if projectHeader != "" {
		if projectUUID, err := uuid.Parse(projectHeader); err == nil {
			return projectUUID
		}
		log.Warn().
			Str("header_value", projectHeader).
			Msg("Invalid UUID format in X-Project header")
	}

	// If not found in header, try to extract from URL path
	// Look for /projects/:projectUUID pattern
	path := c.Request().URL.Path

	// Split path by "/" and look for "projects" segment followed by UUID
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if segment == "projects" && i+1 < len(segments) {
			if projectUUID, err := uuid.Parse(segments[i+1]); err == nil {
				return projectUUID
			}
			log.Warn().
				Str("path_segment", segments[i+1]).
				Msg("Invalid UUID format in URL path")
			break
		}
	}

	return uuid.Nil
}

func readBody(r *http.Request) string {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")

		return ""
	}

	// Restore the io.ReadCloser to its original state
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return string(body)
}

// make sure request body doesn't contain sensitive information
func sanitizeRequestBody(body string) string {
	var generic map[string]interface{}
	if err := json.Unmarshal([]byte(body), &generic); err != nil {
		return body
	}

	// If it has a "password" field, just mask it
	if _, exists := generic["password"]; exists {
		generic["password"] = "***"
		sanitizedBody, _ := json.Marshal(generic)

		return string(sanitizedBody)
	}

	return body
}
