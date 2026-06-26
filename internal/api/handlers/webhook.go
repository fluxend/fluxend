package handlers

import (
	"fluxend/internal/api/dto"
	webhookDto "fluxend/internal/api/dto/webhook"
	"fluxend/internal/api/mapper"
	"fluxend/internal/api/response"
	"fluxend/internal/domain/webhook"
	"fluxend/pkg/auth"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	"strings"
)

type WebhookHandler struct {
	webhookService webhook.Service
}

func NewWebhookHandler(injector *do.Injector) (*WebhookHandler, error) {
	webhookService := do.MustInvoke[webhook.Service](injector)

	return &WebhookHandler{webhookService: webhookService}, nil
}

// List retrieves all webhooks for a table
//
// @Summary List webhooks
// @Description Retrieve all webhook configurations for a table
// @Tags Webhooks
//
// @Accept json
// @Produce json
//
// @Param Authorization header string true "Bearer Token"
// @Param X-Project header string true "Project UUID"
// @Param fullTableName path string true "Full table name (e.g. public.orders)"
//
// @Success 200 {array} response.Response{content=[]webhook.Response} "List of webhooks"
// @Failure 400 {object} response.BadRequestErrorResponse "Bad request response"
// @Failure 401 {object} response.UnauthorizedErrorResponse "Unauthorized response"
// @Failure 500 {object} response.InternalServerErrorResponse "Internal server error response"
//
// @Router /tables/{fullTableName}/webhooks [get]
func (wh *WebhookHandler) List(c echo.Context) error {
	var request dto.DefaultRequestWithProjectHeader
	if err := request.BindAndValidate(c); err != nil {
		return response.UnprocessableResponse(c, err)
	}

	authUser, _ := auth.NewAuth(c).User()

	tableName := extractTableName(c.Param("fullTableName"))

	configs, err := wh.webhookService.List(request.ProjectUUID, tableName, authUser)
	if err != nil {
		return response.ErrorResponse(c, err)
	}

	return response.SuccessResponse(c, mapper.ToWebhookResourceCollection(configs))
}

// Store creates a new webhook for a table
//
// @Summary Create webhook
// @Description Add a new webhook configuration for a table
// @Tags Webhooks
//
// @Accept json
// @Produce json
//
// @Param Authorization header string true "Bearer Token"
// @Param X-Project header string true "Project UUID"
// @Param fullTableName path string true "Full table name (e.g. public.orders)"
// @Param webhook body webhook.CreateRequest true "Webhook URL, events, and active flag"
//
// @Success 201 {object} response.Response{content=webhook.Response} "Webhook created"
// @Failure 422 {object} response.UnprocessableErrorResponse "Unprocessable input response"
// @Failure 400 {object} response.BadRequestErrorResponse "Bad request response"
// @Failure 401 {object} response.UnauthorizedErrorResponse "Unauthorized response"
// @Failure 500 {object} response.InternalServerErrorResponse "Internal server error response"
//
// @Router /tables/{fullTableName}/webhooks [post]
func (wh *WebhookHandler) Store(c echo.Context) error {
	var request webhookDto.CreateRequest
	if err := request.BindAndValidate(c); err != nil {
		return response.UnprocessableResponse(c, err)
	}

	authUser, _ := auth.NewAuth(c).User()

	tableName := extractTableName(c.Param("fullTableName"))

	input := webhook.CreateInput{
		ProjectUUID: request.ProjectUUID,
		TableName:   tableName,
		URL:         request.URL,
		Events:      request.Events,
		IsActive:    request.IsActive,
	}

	created, err := wh.webhookService.Create(input, authUser)
	if err != nil {
		return response.ErrorResponse(c, err)
	}

	return response.CreatedResponse(c, mapper.ToWebhookResponse(&created))
}

// Delete removes a webhook configuration
//
// @Summary Delete webhook
// @Description Remove a webhook configuration from a table
// @Tags Webhooks
//
// @Accept json
// @Produce json
//
// @Param Authorization header string true "Bearer Token"
// @Param X-Project header string true "Project UUID"
// @Param fullTableName path string true "Full table name (e.g. public.orders)"
// @Param webhookUUID path string true "Webhook UUID"
//
// @Success 204 "Webhook deleted"
// @Failure 400 {object} response.BadRequestErrorResponse "Bad request response"
// @Failure 401 {object} response.UnauthorizedErrorResponse "Unauthorized response"
// @Failure 500 {object} response.InternalServerErrorResponse "Internal server error response"
//
// @Router /tables/{fullTableName}/webhooks/{webhookUUID} [delete]
func (wh *WebhookHandler) Delete(c echo.Context) error {
	var request dto.DefaultRequestWithProjectHeader
	if err := request.BindAndValidate(c); err != nil {
		return response.UnprocessableResponse(c, err)
	}

	authUser, _ := auth.NewAuth(c).User()

	webhookUUID, err := uuid.Parse(c.Param("webhookUUID"))
	if err != nil {
		return response.BadRequestResponse(c, "Invalid webhook UUID")
	}

	if _, err := wh.webhookService.Delete(webhookUUID, authUser); err != nil {
		return response.ErrorResponse(c, err)
	}

	return response.DeletedResponse(c, nil)
}

func extractTableName(fullTableName string) string {
	if idx := strings.Index(fullTableName, "."); idx != -1 {
		return fullTableName[idx+1:]
	}

	return fullTableName
}
