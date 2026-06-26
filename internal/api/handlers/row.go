package handlers

import (
	"encoding/json"
	"fluxend/internal/api/dto"
	"fluxend/internal/api/response"
	"fluxend/internal/domain/project"
	"fluxend/internal/domain/webhook"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

type RowHandler struct {
	projectRepo    project.Repository
	webhookService webhook.Service
	httpClient     *http.Client
}

func NewRowHandler(injector *do.Injector) (*RowHandler, error) {
	projectRepo := do.MustInvoke[project.Repository](injector)
	webhookService := do.MustInvoke[webhook.Service](injector)

	return &RowHandler{
		projectRepo:    projectRepo,
		webhookService: webhookService,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Insert proxies a row INSERT to PostgREST and fires webhooks
//
// @Summary Insert row
// @Description Insert a row into a table via the Fluxend proxy
// @Tags Rows
//
// @Accept json
// @Produce json
//
// @Param Authorization header string true "Bearer Token"
// @Param X-Project header string true "Project UUID"
// @Param fullTableName path string true "Full table name (e.g. public.orders)"
//
// @Success 201 "Row created"
// @Failure 400 {object} response.BadRequestErrorResponse "Bad request response"
// @Failure 401 {object} response.UnauthorizedErrorResponse "Unauthorized response"
// @Failure 500 {object} response.InternalServerErrorResponse "Internal server error response"
//
// @Router /tables/{fullTableName}/rows [post]
func (rh *RowHandler) Insert(c echo.Context) error {
	return rh.proxy(c, "insert")
}

// Update proxies a row UPDATE to PostgREST and fires webhooks
//
// @Summary Update rows
// @Description Update rows in a table via the Fluxend proxy
// @Tags Rows
//
// @Accept json
// @Produce json
//
// @Param Authorization header string true "Bearer Token"
// @Param X-Project header string true "Project UUID"
// @Param fullTableName path string true "Full table name (e.g. public.orders)"
//
// @Success 200 "Rows updated"
// @Failure 400 {object} response.BadRequestErrorResponse "Bad request response"
// @Failure 401 {object} response.UnauthorizedErrorResponse "Unauthorized response"
// @Failure 500 {object} response.InternalServerErrorResponse "Internal server error response"
//
// @Router /tables/{fullTableName}/rows [patch]
func (rh *RowHandler) Update(c echo.Context) error {
	return rh.proxy(c, "update")
}

// Delete proxies a row DELETE to PostgREST and fires webhooks
//
// @Summary Delete rows
// @Description Delete rows from a table via the Fluxend proxy
// @Tags Rows
//
// @Accept json
// @Produce json
//
// @Param Authorization header string true "Bearer Token"
// @Param X-Project header string true "Project UUID"
// @Param fullTableName path string true "Full table name (e.g. public.orders)"
//
// @Success 200 "Rows deleted"
// @Failure 400 {object} response.BadRequestErrorResponse "Bad request response"
// @Failure 401 {object} response.UnauthorizedErrorResponse "Unauthorized response"
// @Failure 500 {object} response.InternalServerErrorResponse "Internal server error response"
//
// @Router /tables/{fullTableName}/rows [delete]
func (rh *RowHandler) Delete(c echo.Context) error {
	return rh.proxy(c, "delete")
}

func (rh *RowHandler) proxy(c echo.Context, event string) error {
	var baseRequest dto.DefaultRequestWithProjectHeader
	if err := baseRequest.WithProjectHeader(c); err != nil {
		return response.BadRequestResponse(c, err.Error())
	}

	dbName, err := rh.projectRepo.GetDatabaseNameByUUID(baseRequest.ProjectUUID)
	if err != nil {
		return response.ErrorResponse(c, err)
	}

	tableName := extractTableName(c.Param("fullTableName"))
	postgrestURL := fmt.Sprintf("http://postgrest_%s:3000/%s", dbName, tableName)

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return response.BadRequestResponse(c, "Failed to read request body")
	}

	method := c.Request().Method
	req, err := http.NewRequest(method, postgrestURL, strings.NewReader(string(body)))
	if err != nil {
		return response.BadRequestResponse(c, "Failed to build upstream request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	if auth := c.Request().Header.Get("Authorization"); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	// Forward query params (PostgREST filter syntax: id=eq.1, id=in.(1,2))
	req.URL.RawQuery = c.Request().URL.RawQuery

	resp, err := rh.httpClient.Do(req)
	if err != nil {
		return response.BadRequestResponse(c, "Failed to reach upstream database")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return response.BadRequestResponse(c, "Failed to read upstream response")
	}

	// Fire webhooks with the returned row data (best-effort, non-blocking)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && len(respBody) > 0 {
		projectUUID := baseRequest.ProjectUUID
		go rh.fireWebhooks(projectUUID, tableName, event, respBody)
	}

	// Mirror PostgREST content-type and status back to the client
	c.Response().Header().Set("Content-Type", "application/json")
	return c.Blob(resp.StatusCode, "application/json", respBody)
}

func (rh *RowHandler) fireWebhooks(projectUUID uuid.UUID, tableName, event string, body []byte) {
	var records interface{}
	if err := json.Unmarshal(body, &records); err != nil {
		return
	}

	switch v := records.(type) {
	case []interface{}:
		for _, record := range v {
			rh.webhookService.Fire(projectUUID, tableName, event, record)
		}
	default:
		rh.webhookService.Fire(projectUUID, tableName, event, records)
	}
}
