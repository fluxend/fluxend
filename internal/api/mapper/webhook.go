package mapper

import (
	webhookDto "fluxend/internal/api/dto/webhook"
	webhookDomain "fluxend/internal/domain/webhook"
)

func ToWebhookResponse(config *webhookDomain.Config) webhookDto.Response {
	return webhookDto.Response{
		Uuid:      config.Uuid,
		TableName: config.TableName,
		URL:       config.URL,
		Events:    []string(config.Events),
		IsActive:  config.IsActive,
		CreatedAt: config.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: config.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func ToWebhookResourceCollection(configs []webhookDomain.Config) []webhookDto.Response {
	responses := make([]webhookDto.Response, len(configs))
	for i := range configs {
		responses[i] = ToWebhookResponse(&configs[i])
	}

	return responses
}
