package logging

import (
	"fluxend/internal/domain/logging"
	"github.com/google/uuid"
)

func ToLogListInput(request *ListRequest, projectUUID uuid.NullUUID) *logging.ListInput {
	return &logging.ListInput{
		ProjectUuid: projectUUID,
		UserUuid:    request.UserUuid,
		Status:      request.Status,
		Method:      request.Method,
		Endpoint:    request.Endpoint,
		IPAddress:   request.IPAddress,
		StartTime:   request.StartTime,
		EndTime:     request.EndTime,
	}
}
