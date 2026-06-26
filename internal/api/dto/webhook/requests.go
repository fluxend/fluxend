package webhook

import (
	"fluxend/internal/api/dto"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/labstack/echo/v4"
)

var allowedEvents = []interface{}{"insert", "update", "delete"}

type CreateRequest struct {
	dto.DefaultRequestWithProjectHeader
	URL      string   `json:"url"`
	Events   []string `json:"events"`
	IsActive bool     `json:"is_active"`
}

func (r *CreateRequest) BindAndValidate(c echo.Context) []string {
	if err := c.Bind(r); err != nil {
		return []string{"Invalid request payload"}
	}

	if err := r.WithProjectHeader(c); err != nil {
		return []string{err.Error()}
	}

	eventInterfaces := make([]interface{}, len(r.Events))
	for i, e := range r.Events {
		eventInterfaces[i] = e
	}

	err := validation.ValidateStruct(r,
		validation.Field(&r.URL,
			validation.Required.Error("URL is required"),
			is.URL.Error("URL must be a valid URL"),
		),
		validation.Field(&r.Events,
			validation.Required.Error("Events are required"),
			validation.Length(1, 3).Error("At least one event is required"),
			validation.Each(validation.In(allowedEvents...).Error("Event must be one of: insert, update, delete")),
		),
	)

	return r.ExtractValidationErrors(err)
}
