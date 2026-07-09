package logging

import (
	"fluxend/internal/api/dto"
	"github.com/google/uuid"
	"github.com/guregu/null/v6"
	"github.com/labstack/echo/v4"
	"strconv"
	"time"
)

type ListRequest struct {
	dto.BaseRequest
	UserUuid  uuid.NullUUID `query:"userUuid"`
	Status    null.String   `query:"status"`
	Method    null.String   `query:"method"`
	Endpoint  null.String   `query:"endpoint"`
	IPAddress null.String   `query:"ipAddress"`
	StartTime time.Time     // Will be populated from timestamp parsing
	EndTime   time.Time     // Will be populated from timestamp parsing

	Limit int    `query:"limit"`
	Page  int    `query:"page"`
	Sort  string `query:"sort"`
	Order string `query:"order"`
}

func (r *ListRequest) BindAndValidate(c echo.Context) []string {
	if err := c.Bind(r); err != nil {
		return []string{"Invalid request payload"}
	}

	if startTimeStr := c.QueryParam("startTime"); startTimeStr != "" {
		timestamp, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			return []string{"Invalid startTime format, expected Unix timestamp"}
		}

		r.StartTime = time.Unix(timestamp, 0)
	}

	if endTimeStr := c.QueryParam("endTime"); endTimeStr != "" {
		timestamp, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			return []string{"Invalid endTime format, expected Unix timestamp"}
		}

		r.EndTime = time.Unix(timestamp, 0)
	}

	if !r.StartTime.IsZero() && !r.EndTime.IsZero() {
		if r.EndTime.Before(r.StartTime) {
			return []string{"endTime must be after or equal to startTime"}
		}
	}

	return nil
}
